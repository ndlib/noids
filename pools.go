package main

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/dbrower/noids/noid"
)

// PoolInfo contains the public info for a pool. We
// use separate structures since the private structure contains
// a mutex we do not wish to copy. The private structure is
// the canonical source.
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
	name     string
	store    PoolStore
}

type poolGroup struct {
	sync.RWMutex
	table map[string]*pool
	names []string
}

var (
	DefaultStore PoolStore = NullStore{}

	NameExists = errors.New("Name already exists")
	NoSuchPool = errors.New("Pool could not be found")
	PoolEmpty  = errors.New("Pool is empty")
	PoolClosed = errors.New("Pool is closed")
	InvalidId  = errors.New("Id is invalid for this counter")
)

func NewPoolGroup() *poolGroup {
	return &poolGroup{table: make(map[string]*pool)}
}

// Create a new pool having the given name and template.
func (pg *poolGroup) AddPool(name, template string) (PoolInfo, error) {
	pi := PoolInfo{
		Name:     name,
		Template: template,
		LastMint: time.Now(),
	}
	err := pg.loadFromInfo(&pi)
	if err == nil {
		err = DefaultStore.SavePool(name, pi)
	}
	return pi, err
}

// AllPools returns a list of names for every pool in the system.
func (pg *poolGroup) AllPools() []string {
	pools.RLock()
	defer pools.RUnlock()

	result := make([]string, len(pg.names))
	copy(result, pg.names)

	return result
}

func (pg *poolGroup) lookupPool(name string) (*pool, error) {
	var err error = nil

	pg.RLock()
	p := pg.table[name]
	pg.RUnlock()

	if p == nil {
		err = NoSuchPool
	}
	return p, err
}

// Get information on the pool named.
// Returns an error if the given pool could not be found.
func (pg *poolGroup) GetPool(name string) (PoolInfo, error) {
	result := PoolInfo{Name: name}

	p, err := pg.lookupPool(name)
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
	pi.Name = p.name
	pi.Template = p.noid.String()
	pi.Used, pi.Max = p.noid.Count()
	pi.Closed = p.closed
	pi.LastMint = p.lastMint
}

// Mark the named pool as either open (false) or closed (false).
// If the pool is empty, a PoolEmpty error is returned and the pool
// remains closed.
func (pg *poolGroup) SetPoolState(name string, makeClosed bool) (PoolInfo, error) {
	pi := PoolInfo{Name: name}
	p, err := pg.lookupPool(name)
	if err != nil {
		return pi, err
	}

	p.Lock()
	defer p.Unlock()

	var needSave = false
	if !makeClosed && p.empty {
		copyPoolInfo(&pi, p)
		return pi, PoolEmpty
	}
	if p.closed != makeClosed {
		p.closed = makeClosed
		needSave = true
	}
	copyPoolInfo(&pi, p)
	if needSave {
		p.store.SavePool(p.name, pi)
	}
	return pi, nil
}

// Mint the given number of ids from the pool named.
// Less ids than requested may be returned if the pool
// is empty or closed.
func (pg *poolGroup) PoolMint(name string, count int) ([]string, error) {
	var result []string = make([]string, 0, count)
	p, err := pg.lookupPool(name)
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
		pi := PoolInfo{Name: name}
		copyPoolInfo(&pi, p)
		err = p.store.SavePool(p.name, pi)
	}

	return result, err
}

// Ensure that pool named will never mint the given id.
// Returns the updated pool info
func (pg *poolGroup) PoolAdvancePast(name, id string) (PoolInfo, error) {
	pi := PoolInfo{Name: name}
	p, err := pg.lookupPool(name)
	if err != nil {
		return pi, err
	}

	p.Lock()
	defer p.Unlock()

	var needSave = false
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
		needSave = true
	}

	copyPoolInfo(&pi, p)
	if needSave {
		err = p.store.SavePool(p.name, pi)
	}
	return pi, err
}

// creates a new pool entry using the information in `pi`.
// updates `pi` with the result (e.g. fix the Used and Max fields)
func (pg *poolGroup) loadFromInfo(pi *PoolInfo) error {
	pg.Lock()
	defer pg.Unlock()
	_, ok := pg.table[pi.Name]
	if ok {
		return NameExists
	}
	noid, err := noid.NewNoid(pi.Template)
	if err != nil {
		return err
	}
	p := &pool{
		noid:     noid,
		name:     pi.Name,
		closed:   pi.Closed,
		lastMint: pi.LastMint,
		store:    DefaultStore,
	}
	// don't technically hold the lock for p, but it hasn't been inserted into pools, yet
	copyPoolInfo(pi, p)
	p.empty = pi.Used == pi.Max
	pg.table[pi.Name] = p
	pg.names = append(pg.names, pi.Name)
	return nil
}

func (pg *poolGroup) LoadPoolsFromStore(ps PoolStore) error {
	pis, err := ps.LoadAllPools()
	if err != nil {
		log.Fatal(err)
		return err
	}
	return pg.LoadPools(pis)
}

func (pg *poolGroup) LoadPools(pis []PoolInfo) error {
	for i := range pis {
		log.Println("Loading", pis[i].Name)
		err := pg.loadFromInfo(&pis[i])
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}
