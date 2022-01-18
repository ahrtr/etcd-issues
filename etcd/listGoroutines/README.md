Goroutine leak detection
======
Register a function to be called when the test (or subtest) and all its subtests complete, 
```
    t.Cleanup(func() {
        afterTest(t)
    })
```

In the above function `afterTest`, call the function `allGoroutines` in main.go to list all the existing goroutines, and filter out some 
expected goroutines. If there are still some remaining items, then they are potential candidates for the leaked goroutines.

The example output of the program `main.go` is as below,
```
$ go run main.go 
name: main.allGoroutines, file: /Users/wachao/go/src/github.com/ahrtr/etcd-issues/etcd/listGoroutines/main.go, line: 14, ok: true

Goroutine count: 4

### 0:
goroutine 1 [running]:
main.allGoroutines(0x0)
	/Users/wachao/go/src/github.com/ahrtr/etcd-issues/etcd/listGoroutines/main.go:18 +0x1a5
main.main()
	/Users/wachao/go/src/github.com/ahrtr/etcd-issues/etcd/listGoroutines/main.go:59 +0x4e

### 1:
goroutine 18 [chan receive]:
main.routine1()
	/Users/wachao/go/src/github.com/ahrtr/etcd-issues/etcd/listGoroutines/main.go:40 +0x2f
created by main.Demo1
	/Users/wachao/go/src/github.com/ahrtr/etcd-issues/etcd/listGoroutines/main.go:51 +0x25

### 2:
goroutine 19 [chan receive]:
main.routine1()
	/Users/wachao/go/src/github.com/ahrtr/etcd-issues/etcd/listGoroutines/main.go:40 +0x2f
created by main.main
	/Users/wachao/go/src/github.com/ahrtr/etcd-issues/etcd/listGoroutines/main.go:56 +0x39

### 3:
goroutine 20 [chan receive]:
main.routine2()
	/Users/wachao/go/src/github.com/ahrtr/etcd-issues/etcd/listGoroutines/main.go:46 +0x2f
created by main.main
	/Users/wachao/go/src/github.com/ahrtr/etcd-issues/etcd/listGoroutines/main.go:57 +0x47
```
