package main

import (
	"fmt"
	"os"
	"unsafe"
)

/*
Issue: https://github.com/etcd-io/etcd/issues/13406
The reason of the issue is that the db file is corrupted. The solution is to fix the corrupted
db file, but please note that there are some data loss, and I am not responsible for the data loss!

Author: Benjamin Wang
Date: 2021-10-12
*/

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <filename>\n", os.Args[0])
		os.Exit(1)
	}
	f, err := os.OpenFile(os.Args[1], os.O_RDWR, 0644)
	if err != nil {
		panic(fmt.Sprintf("failed to open the file, error: %v", err))
	}

	if err := fixMetadata(f); err != nil {
		panic(err)
	}

	// The page 165 is a branchPage, and there are 82 entries.
	// The 54th entry has an invalid pgid, which is 1652, so the solution is to
	// drop all the entries starting from 54.
	if err := fixPageCount(f, 165, 54); err != nil {
		panic(err)
	}

	// The page 461 is a branchPage, and there are 5 entries.
	// The 3rd entry has an invalid pgid, which is 1113, so the solution is to
	// drop all the entries starting from 3.
	if err := fixPageCount(f, 461, 3); err != nil {
		panic(err)
	}

	// The page 388 is a branchPage, and there are 111 entries.
	// The 3rd entry has an invalid pgid, which is 1116, so the solution is to
	// drop all the entries starting from 3.
	if err := fixPageCount(f, 388, 3); err != nil {
		panic(err)
	}

	if err := fixMetaChecksum(f); err != nil {
		panic(err)
	}

	fmt.Println("Fixed the db file")
}

func fixMetadata(f *os.File) error {
	buf := make([]byte, 8)

	pgid := (*uint64)(unsafe.Pointer(&buf[0]))
	// The db file size is 2527232, and the pageSize is 4096, so there are 617 pages in total.
	// So the max page id is 616.
	*pgid = 616

	_, err := f.WriteAt(buf, 56)
	if err != nil {
		return fmt.Errorf("failed to write data to meta0, error: %v", err)
	}

	_, err = f.WriteAt(buf, 4096+56)
	if err != nil {
		return fmt.Errorf("failed to write data to meta1, error: %v", err)
	}

	return nil
}

func fixPageCount(f *os.File, pgid uint64, count uint16) error {
	pageBaseAddr := 4096 * pgid

	buf := make([]byte, 2)
	p := (*uint16)(unsafe.Pointer(&buf[0]))
	*p = count

	_, err := f.WriteAt(buf, int64(pageBaseAddr+10))
	if err != nil {
		return fmt.Errorf("failed to write data to count, error: %v", err)
	}

	return nil
}

func fixMetaChecksum(f *os.File) error {
	// The values are calculated using bbolt
	var chk0, chk1 uint64 = 10020837603446655706, 15451967749325709303
	var offset int64 = 72
	buf0, buf1 := make([]byte, 8), make([]byte, 8)

	p0 := (*uint64)(unsafe.Pointer(&buf0[0]))
	p1 := (*uint64)(unsafe.Pointer(&buf1[0]))

	*p0 = chk0
	*p1 = chk1

	if _, err := f.WriteAt(buf0, offset); err != nil {
		return fmt.Errorf("failed to update checksum for meta0: error: %v", err)
	}

	if _, err := f.WriteAt(buf1, 4096+offset); err != nil {
		return fmt.Errorf("failed to update checksum for meta1: error: %v", err)
	}

	return nil
}
