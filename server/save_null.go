package main

import "log"

type NullStore struct{}

func (ns NullStore) SavePool(name string, pi PoolInfo) error {
	log.Println("Save (null)", name)
	return nil
}

func (ns NullStore) LoadAllPools() ([]PoolInfo, error) {
	var pi []PoolInfo
	return pi, nil
}
