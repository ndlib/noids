package server

import "log"

type NullSaver struct{}

func (ns NullSaver) SavePool(name string, pi PoolInfo) error {
	log.Println("Save (null)", name)
	return nil
}

func (ns NullSaver) LoadAllPools() ([]PoolInfo, error) {
	var pi []PoolInfo
	return pi, nil
}
