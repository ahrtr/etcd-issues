package main

import (
	"fmt"
	bolt "go.etcd.io/bbolt"
	"os"
)

const TestFreelistType = "TEST_FREELIST_TYPE"

// DB is a test wrapper for bolt.DB.
type DB struct {
	*bolt.DB
	f string
	o *bolt.Options
}

// MustOpenDB returns a new, open DB at a temporary location.
func MustOpenDB() *DB {
	return MustOpenWithOption(nil)
}

// MustClose closes the database and deletes the underlying file. Panic on error.
func (db *DB) MustClose() {
	if err := db.Close(); err != nil {
		panic(err)
	}
}

// MustOpenWithOption returns a new, open DB at a temporary location with given options.
func MustOpenWithOption(o *bolt.Options) *DB {
	f := tempfile()
	if o == nil {
		o = bolt.DefaultOptions
	}

	freelistType := bolt.FreelistArrayType
	if env := os.Getenv(TestFreelistType); env == string(bolt.FreelistMapType) {
		freelistType = bolt.FreelistMapType
	}
	o.FreelistType = freelistType

	db, err := bolt.Open(f, 0666, o)
	if err != nil {
		panic(err)
	}
	return &DB{
		DB: db,
		f:  f,
		o:  o,
	}
}

// tempfile returns a temporary file path.
func tempfile() string {
	fileName := "db"
	if _, err := os.Stat(fileName); err == nil {
		if err := os.Remove(fileName); err != nil {
			panic(err)
		}
	}
	return fileName
}

func main() {
	fmt.Println("Starting")
	generateDBFile()
	fmt.Println("Done")
}

/*
Structure:
[b] root
	[b] ben_a
		[k/v] 3 entries
		[b] ben_a_1
			[k/v] 10 entries
		[b} ben_a_2
			[k/v] 10 entries
	[b] ben_b
		[k/v] 4 entries
		[b] ben_b_1
			[k/v] 10 entries
		[b] ben_b_2
			[k/v] 10 entries
 */
func generateDBFile() {
	db := MustOpenDB()
	defer db.MustClose()

	if err := db.Update(func(tx *bolt.Tx) error {
		// ================= ben_a
		b1, err := tx.CreateBucket([]byte("ben_a"))
		if err != nil {
			panic(err)
		}
		writeDataToBucket(b1, 3)

		b1_1, err := b1.CreateBucket([]byte("ben_a_1"))
		if err != nil {
			panic(err)
		}
		writeDataToBucket(b1_1, 10)

		b1_2, err := b1.CreateBucket([]byte("ben_a_2"))
		if err != nil {
			panic(err)
		}
		writeDataToBucket(b1_2, 10)

		// ================= ben_a
		b2, err := tx.CreateBucket([]byte("ben_b"))
		if err != nil {
			panic(err)
		}
		writeDataToBucket(b2, 4)

		b2_1, err := b2.CreateBucket([]byte("ben_b_1"))
		if err != nil {
			panic(err)
		}
		writeDataToBucket(b2_1, 10)

		b2_2, err := b2.CreateBucket([]byte("ben_b_2"))
		if err != nil {
			panic(err)
		}
		writeDataToBucket(b2_2, 10)

		return nil
	}); err != nil {
		panic(err)
	}
}

func writeDataToBucket(b *bolt.Bucket, count int) {
	for i:=0; i<count; i++ {
		key, value := fmt.Sprintf("key-%d", i), fmt.Sprintf("value-%d", i)
		if err := b.Put([]byte(key), []byte(value)); err != nil {
			panic(err)
		}
	}
}
