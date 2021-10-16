issue https://github.com/etcd-io/etcd/issues/13406
======
# Issue
etcd version: 3.5.

After restarting from a zfs snapshot rollback, the etcd failed to get started, and the stack trace is as below,
```
unexpected fault address 0x7f31a61db000                                                                              
fatal error: fault                                                                                                   
[signal SIGBUS: bus error code=0x2 addr=0x7f31a61db000 pc=0x9dd562]                                                                                                                                                                         
                                                          
goroutine 157 [running]:                                                                                             
runtime.throw(0x1202783, 0x5)                                                                                        
        /usr/local/go/src/runtime/panic.go:1117 +0x72 fp=0xc0000c1ea0 sp=0xc0000c1e70 pc=0x4385d2                                                                                                                                           
runtime.sigpanic()                                                                                                   
        /usr/local/go/src/runtime/signal_unix.go:731 +0x2c8 fp=0xc0000c1ed8 sp=0xc0000c1ea0 pc=0x44fe48                                                                                                                                     
go.etcd.io/bbolt.(*Tx).checkBucket.func1(0x7f31a61db000, 0x2)                                                        
        /go/pkg/mod/go.etcd.io/bbolt@v1.3.6/tx.go:465 +0x62 fp=0xc0000c1fa0 sp=0xc0000c1ed8 pc=0x9dd562                                                                                                                                     
go.etcd.io/bbolt.(*Tx).forEachPage(0xc0000b01c0, 0x674, 0x2, 0xc0000c20c8)                                                                                                                                                                  
        /go/pkg/mod/go.etcd.io/bbolt@v1.3.6/tx.go:625 +0x89 fp=0xc0000c1fe8 sp=0xc0000c1fa0 pc=0x9dbde9                                                                                                                                     
go.etcd.io/bbolt.(*Tx).forEachPage(0xc0000b01c0, 0xa5, 0x1, 0xc0000c20c8)                                                                                                                                                                   
        /go/pkg/mod/go.etcd.io/bbolt@v1.3.6/tx.go:631 +0xe5 fp=0xc0000c2030 sp=0xc0000c1fe8 pc=0x9dbe45                                                                                                                                     
go.etcd.io/bbolt.(*Tx).forEachPage(0xc0000b01c0, 0x1cd, 0x0, 0xc0000c20c8)                                                                                                                                                                  
        /go/pkg/mod/go.etcd.io/bbolt@v1.3.6/tx.go:631 +0xe5 fp=0xc0000c2078 sp=0xc0000c2030 pc=0x9dbe45                                                                                                                                     
go.etcd.io/bbolt.(*Tx).checkBucket(0xc0000b01c0, 0xc000663180, 0xc0000c2370, 0xc0000c2340, 0xc0000424e0)                                                                                                                                    
        /go/pkg/mod/go.etcd.io/bbolt@v1.3.6/tx.go:464 +0xd4 fp=0xc0000c2108 sp=0xc0000c2078 pc=0x9db514                                                                                                                                     
go.etcd.io/bbolt.(*Tx).checkBucket.func2(0x7f31a5be81b9, 0x3, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0)                                                                                                                                                 
        /go/pkg/mod/go.etcd.io/bbolt@v1.3.6/tx.go:489 +0xc5 fp=0xc0000c2160 sp=0xc0000c2108 pc=0x9ddaa5                                                                                                                                     
go.etcd.io/bbolt.(*Bucket).ForEach(0xc0000b01d8, 0xc0000c21f0, 0x0, 0xc0000c2220)                                                                                                                                                           
        /go/pkg/mod/go.etcd.io/bbolt@v1.3.6/bucket.go:390 +0x103 fp=0xc0000c21d0 sp=0xc0000c2160 pc=0x9c86e3                                                                                                                                
go.etcd.io/bbolt.(*Tx).checkBucket(0xc0000b01c0, 0xc0000b01d8, 0xc0000c2370, 0xc0000c2340, 0xc0000424e0)                                                                                                                                    
        /go/pkg/mod/go.etcd.io/bbolt@v1.3.6/tx.go:487 +0x146 fp=0xc0000c2260 sp=0xc0000c21d0 pc=0x9db586                                                                                                                                    
go.etcd.io/bbolt.(*DB).freepages(0xc0000f8480, 0x0, 0x0, 0x0)                                                        
        /go/pkg/mod/go.etcd.io/bbolt@v1.3.6/db.go:1059 +0x210 fp=0xc0000c2458 sp=0xc0000c2260 pc=0x9d01f0                                                                                                                                   
go.etcd.io/bbolt.(*DB).loadFreelist.func1()                                                                           
        /go/pkg/mod/go.etcd.io/bbolt@v1.3.6/db.go:320 +0x114 fp=0xc0000c2490 sp=0xc0000c2458 pc=0x9dcc74                                                                                                                                    
sync.(*Once).doSlow(0xc0000f85f0, 0xc0000c24f0)                                                                      
        /usr/local/go/src/sync/once.go:68 +0xec fp=0xc0000c24e0 sp=0xc0000c2490 pc=0x47d94c                                                                                                                                                 
sync.(*Once).Do(...)                                                                                                 
        /usr/local/go/src/sync/once.go:59                                                                            
go.etcd.io/bbolt.(*DB).loadFreelist(0xc0000f8480)                                                                    
        /go/pkg/mod/go.etcd.io/bbolt@v1.3.6/db.go:316 +0x6a fp=0xc0000c2510 sp=0xc0000c24e0 pc=0x9cce0a                                                                                                                                     
go.etcd.io/bbolt.Open(0xc0004b11a0, 0x13, 0x180, 0xc0000c37c0, 0x40dbbb, 0xc0004b11a0, 0x18)                                                                                                                                                
        /go/pkg/mod/go.etcd.io/bbolt@v1.3.6/db.go:293 +0x3af fp=0xc0000c35d0 sp=0xc0000c2510 pc=0x9cc8af                                                                                                                                    
go.etcd.io/etcd/server/v3/mvcc/backend.newBackend(0xc0004b11a0, 0x13, 0x5f5e100, 0x2710, 0x1204c4f, 0x7, 0x280000000, 0xc000592640, 0x0, 0x137fe00, ...)
        /etcd/server/mvcc/backend/backend.go:180 +0x145 fp=0xc0000c3828 sp=0xc0000c35d0 pc=0xb76b85                                                                                                                                         
go.etcd.io/etcd/server/v3/mvcc/backend.New(...)                                                                      
        /etcd/server/mvcc/backend/backend.go:156                                                                     
go.etcd.io/etcd/server/v3/etcdserver.newBackend(0x7fff6ea0b794, 0x3, 0x0, 0x0, 0x0, 0x0, 0xc0000ab7a0, 0x1, 0x1, 0xc0000ab4d0, ...)
        /etcd/server/etcdserver/backend.go:55 +0x1f8 fp=0xc0000c39c0 sp=0xc0000c3828 pc=0xcfe9b8                                                                                                                                            
go.etcd.io/etcd/server/v3/etcdserver.openBackend.func1(0xc000042480, 0xc000150600, 0x137fe00, 0xc0000abd40)                                                                                                                                 
        /etcd/server/etcdserver/backend.go:76 +0x98 fp=0xc0000c3fc0 sp=0xc0000c39c0 pc=0xd41d78                                                                                                                                             
runtime.goexit()                                                                                                     
        /usr/local/go/src/runtime/asm_amd64.s:1371 +0x1 fp=0xc0000c3fc8 sp=0xc0000c3fc0 pc=0x4722a1                                                                                                                                         
created by go.etcd.io/etcd/server/v3/etcdserver.openBackend                                                           
        /etcd/server/etcdserver/backend.go:75 +0x12b                   
```

The file 'db' is the corrupted file.

# Root cause & Solution
The reason of the issue is that the db file is corrupted. The file size is 2527232 bytes, and the pageSize in meta is 4096, so there are 617 (2527232/4096) pages in total. But the pgid value in meta page is 1706, which is out of range.
There are also some invalid entries in the B-tree internal nodes as well.

The solution is to fix the corrupted db file, but please note that there may be some data loss, and I am not responsible for the data loss!

Run the following command to fix the db file,
```go
go run main.go <path-to-the-db-file>
```