package server

import (
	"log"
	"time"
)

var (
	FSave      = 1
	FForceSave = 2
)

type PoolSaver interface {
	SavePool(name string, info PoolInfo) error
	LoadAllPools() ([]PoolInfo, error)
}

func StartSaver(ps PoolSaver) chan<- int {
	c := make(chan int, 2)
	si := saverInfo{
		control: c,
		ps:      ps,
		timeout: 5,
	}
	si.LoadPools()
	go si.saver()
	return c
}

type saverInfo struct {
	control <-chan int
	ps      PoolSaver
	timeout int // number of seconds to wait between polls
}

/* this currently polls...make saves be requested by the pool handler */

func (si saverInfo) saver() {
	log.Println("Starting saver")
	tick := time.NewTicker(time.Duration(si.timeout) * time.Second)
	defer tick.Stop()
	for {
		select {
		case command, ok := <-si.control:
			if !ok {
				return
			}
			var force bool
			if command == FForceSave {
				force = true
			}
			si.saveSweep(force)
		case <-tick.C:
			si.saveSweep(false)
		}
	}
}

// This may miss a new pool since we do not keep
// the overall pool lock while asking for each pool.
// But, in that case, we will get it during the next sweep
func (si saverInfo) saveSweep(force bool) {
	// log.Println("Save Sweep")
	poolList := AllPools()
	for _, pname := range poolList {
		p, err := lookupPool(pname)
		if err != nil {
			continue
		}
		p.Lock()
		if p.needSave || force {
			log.Printf("Pool %s needs save\n", pname)
			err := si.savePool(pname, p)
			if err != nil {
				log.Println(err)
			} else {
				p.needSave = false
			}
		}
		p.Unlock()
	}
}

// called while the lock on p is held
func (si saverInfo) savePool(name string, p *pool) error {
	pi := PoolInfo{Name: name}
	copyPoolInfo(&pi, p)
	return si.ps.SavePool(name, pi)
}

func (si saverInfo) LoadPools() {
	log.Println("Loading saved pools")
	pis, err := si.ps.LoadAllPools()

	if err != nil {
		log.Fatal(err)
	}

	for i := range pis {
		log.Println("Loading", pis[i].Name)
		err := loadFromInfo(&pis[i], false)
		if err != nil {
			log.Println(err)
		}
	}
}
