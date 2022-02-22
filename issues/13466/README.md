issue https://github.com/etcd-io/etcd/issues/13466
======
## Steps to reproduce this issue
The steps are performed on etcd version **3.5.0**.
1. Create a new cluster with 3 members, add the option "--snapshot-count 10" for each instance on startup. 
2. Add 15 keys using a command below,
```
$ for i in {1..15}; do etcdctl  put k$i v$i; done
```
3. Remove one member 
```
$ etcdctl member remove fd422379fda50e48
```
4. Restart one etcd instance, then you will see the instance panic.
```
{"level":"panic","ts":"2021-11-19T10:23:36.501+0800","caller":"rafthttp/transport.go:349","msg":"unexpected removal of unknown remote peer","remote-peer-id":"fd422379fda50e48","stacktrace":"go.etcd.io/etcd/server/v3/etcdserver/api/rafthttp.(*Transport).removePeer\n\t/Users/wachao/go/src/go.etcd.io/etcd/server/etcdserver/api/rafthttp/transport.go:349\ngo.etcd.io/etcd/server/v3/etcdserver/api/rafthttp.(*Transport).RemovePeer\n\t/Users/wachao/go/src/go.etcd.io/etcd/server/etcdserver/api/rafthttp/transport.go:330\ngo.etcd.io/etcd/server/v3/etcdserver.(*EtcdServer).applyConfChange\n\t/Users/wachao/go/src/go.etcd.io/etcd/server/etcdserver/server.go:2012\ngo.etcd.io/etcd/server/v3/etcdserver.(*EtcdServer).apply\n\t/Users/wachao/go/src/go.etcd.io/etcd/server/etcdserver/server.go:1852\ngo.etcd.io/etcd/server/v3/etcdserver.(*EtcdServer).applyEntries\n\t/Users/wachao/go/src/go.etcd.io/etcd/server/etcdserver/server.go:1078\ngo.etcd.io/etcd/server/v3/etcdserver.(*EtcdServer).applyAll\n\t/Users/wachao/go/src/go.etcd.io/etcd/server/etcdserver/server.go:900\ngo.etcd.io/etcd/server/v3/etcdserver.(*EtcdServer).run.func8\n\t/Users/wachao/go/src/go.etcd.io/etcd/server/etcdserver/server.go:832\ngo.etcd.io/etcd/pkg/v3/schedule.(*fifo).run\n\t/Users/wachao/go/src/go.etcd.io/etcd/pkg/schedule/schedule.go:157"}
panic: unexpected removal of unknown remote peer

goroutine 209 [running]:
go.uber.org/zap/zapcore.(*CheckedEntry).Write(0xc000186cc0, {0xc00022c740, 0x1, 0x1})
	/Users/wachao/go/gopath/pkg/mod/go.uber.org/zap@v1.17.0/zapcore/entry.go:234 +0x499
go.uber.org/zap.(*Logger).Panic(0x1b716c0, {0x1c04d22, 0xc0002c91b0}, {0xc00022c740, 0x1, 0x1})
	/Users/wachao/go/gopath/pkg/mod/go.uber.org/zap@v1.17.0/logger.go:227 +0x59
go.etcd.io/etcd/server/v3/etcdserver/api/rafthttp.(*Transport).removePeer(0xc000174e00, 0xc0002c92d8)
	/Users/wachao/go/src/go.etcd.io/etcd/server/etcdserver/api/rafthttp/transport.go:349 +0x26a
go.etcd.io/etcd/server/v3/etcdserver/api/rafthttp.(*Transport).RemovePeer(0xc000174e00, 0xc00000e018)
	/Users/wachao/go/src/go.etcd.io/etcd/server/etcdserver/api/rafthttp/transport.go:330 +0x85
go.etcd.io/etcd/server/v3/etcdserver.(*EtcdServer).applyConfChange(0xc00010c600, {0x1, 0xfd422379fda50e48, {0x0, 0x0, 0x0}, 0x32697d35fb628418}, 0xc0002b8000, 0x0)
	/Users/wachao/go/src/go.etcd.io/etcd/server/etcdserver/server.go:2012 +0x2c6
go.etcd.io/etcd/server/v3/etcdserver.(*EtcdServer).apply(0xc00010c600, {0xc00033c120, 0x4, 0x0}, 0x0)
	/Users/wachao/go/src/go.etcd.io/etcd/server/etcdserver/server.go:1852 +0x5e5
go.etcd.io/etcd/server/v3/etcdserver.(*EtcdServer).applyEntries(0xc00010c600, 0xc0002b8000, 0xc0003cfb80)
	/Users/wachao/go/src/go.etcd.io/etcd/server/etcdserver/server.go:1078 +0x27d
go.etcd.io/etcd/server/v3/etcdserver.(*EtcdServer).applyAll(0xc00010c600, 0xc0002b8000, 0xc0003cfb80)
	/Users/wachao/go/src/go.etcd.io/etcd/server/etcdserver/server.go:900 +0x65
go.etcd.io/etcd/server/v3/etcdserver.(*EtcdServer).run.func8({0xc00054e790, 0xc00054e758})
	/Users/wachao/go/src/go.etcd.io/etcd/server/etcdserver/server.go:832 +0x25
go.etcd.io/etcd/pkg/v3/schedule.(*fifo).run(0xc0000aa9c0)
	/Users/wachao/go/src/go.etcd.io/etcd/pkg/schedule/schedule.go:157 +0x119
created by go.etcd.io/etcd/pkg/v3/schedule.NewFIFOScheduler
	/Users/wachao/go/src/go.etcd.io/etcd/pkg/schedule/schedule.go:70 +0x15c
```

## Root cause analysis
The value for "--snapshot-count" is 10, so at least a snapshot should have already been created at the step 2. When performing step 3, the raft log (raftpb.ConfChangeRemoveNode) is persisted in the WAL files.

When stopping & starting one etcd instance, it loads the member info from the db file, please see [cluster.go#L257-L263](https://github.com/etcd-io/etcd/blob/946a5a6f25c3b6b89408ab447852731bde6e6289/server/etcdserver/api/membership/cluster.go#L257-L263), 
so the RaftCluster.members has only 2 members, which means the removed member isn't included in the members map. But etcd replays the WAL files based on the latest snapshot, so it will remove the already removed member again, 
accordingly the etcd instance panics, please see [transport.go#L346](https://github.com/etcd-io/etcd/blob/946a5a6f25c3b6b89408ab447852731bde6e6289/server/etcdserver/api/rafthttp/transport.go#L346).

The main branch has this issue as well, because it also loads the member info from the db file firstly, see [cluster.go#L264-L270](https://github.com/etcd-io/etcd/blob/7e0248b367be6417dd6379588a47ea6f278481cf/server/etcdserver/api/membership/cluster.go#L264-L270).

Please note that etcd 3.5.1 doesn't have this issue, because it loads the member info from the v2store firstly, 
see [cluster.go#L259-L265](https://github.com/etcd-io/etcd/blob/77d760bf1bd068c7e8a37b10b029101911875b6b/server/etcdserver/api/membership/cluster.go#L259-L265), so the RaftCluster.members has 3 members,
including the removed member.

## Workaround
The easiest way is to upgrade etcd to 3.5.1. If you don't want to upgrade etcd, then you can follow the steps below to workaround this issue,
1. Backup the etcd binary;
2. Manually change the [log level](https://github.com/etcd-io/etcd/blob/946a5a6f25c3b6b89408ab447852731bde6e6289/server/etcdserver/api/rafthttp/transport.go#L346) from Panic to Warn;
3. Build & replace the binary in your running environment, and add "--snapshot-count 2" to the etcd instance
4. Start the etcd instance;
5. Add a couple of k/v using etcdctl, and them remove the k/v;
6. Stop the etcd instance;
7. Restore the etcd binary backed up at the first step;
8. Start the etcd instance.

PLease note that you may need to perform the above steps for each etcd member.