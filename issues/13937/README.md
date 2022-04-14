issue https://github.com/etcd-io/etcd/issues/13937
======

## Table of Contents

- **[Background](#background)**
- **[How to reproduce this issue](#how-to-reproduce-this-issue)**
- **[Root cause](#root-cause)**
- **[How is the issue resolved](#how-is-the-issue-resolved)**
- **[How to workaround this issue](#how-to-workaround-this-issue)**

## Background
This is a regression introduced in 3.5.3 in [pull/13908](https://github.com/etcd-io/etcd/pull/13908). But the good news is it shouldn't be a big problem.

If the Auth isn't enabled, then you will not run into this issue. Usually K8s depends on 
certificate/TLS to communicate with etcd, and the Auth isn't enabled. So normally it should be fine.

If the Auth is enabled, it's also not easy to run into this issue, because the default 
value for `--snapshot-count` is 100000. The `consistent_index` will be updated each time 
when there is a successful applying. This issue will be bypassed if there is at least one
successful applying after generating each snapshot.

## How to reproduce this issue
Steps:
1. Start an etcd cluster with 3 members, and with `--snapshot-count 3`;
2. Enable auth (`etcdctl auth enable`);
3. Execute `for i in {1..20}; do etcdctl put k$i v$i; done`, you will see error "user name is empty", and it's expected;
4. Stop all members;
5. Start all members again, then you may see panic below,

```
{"level":"panic","ts":"2022-04-15T05:29:04.867+0800","caller":"etcdserver/server.go:515","msg":"failed to recover v3 backend from snapshot","error":"failed to find database snapshot file (snap: snapshot file doesn't exist)","stacktrace":"go.etcd.io/etcd/server/v3/etcdserver.NewServer\n\t/go/src/go.etcd.io/etcd/release/etcd/server/etcdserver/server.go:515\ngo.etcd.io/etcd/server/v3/embed.StartEtcd\n\t/go/src/go.etcd.io/etcd/release/etcd/server/embed/etcd.go:245\ngo.etcd.io/etcd/server/v3/etcdmain.startEtcd\n\t/go/src/go.etcd.io/etcd/release/etcd/server/etcdmain/etcd.go:228\ngo.etcd.io/etcd/server/v3/etcdmain.startEtcdOrProxyV2\n\t/go/src/go.etcd.io/etcd/release/etcd/server/etcdmain/etcd.go:123\ngo.etcd.io/etcd/server/v3/etcdmain.Main\n\t/go/src/go.etcd.io/etcd/release/etcd/server/etcdmain/main.go:40\nmain.main\n\t/go/src/go.etcd.io/etcd/release/etcd/server/main.go:32\nruntime.main\n\t/go/gos/go1.16.15/src/runtime/proc.go:225"}
panic: failed to recover v3 backend from snapshot

goroutine 1 [running]:
go.uber.org/zap/zapcore.(*CheckedEntry).Write(0xc0000ae0c0, 0xc000532180, 0x1, 0x1)
	/go/pkg/mod/go.uber.org/zap@v1.17.0/zapcore/entry.go:234 +0x58d
go.uber.org/zap.(*Logger).Panic(0xc000710af0, 0x1e32832, 0x2a, 0xc000532180, 0x1, 0x1)
	/go/pkg/mod/go.uber.org/zap@v1.17.0/logger.go:227 +0x85
go.etcd.io/etcd/server/v3/etcdserver.NewServer(0x7ffeefbff874, 0x6, 0x0, 0x0, 0x0, 0x0, 0xc0005626c0, 0x1, 0x1, 0xc000562b40, ...)
	/go/src/go.etcd.io/etcd/release/etcd/server/etcdserver/server.go:515 +0x1656
go.etcd.io/etcd/server/v3/embed.StartEtcd(0xc000020000, 0xc000020600, 0x0, 0x0)
	/go/src/go.etcd.io/etcd/release/etcd/server/embed/etcd.go:245 +0xef8
go.etcd.io/etcd/server/v3/etcdmain.startEtcd(0xc000020000, 0x1e06f77, 0x6, 0xc0005a6001, 0x2)
	/go/src/go.etcd.io/etcd/release/etcd/server/etcdmain/etcd.go:228 +0x32
go.etcd.io/etcd/server/v3/etcdmain.startEtcdOrProxyV2(0xc000001500, 0x18, 0x18)
	/go/src/go.etcd.io/etcd/release/etcd/server/etcdmain/etcd.go:123 +0x257a
go.etcd.io/etcd/server/v3/etcdmain.Main(0xc000001500, 0x18, 0x18)
	/go/src/go.etcd.io/etcd/release/etcd/server/etcdmain/main.go:40 +0x13f
main.main()
	/go/src/go.etcd.io/etcd/release/etcd/server/main.go:32 +0x45
```

## Root cause
etcd checks the permission before applying each request. The credentials (username & password) are not
provided when executing `etcdctl put k v`, so etcd fails to check the permission, accordingly it doesn't 
apply the request at all. It leads to the result the consistent_index isn't moved forward because 
the LockInsideApply method isn't called at all.

## How is the issue resolved
We need to move consistent_index forward when etcd fails to apply a request for whatever reasons. 
The issue is resolved in [pull/13942](https://github.com/etcd-io/etcd/pull/13942) (for main) and [pull/13946](https://github.com/etcd-io/etcd/pull/13946) (for release-3.5).
The fix will be included in 3.5.4 and 3.6.0.

## How to workaround this issue
1. Build a new binary `etcd` with the following patch on **3.5.3**;
```
$ git diff
diff --git a/server/etcdserver/backend.go b/server/etcdserver/backend.go
index 2beef5763..44e583b66 100644
--- a/server/etcdserver/backend.go
+++ b/server/etcdserver/backend.go
@@ -104,6 +104,9 @@ func recoverSnapshotBackend(cfg config.ServerConfig, oldbe backend.Backend, snap
        if snapshot.Metadata.Index <= consistentIndex {
                return oldbe, nil
        }
+       if true {
+               return oldbe, nil
+       }
        oldbe.Close()
        return openSnapshotBackend(cfg, snap.New(cfg.Logger, cfg.SnapDir()), snapshot, hooks)
 }
```
2. Stop the cluster;
3. Replace the binary and start the etcd cluster again (please backup the original binary beforehand);
4. Add & delete a K/V using command `etcdctl put k v` and `etcdctl del k`;
5. Stop the cluster, restore the binary, start the cluster again.

