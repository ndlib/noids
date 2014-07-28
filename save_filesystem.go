package main

import (
	"encoding/json"
	"log"
	"os"
	"path"
	"regexp"
)

type dirstore struct {
	root string
}

// Create a PoolStore which will serialize noid pools as
// json files in a directory.
func NewJsonFileStore(dirname string) PoolStore {
	return &dirstore{root: dirname}
}

func (d *dirstore) SavePool(name string, pi PoolInfo) error {
	log.Println("Save (filesystem)", name)
	f, err := os.Create(sanitizeName(d.root, name))
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.Encode(pi) // TODO: err here?

	return nil
}

func (d *dirstore) LoadAllPools() ([]PoolInfo, error) {
	var pis []PoolInfo
	f, err := os.Open(d.root)
	if err != nil {
		return pis, err
	}
	for {
		names, err := f.Readdirnames(10)
		if err != nil {
			break
		}
		for _, s := range names {
			pi, err := d.loadpool(s)
			if err != nil {
				return pis, err
			}
			pis = append(pis, pi)
		}
	}
	return pis, nil
}

func (d *dirstore) loadpool(filename string) (PoolInfo, error) {
	var pi PoolInfo
	f, err := os.Open(sanitizeName(d.root, filename))
	if err != nil {
		return pi, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	err = dec.Decode(&pi)

	return pi, err
}

var (
	badchars = regexp.MustCompile(`\.\.|/`)
)

func sanitizeName(root, s string) string {
	s = badchars.ReplaceAllLiteralString(s, "_")
	return path.Join(root, s)
}
