package main

import (
	"database/sql"
	"testing"
	"time"

	_ "code.google.com/p/go-sqlite/go1/sqlite3"
)

func TestDbSavePool(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skip(err)
		return
	}
	defer db.Close()

	ps := NewDbFileStore(db)
	err = ps.SavePool("test", PoolInfo{
		Name:     "test",
		Template: ".zd+0",
		Closed:   false,
		LastMint: time.Now(),
	})

	pis, err := ps.LoadAllPools()
	if err != nil {
		t.Logf(err.Error())
		return
	}
	if len(pis) == 1 && pis[0].Name == "test" {
		// good
	} else {
		t.Logf("pool was not saved")
		t.Logf("Got %v", pis)
		return
	}
}
