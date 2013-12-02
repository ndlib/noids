package server

import (
	"errors"
	"fmt"
	"sync"

	"github.com/dbrower/noids/noid"
)

type PoolInfo struct {
	Name, Template string
	Used, Max      int
	Closed         bool
}

type pool struct {
	m      sync.Mutex
	noid   noid.Noid
	closed bool
	empty  bool
}

type poolNames struct {
	m     sync.Mutex
	table map[string]*pool
	names []string
}

var (
	pools poolNames = poolNames{table: make(map[string]*pool)}

	NameExists = errors.New("Name already exists")
	NoSuchPool = errors.New("Pool could not be found")
	PoolEmpty  = errors.New("Pool is empty")
)

func AddPool(name, template string) error {
	pools.m.Lock()
	defer pools.m.Unlock()
	fmt.Printf("%v\n", pools)
	_, ok := pools.table[name]
	if ok {
		return NameExists
	}
	noid, err := noid.NewNoid(template)
	if err != nil {
		return err
	}
	pools.table[name] = &pool{
		noid:   noid,
		closed: false,
	}
	pools.names = append(pools.names, name)
	return nil
}

func AllPools() []string {
	pools.m.Lock()
	defer pools.m.Unlock()

	result := make([]string, len(pools.names))
	copy(result, pools.names)

	return result
}

func lookupPool(name string) (*pool, error) {
	var err error = nil

	pools.m.Lock()
	p := pools.table[name]
	pools.m.Unlock()

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

	p.m.Lock()
	defer p.m.Unlock()

	result.Template = p.noid.String()
	result.Used, result.Max = p.noid.Count()
	result.Closed = p.closed
	return result, nil
}

func SetPoolState(name string, newClosed bool) error {
	p, err := lookupPool(name)
	if err != nil {
		return err
	}

	p.m.Lock()
	defer p.m.Unlock()

	if !newClosed && p.empty {
		return PoolEmpty
	}
	p.closed = newClosed
	return nil
}

func PoolMint(name string, count int) ([]string, error) {
	var result []string
	p, err := lookupPool(name)
	if err != nil {
		return result, err
	}

	p.m.Lock()
	defer p.m.Unlock()

	for count > 0 {
		id := p.noid.Mint()
		if id == "" {
			p.empty = true
			p.closed = true
			break
		}
		result = append(result, id)
		count--
	}

	return result, nil
}
