package server

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/dbrower/noids/noid"
)

type PoolInfo struct {
	Name, Template string
	Used, Max      int
	Closed         bool
	LastMint       time.Time
}

type pool struct {
	sync.Mutex
	noid     noid.Noid
	closed   bool
	empty    bool
	lastMint time.Time
	needSave bool
}

type poolNames struct {
	sync.RWMutex
	table map[string]*pool
	names []string
}

var (
	pools poolNames = poolNames{table: make(map[string]*pool)}

	NameExists = errors.New("Name already exists")
	NoSuchPool = errors.New("Pool could not be found")
	PoolEmpty  = errors.New("Pool is empty")
	PoolClosed = errors.New("Pool is closed")
	InvalidId  = errors.New("Id is invalid for this counter")
)

func AddPool(name, template string) (PoolInfo, error) {
	pi := PoolInfo{
		Name:     name,
		Template: template,
		LastMint: time.Now(),
	}
	err := loadFromInfo(&pi, true)
	return pi, err
}

func AllPools() []string {
	pools.RLock()
	defer pools.RUnlock()

	result := make([]string, len(pools.names))
	copy(result, pools.names)

	return result
}

func lookupPool(name string) (*pool, error) {
	var err error = nil

	pools.RLock()
	p := pools.table[name]
	pools.RUnlock()

	if p == nil {
		err = NoSuchPool
	}
	return p, err
}

func GetPool(name string) (PoolInfo, error) {
	result := PoolInfo{Name: name}

	p, err := lookupPool(name)
	if err != nil {
		return result, err
	}

	p.Lock()
	defer p.Unlock()

	copyPoolInfo(&result, p)

	return result, nil
}

// Copies the information in p into pi.
// expects the caller to be holding the lock on p
func copyPoolInfo(pi *PoolInfo, p *pool) {
	pi.Template = p.noid.String()
	pi.Used, pi.Max = p.noid.Count()
	pi.Closed = p.closed
	pi.LastMint = p.lastMint
}

func SetPoolState(name string, newClosed bool) (PoolInfo, error) {
	pi := PoolInfo{Name: name}
	p, err := lookupPool(name)
	if err != nil {
		return pi, err
	}

	p.Lock()
	defer p.Unlock()

	if !newClosed && p.empty {
		copyPoolInfo(&pi, p)
		return pi, PoolEmpty
	}
	if p.closed != newClosed {
		p.closed = newClosed
		p.needSave = true
	}
	copyPoolInfo(&pi, p)
	return pi, nil
}

func PoolMint(name string, count int) ([]string, error) {
	var result []string = make([]string, 0, count)
	p, err := lookupPool(name)
	if err != nil {
		return result, err
	}

	p.Lock()
	defer p.Unlock()

	if p.closed {
		return result, PoolClosed
	}

	for ; count > 0; count-- {
		id := p.noid.Mint()
		if id == "" {
			p.empty = true
			p.closed = true
			break
		}
		result = append(result, id)
	}

	if len(result) > 0 {
		p.lastMint = time.Now()
		p.needSave = true
	}

	return result, nil
}

func PoolAdvancePast(name, id string) (PoolInfo, error) {
	pi := PoolInfo{Name: name}
	p, err := lookupPool(name)
	if err != nil {
		return pi, err
	}

	p.Lock()
	defer p.Unlock()

	index := p.noid.Index(id)
	log.Printf("Index(%v) = %v\n", id, index)
	if index == -1 {
		copyPoolInfo(&pi, p)
		return pi, InvalidId
	}
	position, _ := p.noid.Count()
	if index >= position {
		p.noid.AdvanceTo(index + 1)
		p.lastMint = time.Now()
		p.needSave = true
	}

	copyPoolInfo(&pi, p)
	return pi, nil
}

// creates a new pool entry using the information in `pi`.
// updates `pi` with the result (e.g. fix the Used and Max fields)
func loadFromInfo(pi *PoolInfo, needSave bool) error {
	pools.Lock()
	defer pools.Unlock()
	_, ok := pools.table[pi.Name]
	if ok {
		return NameExists
	}
	noid, err := noid.NewNoid(pi.Template)
	if err != nil {
		return err
	}
	p := &pool{
		noid:     noid,
		needSave: needSave,
		closed:   pi.Closed,
		lastMint: pi.LastMint,
	}
	// don't technically hold the lock for p, but it hasn't been inserted into pools, yet
	copyPoolInfo(pi, p)
	p.empty = pi.Used == pi.Max
	pools.table[pi.Name] = p
	pools.names = append(pools.names, pi.Name)
	return nil
}
