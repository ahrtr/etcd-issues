package main

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

/*
When this program is blocked on the time.Sleep(10*time.Second), you can
still cat the "/tmp/db" file. But if you run this program again, then you
will get a panic.
*/

func main() {
	fmt.Println("Starting")
	file, err := os.OpenFile("/tmp/db", os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	fmt.Println("Attempting to obtain the exclusive lock")
	if err = syscall.Flock(int(file.Fd()), syscall.LOCK_NB|syscall.LOCK_EX); err != nil {
		// panic here if you run this program again when previous one is blocked on the sleep.
		panic(err)
	}
	fmt.Println("Locked successfully")

	time.Sleep(10 * time.Second)

	fmt.Println("Releasing the lock")
	if err = syscall.Flock(int(file.Fd()), syscall.LOCK_UN); err != nil {
		panic(err)
	}
	fmt.Println("Lock released")

	if err = file.Close(); err != nil {
		panic(err)
	}

	fmt.Println("Done")
}
