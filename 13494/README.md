issue https://github.com/etcd-io/etcd/issues/13494
======
## Issue
etcd version: 3.5.0. Actually release-3.5 and main (3.6) branches have this issue as well.

The etcd server crashed on startup, and the error message & stack is,

```
{"level":"info","ts":"2021-11-11T05:00:51.444Z","caller":"membership/cluster.go:285","msg":"set cluster version from store","cluster-version":"3.5"}
{"level":"warn","ts":"2021-11-11T05:00:51.445Z","caller":"auth/store.go:1220","msg":"simple token is not cryptographically signed"}
{"level":"info","ts":"2021-11-11T05:00:51.445Z","caller":"mvcc/kvstore.go:345","msg":"restored last compact revision","meta-bucket-name":"meta","meta-bucket-name-key":"finishedCompactRev","restored-compact-revision":1313948875}
{"level":"info","ts":"2021-11-11T05:00:52.818Z","caller":"mvcc/kvstore.go:415","msg":"kvstore restored","current-rev":1314002198}
panic: assertion failed: tx closed

goroutine 1 [running]:
go.etcd.io/bbolt._assert(...)
/home/remote/sbatsche/.gvm/pkgsets/go1.16.3/global/pkg/mod/go.etcd.io/bbolt@v1.3.6/db.go:1230
go.etcd.io/bbolt.(*Cursor).seek(0xc0005a27e0, 0x19ff0c8, 0x4, 0x4, 0x7fbec6ccffff, 0x400, 0x7fbec6b5a200, 0x20300200000000, 0x7fbec6ccffff, 0xc0005a2820, ...)
/home/remote/sbatsche/.gvm/pkgsets/go1.16.3/global/pkg/mod/go.etcd.io/bbolt@v1.3.6/cursor.go:155 +0x185
go.etcd.io/bbolt.(*Bucket).Bucket(0xc000448398, 0x19ff0c8, 0x4, 0x4, 0x7fbec6b5a100)
/home/remote/sbatsche/.gvm/pkgsets/go1.16.3/global/pkg/mod/go.etcd.io/bbolt@v1.3.6/bucket.go:105 +0xda
go.etcd.io/bbolt.(*Tx).Bucket(...)
/home/remote/sbatsche/.gvm/pkgsets/go1.16.3/global/pkg/mod/go.etcd.io/bbolt@v1.3.6/tx.go:101
go.etcd.io/etcd/server/v3/mvcc/backend.(*batchTx).UnsafeRange(0xc0002b2e10, 0x139a658, 0x1a62a40, 0x19ffd90, 0x10, 0x10, 0x0, 0x0, 0x0, 0x0, ...)
/tmp/etcd-release-3.5.0/etcd/release/etcd/server/mvcc/backend/batch_tx.go:149 +0x67
go.etcd.io/etcd/server/v3/etcdserver/cindex.unsafeReadConsistentIndex(0x7fbec70e0e80, 0xc0002b2e10, 0x40b945, 0x10e8680)
/tmp/etcd-release-3.5.0/etcd/release/etcd/server/etcdserver/cindex/cindex.go:136 +0x93
go.etcd.io/etcd/server/v3/etcdserver/cindex.ReadConsistentIndex(0x7fbec70e0e80, 0xc0002b2e10, 0x0, 0x0)
/tmp/etcd-release-3.5.0/etcd/release/etcd/server/etcdserver/cindex/cindex.go:154 +0x7f
go.etcd.io/etcd/server/v3/etcdserver/cindex.(*consistentIndex).ConsistentIndex(0xc0004b36e0, 0x0)
/tmp/etcd-release-3.5.0/etcd/release/etcd/server/etcdserver/cindex/cindex.go:77 +0xbd
go.etcd.io/etcd/server/v3/etcdserver.NewServer(0x7fff6ca691e8, 0x2a, 0x0, 0x0, 0x0, 0x0, 0xc0004e2000, 0x1, 0x1, 0xc0004e2240, ...)
/tmp/etcd-release-3.5.0/etcd/release/etcd/server/etcdserver/server.go:613 +0x2c71
```

The complete log is in [etcd.log](etcd.log).

## Analysis
When there is a local *.snap.db file, then etcd may recover v3 backend from it instead of the db file. In this case, 
etcd closes the old backend, and the *.snap.db is renamed db, please see [backend.go#L107](https://github.com/etcd-io/etcd/blob/e2273f94c4e1cd3c7add401009ac58399864783f/server/storage/backend.go#L107) 
and [backend.go#L64](https://github.com/etcd-io/etcd/blob/e2273f94c4e1cd3c7add401009ac58399864783f/server/storage/backend.go#L64).
But the [consistentIndex](https://github.com/etcd-io/etcd/blob/7572a61a39d4eaad596ab8d9364f7df9a84ff4a3/server/etcdserver/cindex/cindex.go#L58) still holds the old backend, obviously it isn't correct.

This issue is resolved in PRs [pull/13500](https://github.com/etcd-io/etcd/pull/13500) (for 3.6) and [pull/13501](https://github.com/etcd-io/etcd/pull/13501) (for 3.5).

If an etcd node runs into this issue, the local db file is replaced by the *.snap.db automatically. So when etcd restarts, then this issue disappears.
