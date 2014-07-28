package main

import (
	"database/sql"
	"log"
	"time"
)

type dbStore struct {
	DB *sql.DB
}

const dbSchema = `CREATE TABLE IF NOT EXISTS noids (
name VARCHAR(255) PRIMARY KEY,
template VARCHAR(255),
closed BOOLEAN,
lastmint VARCHAR(64)
);`

// Create a PoolStore which will serialize noid pools as
// records in a SQL database
func NewDbFileStore(db *sql.DB) PoolStore {
	// create table if necessary
	_, err := db.Exec(dbSchema)
	if err != nil {
		log.Printf("NewDbFileStore: %s", err.Error())
		return nil
	}
	return &dbStore{DB: db}
}

func (d *dbStore) SavePool(name string, pi PoolInfo) error {
	log.Println("Save (db)", name)
	lastmintText, err := pi.LastMint.MarshalText()
	result, err := d.DB.Exec("UPDATE noids SET template = ?, closed = ?, lastmint = ? WHERE name = ?", pi.Template, pi.Closed, string(lastmintText), name)
	if err != nil {
		return err
	}
	nrows, err := result.RowsAffected()
	if err != nil {
		// driver does not support row count
		// see if the record is in the database in the first place
		// TODO(dbrower)
		return err
	}
	switch {
	case nrows == 0:
		log.Println("Creating new db record for", name)
		_, err = d.DB.Exec("INSERT INTO noids VALUES (?, ?, ?, ?)", name, pi.Template, pi.Closed, string(lastmintText))
	case nrows == 1:
	default:
		log.Printf("There is more than one row in the database for pool '%s'", name)
		// TODO(dbrower): make error constant for this
		err = nil
	}
	return err
}

func (d *dbStore) LoadAllPools() ([]PoolInfo, error) {
	var pis []PoolInfo

	rows, err := d.DB.Query("SELECT name, template, closed, lastmint FROM noids")
	if err != nil {
		return pis, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			name, template, lastmint sql.NullString
			closed                   sql.NullBool
			lm                       time.Time
		)
		err := rows.Scan(&name, &template, &closed, &lastmint)
		if err != nil {
			return pis, err
		}
		err = (&lm).UnmarshalText([]byte(lastmint.String))
		if err != nil {
			return pis, err
		}
		pi := PoolInfo{
			Name:     name.String,
			Template: template.String,
			Closed:   closed.Bool,
			LastMint: lm,
		}
		pis = append(pis, pi)
	}
	if err := rows.Err(); err != nil {
		return pis, err
	}
	return pis, nil
}
