package server

import (
	"errors"
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
	m        sync.Mutex
	noid     noid.Noid
	closed   bool
	empty    bool
	lastMint time.Time
	needSave bool
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
	err := loadFromInfo(
		PoolInfo{
			Name:     name,
			Template: template,
		},
		true)
	return err
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

	copyPoolInfo(&result, p)

	return result, nil
}

// this expects to be called while the lock on p is held
func copyPoolInfo(pi *PoolInfo, p *pool) {
	pi.Template = p.noid.String()
	pi.Used, pi.Max = p.noid.Count()
	pi.Closed = p.closed
	pi.LastMint = p.lastMint
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
	if p.closed != newClosed {
		p.closed = newClosed
		p.needSave = true
	}
	return nil
}

func PoolMint(name string, count int) ([]string, error) {
	var result []string = make([]string, 0, count)
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

	if len(result) > 0 {
		p.lastMint = time.Now()
		p.needSave = true
	}

	return result, nil
}

func loadFromInfo(pi PoolInfo, needSave bool) error {
	pools.m.Lock()
	defer pools.m.Unlock()
	_, ok := pools.table[pi.Name]
	if ok {
		return NameExists
	}
	noid, err := noid.NewNoid(pi.Template)
	if err != nil {
		return err
	}
	pools.table[pi.Name] = &pool{
		noid:     noid,
		needSave: needSave,
		closed:   pi.Closed,
		empty:    pi.Used == pi.Max,
		lastMint: pi.LastMint,
	}
	pools.names = append(pools.names, pi.Name)
	return nil
}
