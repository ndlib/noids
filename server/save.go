package server

import (
	"encoding/json"
	"log"
	"os"
	"path"
	"time"
)

var (
	FSave      = 1
	FForceSave = 2
)

func StartSaver(storageDir string) chan<- int {
	if storageDir == "" {
		return nil
	}
	c := make(chan int, 2)
	si := saverInfo{
		control:    c,
		storageDir: storageDir,
		timeout:    5,
	}
	go si.saver()
	return c
}

type saverInfo struct {
	control    <-chan int
	storageDir string
	timeout    int // number of seconds to wait between polls
}

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
		if err == nil {
			p.m.Lock()
			if p.needSave || force {
				log.Printf("Pool %s needs save\n", pname)
				err := si.savePool(pname, p)
				if err != nil {
					log.Println(err)
				} else {
					p.needSave = false
				}
			}
			p.m.Unlock()
		}
	}
}

// called while the lock on p is held
func (si saverInfo) savePool(name string, p *pool) error {
	// TODO: sanitize the name....
	f, err := os.Create(path.Join(si.storageDir, name))
	if err != nil {
		return err
	}
	defer f.Close()

	pi := PoolInfo{Name: name}
	copyPoolInfo(&pi, p)
	enc := json.NewEncoder(f)
	enc.Encode(pi)

	return nil
}

func LoadPools(storageDir string) {
	log.Println("Loading pools in", storageDir)
	f, err := os.Open(storageDir)
	if err != nil {
		log.Fatal(err)
	}
	for {
		names, err := f.Readdirnames(10)
		if err != nil {
			break
		}
		for _, s := range names {
			loadpool(path.Join(storageDir, s), s)
		}
	}
}

func loadpool(filename, poolname string) {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	var pi PoolInfo
	dec := json.NewDecoder(f)
	err = dec.Decode(&pi)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Loading", pi.Name)
	loadFromInfo(pi, false)
}
