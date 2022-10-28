issue https://github.com/etcd-io/etcd/issues/14571
======

## Table of Contents

- **[Background](#background)**
- **[How to reproduce this issue](#how-to-reproduce-this-issue)**
- **[Root cause](#root-cause)**
- **[How to fix this issue](#how-to-fix-this-issue)**
- **[How to workaround this issue](#how-to-workaround-this-issue)**

## Background
This is a regression issue introduced in `main` by [pull/13954](https://github.com/etcd-io/etcd/pull/13954),
which was backported to 3.4.20 in [pull/14230](https://github.com/etcd-io/etcd/pull/14230) and to 3.5.5 in
[pull/14227](https://github.com/etcd-io/etcd/pull/14227).

So the affected versions are **3.4.20, 3.4.21 and 3.5.5**.
**Note: there is no any impact if auth isn't enabled.**

This issue will be resolved in 3.4.22, 3.5.6 and of course `main` branch. The PR [pull/14574](https://github.com/etcd-io/etcd/pull/14574)
has already been merged into `main` branch, and will be backported to release-3.5 and release-3.4 very soon.

This issue can only happen when auth is enabled. With the auth enabled, when adding a new member 
into a cluster, then the new member might be running into inconsistency issue.

Note that network partition might also lead to this issue. If users setup & enable auth during network partition, 
when the isolated member rejoins the cluster, it may recover from a snapshot, and so run into this issue as well.
Usually I don't think users will do this operation (setup & enable auth during network partition), so it should 
be low possibility to run into this issue due to network partition.

## How to reproduce this issue
Steps:
1. Start an etcd cluster with 3 members, and each member is started with `--snapshot-count=10`.
2. Setup roles, users and enable auth per guide https://etcd.io/docs/v3.5/demo/. The commands are roughly something like below,
```
$ etcdctl role add root
$ etcdctl user add root:root
$ etcdctl user grant-role root root

$ etcdctl role add role0
$ etcdctl role grant-permission role0 readwrite foo
$ etcdctl user add user0:user0
$ etcdctl user grant-role user0 role0

$ etcdctl auth enable
```
3. Write more than 10 K/V entries to trigger at least one snapshot. Let's add `foo`:`bar` as well.
```
$ for i in {1..10}; do etcdctl put k$i v$i --user root:root; done

$ etcdctl put foo bar --user user0:user0
```
4. Add a new member into the cluster, and start the member.
```
$ etcdctl  member add infra4 --peer-urls=http://127.0.0.1:42380 --user root:root

$ etcd --name infra4 --listen-client-urls http://127.0.0.1:42379 --advertise-client-urls http://127.0.0.1:42379 --listen-peer-urls http://127.0.0.1:42380 --initial-advertise-peer-urls http://127.0.0.1:42380 --initial-cluster-token etcd-cluster-1 --initial-cluster 'infra1=http://127.0.0.1:12380,infra2=http://127.0.0.1:22380,infra3=http://127.0.0.1:32380,infra4=http://127.0.0.1:42380' --initial-cluster-state existing --enable-pprof --logger=zap --log-outputs=stderr --snapshot-count 10
```
Please note that the new member will recover from a snapshot coming from the leader.

5. Update the value for key `foo` to `bar2`. Please make sure the `etcdctl` doesn't connect to the new added member. You will get an `OK` response.
```
$ etcdctl put foo bar2 --user user0:user0
```
6. Restart the new added member. Try to read the key `foo`, you will get stale data from the new added member, and can get new value from any other members.
```
$ etcdctl get foo  --user user0:user0 --endpoints http://127.0.0.1:42379
foo
bar

$ etcdctl get foo  --user user0:user0 --endpoints http://127.0.0.1:32379
foo
bar2

$ etcdctl get foo  --user user0:user0 --endpoints http://127.0.0.1:22379
foo
bar2

$ etcdctl get foo  --user user0:user0 --endpoints http://127.0.0.1:2379
foo
bar2
```

The new added member's revision is also not consistent with other members.
```
$ etcdctl endpoint status --cluster -w json --user root:root | jq
[
  {
    "Endpoint": "http://127.0.0.1:42379",
    "Status": {
      "header": {
        "cluster_id": 17237436991929494000,
        "member_id": 643426431064253700,
        "revision": 2,
        "raft_term": 2
      },
      "version": "3.5.5",
      "dbSize": 24576,
      "leader": 10501334649042878000,
      "raftIndex": 86,
      "raftTerm": 2,
      "raftAppliedIndex": 86,
      "dbSizeInUse": 24576
    }
  },
  {
    "Endpoint": "http://127.0.0.1:2379",
    "Status": {
      "header": {
        "cluster_id": 17237436991929494000,
        "member_id": 9372538179322590000,
        "revision": 3,
        "raft_term": 2
      },
      "version": "3.5.5",
      "dbSize": 24576,
      "leader": 10501334649042878000,
      "raftIndex": 87,
      "raftTerm": 2,
      "raftAppliedIndex": 87,
      "dbSizeInUse": 24576
    }
  },
  {
    "Endpoint": "http://127.0.0.1:22379",
    "Status": {
      "header": {
        "cluster_id": 17237436991929494000,
        "member_id": 10501334649042878000,
        "revision": 3,
        "raft_term": 2
      },
      "version": "3.5.5",
      "dbSize": 24576,
      "leader": 10501334649042878000,
      "raftIndex": 88,
      "raftTerm": 2,
      "raftAppliedIndex": 88,
      "dbSizeInUse": 24576
    }
  },
  {
    "Endpoint": "http://127.0.0.1:32379",
    "Status": {
      "header": {
        "cluster_id": 17237436991929494000,
        "member_id": 18249187646912140000,
        "revision": 3,
        "raft_term": 2
      },
      "version": "3.5.5",
      "dbSize": 24576,
      "leader": 10501334649042878000,
      "raftIndex": 89,
      "raftTerm": 2,
      "raftAppliedIndex": 89,
      "dbSizeInUse": 24576
    }
  }
]
```
## Root cause
The root cause is pretty simple. etcd doesn't load the auth info (e.g. user & role) from db when restoring from a snapshot, so it will definitely fail
to apply data due to permission denied. The error message is something like,
```
{"level":"error","ts":"2022-10-28T13:00:47.056+0800","caller":"auth/range_perm_cache.go:114","msg":"user doesn't exist","user-name":"user0","stacktrace":"go.etcd.io/etcd/server/v3/auth.(*authStore).isRangeOpPermitted\n\t/tmp/etcd-release-3.5.5/etcd/release/etcd/server/auth/range_perm_cache.go:114\ngo.etcd.io/etcd/server/v3/auth.(*authStore).isOpPermitted\n\t/tmp/etcd-release-3.5.5/etcd/release/etcd/server/auth/store.go:870\ngo.etcd.io/etcd/server/v3/auth.(*authStore).IsPutPermitted\n\t/tmp/etcd-release-3.5.5/etcd/release/etcd/server/auth/store.go:878\ngo.etcd.io/etcd/server/v3/etcdserver.(*authApplierV3).Put\n\t/tmp/etcd-release-3.5.5/etcd/release/etcd/server/etcdserver/apply_auth.go:68\ngo.etcd.io/etcd/server/v3/etcdserver.(*applierV3backend).Apply\n\t/tmp/etcd-release-3.5.5/etcd/release/etcd/server/etcdserver/apply.go:171\ngo.etcd.io/etcd/server/v3/etcdserver.(*authApplierV3).Apply\n\t/tmp/etcd-release-3.5.5/etcd/release/etcd/server/etcdserver/apply_auth.go:61\ngo.etcd.io/etcd/server/v3/etcdserver.(*EtcdServer).applyEntryNormal\n\t/tmp/etcd-release-3.5.5/etcd/release/etcd/server/etcdserver/server.go:2241\ngo.etcd.io/etcd/server/v3/etcdserver.(*EtcdServer).apply\n\t/tmp/etcd-release-3.5.5/etcd/release/etcd/server/etcdserver/server.go:2143\ngo.etcd.io/etcd/server/v3/etcdserver.(*EtcdServer).applyEntries\n\t/tmp/etcd-release-3.5.5/etcd/release/etcd/server/etcdserver/server.go:1384\ngo.etcd.io/etcd/server/v3/etcdserver.(*EtcdServer).applyAll\n\t/tmp/etcd-release-3.5.5/etcd/release/etcd/server/etcdserver/server.go:1199\ngo.etcd.io/etcd/server/v3/etcdserver.(*EtcdServer).run.func8\n\t/tmp/etcd-release-3.5.5/etcd/release/etcd/server/etcdserver/server.go:1122\ngo.etcd.io/etcd/pkg/v3/schedule.(*fifo).run\n\t/tmp/etcd-release-3.5.5/etcd/release/etcd/pkg/schedule/schedule.go:157"}
```

## How to fix this issue
It's also pretty simple, we just need to load the auth info when recovering from a snapshot. It's just one line code change.
Please refer to [pull/14574](https://github.com/etcd-io/etcd/pull/14574)

## How to workaround this issue
1. If you are already on the affected versions (3.4.20, 3.4.21, 3.5.5) and auth is enabled, then please do not add new member until you upgrade to 3.4.22 or 3.5.6.
2. If you already run into this issue, then remove the problematic member (it should be the new added member) from the cluster firstly 
(Note: you need to clean up the member's data after removing the member), and upgrade to 3.4.22 or 3.5.6. Afterwards you can add the member back if you want.
