package main

import (
	"encoding/binary"
	"errors"
	"flag"
	"fmt"

	bolt "go.etcd.io/bbolt"
	"go.etcd.io/etcd/server/v3/datadir"
	"go.etcd.io/etcd/server/v3/mvcc/backend"
	"go.etcd.io/etcd/server/v3/mvcc/buckets"
)

/*
Workaround for https://github.com/etcd-io/etcd/issues/14382

Steps:
1. Get the consistent index value from the error log
	It's 1200012 in example below.
    {"level":"warn","ts":"2022-07-23T18:21:36.260Z","caller":"snap/db.go:88","msg":"failed to find [SNAPSHOT-INDEX].snap.db","snapshot-index":1200012,"snapshot-file-path":"/bitnami/etcd/data/member/snap/0000000000124f8c.snap.db","error":"snap: snapshot file doesn't exist"}
2. Update the consistent index
    Run commands something like below to update the consistent index,
    go build
    ./etcd-db-editor -data-dir ~/tmp/etcd/infra1.etcd/  -consistent-index 1200012
3. Start the etcd cluster again
4. Perform the compaction and defragmentation operation per https://etcd.io/docs/v3.5/op-guide/maintenance/.
    Also run `etcdctl alarm disarm` afterwards.
*/

var (
	dataDir         string
	consistentIndex uint64
)

func init() {
	flag.StringVar(&dataDir, "data-dir", ".", "etcd data directory")
	flag.Uint64Var(&consistentIndex, "consistent-index", 0, "new consistent index to set")
}

func main() {
	flag.Parse()

	if err := validate(); err != nil {
		panic(err)
	}

	dbFile := datadir.ToBackendFileName(dataDir)

	db, err := openDB(dbFile)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	UpdateConsistentIndex(db, consistentIndex)
}

func validate() error {
	if dataDir == "" {
		return errors.New("please set the data directory")
	}

	if consistentIndex == 0 {
		return errors.New("please set the consistent index")
	}

	fmt.Printf("Data directory: %s\n", dataDir)
	fmt.Printf("Consistent Index: %d\n\n", consistentIndex)
	return nil
}

func openDB(dbFile string) (*bolt.DB, error) {
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open boltDB file, %w", err)
	}

	return db, nil
}

func UpdateConsistentIndex(db *bolt.DB, index uint64) error {
	bs1 := make([]byte, 8)
	binary.BigEndian.PutUint64(bs1, index)

	if err := db.Update(func(tx *bolt.Tx) error {
		return saveField(tx, buckets.Meta, buckets.MetaConsistentIndexKeyName, bs1)
	}); err != nil {
		return fmt.Errorf("failed to update consistent index, %w", err)
	}

	fmt.Printf("Update consistent index (%d) successfully\n", index)
	return nil
}

func saveField(tx *bolt.Tx, bucketType backend.Bucket, key []byte, value []byte) error {
	bucket := tx.Bucket(bucketType.Name())
	if bucket == nil {
		bucketName := string(bucketType.Name())
		return fmt.Errorf("bucket (%s) doesn't exist", bucketName)
	}

	if err := bucket.Put(key, value); err != nil {
		return fmt.Errorf("failed to save data, %w", err)
	}

	return nil
}
