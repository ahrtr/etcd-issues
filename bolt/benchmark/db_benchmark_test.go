package bbolt_test

import (
	"math/rand"
	"os"
	"runtime/pprof"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"
	"go.etcd.io/bbolt/internal/btesting"
)

func Test_Large_Write_1_DB(t *testing.T) {
	var (
		f           = "/Users/wachao/tmp/etcd/bbolt/db"
		cpuProfile1 = "/Users/wachao/tmp/etcd/bbolt/profile1"
		cpuProfile2 = "/Users/wachao/tmp/etcd/bbolt/profile2"
	)

	db := openDBAndCpuProfile(t, f, cpuProfile1)
	defer db.Close()

	pf2 := mustCreateFile(t, cpuProfile2)
	defer pf2.Close()

	// CPU profile for write & commit
	if err := pprof.StartCPUProfile(pf2); err != nil {
		t.Fatalf("could not start CPU profile1: %v", err)
	}
	defer pprof.StopCPUProfile()
	t.Log("Starting a writable transaction")
	tx, err := db.Begin(true)
	if err != nil {
		t.Fatalf("Failed to start writable transaction:  %v", err)
	}
	t.Log("Opening bucket")
	b := tx.Bucket([]byte("data1"))
	t.Log("Writing data")
	if err := b.Put([]byte("newk1"), []byte("newv1")); err != nil {
		t.Fatalf("Failed to put data: %v", err)
	}
	t.Log("Committing the transaction")
	t1 := time.Now()
	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit the transaction: %v", err)
	}
	t.Logf("Duration of committing the transaction: %s", time.Since(t1))
}

func Test_Large_Write_400_DB(t *testing.T) {
	var (
		f           = "/Users/wachao/tmp/etcd/bbolt/db"
		cpuProfile1 = "/Users/wachao/tmp/etcd/bbolt/profile1"
		cpuProfile2 = "/Users/wachao/tmp/etcd/bbolt/profile2"
	)

	db := openDBAndCpuProfile(t, f, cpuProfile1)
	defer db.Close()

	pf2 := mustCreateFile(t, cpuProfile2)
	defer pf2.Close()

	// CPU profile for write & commit
	if err := pprof.StartCPUProfile(pf2); err != nil {
		t.Fatalf("could not start CPU profile1: %v", err)
	}
	defer pprof.StopCPUProfile()
	t.Log("Starting a writable transaction")
	tx, err := db.Begin(true)
	if err != nil {
		t.Fatalf("Failed to start writable transaction:  %v", err)
	}

	t1 := time.Now()
	for _, bn := range []string{"data1", "data2", "data3", "data4"} {
		t.Logf("Opening bucket: %s", bn)
		b := tx.Bucket([]byte(bn))
		t.Log("Writing data")
		for i := 0; i < 100; i++ {
			k := rand.Int31n(10000000)
			if err := b.Put(u64tokey(uint64(k)), make([]byte, 950+i)); err != nil {
				t.Fatal(err)
			}
		}
	}
	t.Logf("Duration of writing the data: %s", time.Since(t1))

	t.Log("Committing the transaction")
	t1 = time.Now()
	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit the transaction: %v", err)
	}
	t.Logf("Duration of committing the transaction: %s", time.Since(t1))
}

func Test_Large_Delete_Bucket_DB(t *testing.T) {
	var (
		f           = "/Users/wachao/tmp/etcd/bbolt/db"
		cpuProfile1 = "/Users/wachao/tmp/etcd/bbolt/profile1"
		cpuProfile2 = "/Users/wachao/tmp/etcd/bbolt/profile2"
	)

	pf2 := mustCreateFile(t, cpuProfile2)
	defer pf2.Close()

	db := openDBAndCpuProfile(t, f, cpuProfile1)
	defer db.Close()

	// CPU profile for write & commit
	if err := pprof.StartCPUProfile(pf2); err != nil {
		t.Fatalf("could not start CPU profile1: %v", err)
	}
	defer pprof.StopCPUProfile()
	t.Log("Starting a writable transaction")
	tx, err := db.Begin(true)
	if err != nil {
		t.Fatalf("Failed to start writable transaction:  %v", err)
	}
	t.Log("Deleting bucket")
	t1 := time.Now()
	if err := tx.DeleteBucket([]byte("data1")); err != nil {
		t.Fatalf("Failed to delete bucket: %v", err)
	}
	t.Logf("Duration of deleting the bucket: %s", time.Since(t1))

	t.Log("Committing the transaction")
	t1 = time.Now()
	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit the transaction: %v", err)
	}
	t.Logf("Duration of committing the transaction: %s", time.Since(t1))
}

func openDBAndCpuProfile(t *testing.T, dbFilename, profileFilename string) *bolt.DB {
	pf := mustCreateFile(t, profileFilename)
	defer pf.Close()

	// CPU profile for op.Open
	if err := pprof.StartCPUProfile(pf); err != nil {
		t.Fatalf("could not start CPU profile: %v", err)
	}
	defer pprof.StopCPUProfile()

	t.Log("Opening the db file")
	t1 := time.Now()
	db, err := bolt.Open(dbFilename, 0666, &bolt.Options{
		NoFreelistSync: true,
		FreelistType:   bolt.FreelistMapType,
	})
	if err != nil {
		t.Fatalf("Failed to open db: %v", err)
	}
	t.Logf("Duration to open the db file: %s", time.Since(t1))

	return db
}

// pageSize: 256K
// k/v pairs/bucket: 10M
func Test_Generate_Large_256K_10M_DB(t *testing.T) {
	generateLargeDb(t, 256*1024, 10000000)
}

func generateLargeDb(t *testing.T, pgSize, count int) {
	// Open a data file.
	f := "/Users/wachao/tmp/etcd/bbolt/db"
	var (
		/*
			pageSize: 4096 with freelist
			count		db size		time
			10000		100M		2.855s
			100000		850M		7.228s
			1000000		8.3G		69.228s
			10000000	83.26G		4474.413s
		*/
		/*
			pageSize: 4096
			count		db size		time
			10000		100M		2.222s
			100000		870M		4.943s
			1000000		8.34G		50.817s
			10000000	83.26G		5048.921s
		*/
		/*
			pageSize: 8192
			count		db size		time
			10000		126.9M		0.684s
			100000		1.12G		4.837s
			1000000		11.03G		45.608s
			10000000	110.11G		4648.309s
		*/
		/*
			pageSize: 16384
			count		db size		time
			10000		110.8M		2.139s
			100000		956.8M		4.128s
			1000000		9.42G		41.477s
			10000000	94.01G		3220.117s
		*/
		/*
			pageSize: 32768
			count		db size		time
			10000		104.4M		0.637s
			100000		892.5M		5.929s
			1000000		8.77G		67.940s
			10000000	87.57G		3017.547s
		*/
		/*
			pageSize: 65536
			count		db size		time
			10000		101.8M		1.954s
			100000		863.6M		3.419s
			1000000		8.48G		28.042s
			10000000	84.66G		2562.941s
		*/
		/*
			pageSize: 131072
			count		db size		time
			10000		100.8M		2.592s
			100000		850M		4.297s
			1000000		8.34G		32.437s
			10000000	83.28G		2438.051s
		*/
		/*
			pageSize: 262144
			count		db size		time
			10000		102M		1.875s
			100000		850.7M		4.114s
			1000000		8.34G		30.368s
			10000000	83.26G		2231.152
		*/
		/*
			pageSize: 524288
			count		db size		time
			10000		103.3M		2.592s
			100000		845.7M		3.953s
			1000000		8.28G		30.640s
			10000000	82.59G		2421.050s
		*/
		buckets = []string{"data1", "data2", "data3", "data4"}
	)

	db := btesting.MustOpenDBWithOption(t, f, &bolt.Options{PageSize: pgSize, NoFreelistSync: true})

	for _, bucket := range buckets {
		if err := db.Update(func(tx *bolt.Tx) error {
			t.Logf("Creating bucket %q", string(bucket))
			b, _ := tx.CreateBucketIfNotExists([]byte(bucket))
			t.Logf("Populate bucket %q with %d records", string(bucket), count)
			for i := 0; i < count; i++ {
				// key size: 16 bytes
				// value size: 1000 bytes
				if err := b.Put(u64tokey(uint64(i)), make([]byte, 1000)); err != nil {
					t.Fatal(err)
				}
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}

	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	t.Log("Done!")
}

func u64tokey(v uint64) []byte {
	b1 := u64tob(v)
	b2 := u64tob(v)

	b1 = append(b1, b2...)
	return b1
}

func mustCreateFile(t *testing.T, filename string) *os.File {
	f, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Failed to create file %q, error: %v", filename, err)
	}
	return f
}
