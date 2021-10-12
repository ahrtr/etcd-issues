issue https://github.com/etcd-io/etcd/issues/13406
======
The reason of the issue is that the db file is corrupted. The solution is to fix the corrupted
db file, but please note that there may be some data loss, and I am not responsible for the data loss!

Run the following command to fix the db file,
```go
go run main.go <path-to-the-db-file>
```