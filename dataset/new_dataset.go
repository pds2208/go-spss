package dataset

import (
	"log"
	"sync"
	"upper.io/db.v3"
	"upper.io/db.v3/lib/sqlbuilder"
	"upper.io/db.v3/sqlite"
)

type DatasetNew struct {
	sess sqlbuilder.Database
	mux  sync.Mutex
}

type SpssFile struct {
	Shiftno float64 `spss:"Shiftno"`
	Serial  float64 `spss:"Serial"`
	Version string  `spss:"Version"`
}

type Addresses struct {
	Row      int64   `db:"Row"`
	Name     string  `db:"Name"`
	Address  string  `db:"Address"`
	PostCode string  `db:"PostCode"`
	HowMany  float32 `db:"HowMany"`
}

func NewAddresses(name string) (db.Collection, error) {
	globalLock.Lock()
	defer globalLock.Unlock()

	var settings = sqlite.ConnectionURL{
		Database: "LFS.db", // file::memory:?cache=shared
	}

	sess, err := sqlite.Open(settings)

	if err != nil {
		panic(err)
	}

	_, err = sess.Query("drop table if exists %s", name)

	addresses := sess.Collection("addresses")
	_ = addresses.Truncate()

	_, err = sess.Query("create table %s (Row INTEGER PRIMARY KEY)", name)
	if err != nil {
		log.Printf("Cannot create table: %s: %q\n", name, err)
		return nil, err
	}

	return sess.Collection("addresses"), nil
}
