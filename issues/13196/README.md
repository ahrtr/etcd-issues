issue https://github.com/etcd-io/etcd/issues/13196
======
# Background
Reproduced on etcd version: 3.5.0.

When upgrading etcd from an old version to 3.5.0, then some zombie members may be displayed. In the following example, 
"etcdctl endpoint status" shows 3 members and it's correct. However "etcdctl member list" shows 11 members, obviously 
there are 8 zombie members.
```
$ etcdctl endpoint status
cfg01.prod:2379, b9428e2a6571f403, 3.5.0, 103 MB, false, 8048, 65770996
cfg02.prod:2379, 71d808e4df973c61, 3.5.0, 106 MB, true, 8048, 65770996
cfg03.prod:2379, 24cfcc8c8928a23, 3.5.0, 105 MB, false, 8048, 65770997

$ etcdctl member list
24cfcc8c8928a23, started, cfg03, http://10.10.21.26:2380, http://10.10.21.26:2379,http://127.0.1.1:2379
57179c3602d3017, started, cfg03, http://10.10.21.26:2380, http://10.10.21.26:2379
71d808e4df973c61, started, cfg02, http://10.10.21.25:2380, http://10.10.21.25:2379,http://127.0.1.1:2379
774d4084c61cdba9, started, cfg02, http://10.10.21.25:2380, http://10.10.21.25:2379
a53e9eafb0791018, started, cfg01, http://10.10.21.27:2380, http://10.10.21.27:2379,http://127.0.1.1:2379
b05ffc3c3e406282, started, cfg01, http://10.10.21.27:2380, http://10.10.21.27:2379
b9428e2a6571f403, started, cfg01, http://10.10.21.27:2380, http://10.10.21.27:2379,http://127.0.1.1:2379
cfaa8b1d45d54d63, started, cfg03, http://10.10.21.26:2380, http://10.10.21.26:2379
e2e14bd3143026eb, started, cfg03, http://10.10.21.26:2380, http://10.10.21.26:2379,http://127.0.1.1:2379
f41b8cf40578fb5b, started, cfg02, http://10.10.21.25:2380, http://10.10.21.25:2379
fb45dc04aab58183, started, cfg01, http://10.10.21.27:2380, http://10.10.21.27:2379
```

There also will be some error messages in the log file,
```
{"level":"warn","ts":"2021-06-16T08:40:21.800Z","caller":"rafthttp/probing_status.go:68","msg":"prober detected unhealthy status","round-tripper-name":"ROUND_TRIPPER_SNAPSHOT","remote-peer-id":"8e9e05c52164694d","rtt":"0s","error":"dial tcp 192.168.0.9:2380: connect: no route to host"}
```

The zombie members can't even be manually removed, 
```
roachbait:~% kubectl -n kube-system exec -ti etcd-2 -- etcdctl member remove 407604e04d511f63
{"level":"warn","ts":"2021-06-16T08:42:46.434Z","logger":"etcd-client","caller":"v3/retry_interceptor.go:62","msg":"retrying of unary invoker failed","target":"etcd-endpoints://0xc0002f08c0/#initially=[https://192.168.0.129:2379/]","attempt":0,"error":"rpc error: code = NotFound desc = etcdserver: member not found"}
Error: etcdserver: member not found
command terminated with exit code 1
```

# Analysis
Old etcd versions (<3.5.0) load members from v2store, and WAL files are replayed against the latest snapshot, everything is 
working well. But in etcd 3.5.0, the backend bolt db (store-v3) is considered authoritative; etcd loads the member info from the 
db file, which may not be consistent with the [v2store + WAL files] probably due to issues originating from somewhere in v3.1 ~ v3.3. 

A workaround solution **[pull/13348](https://github.com/etcd-io/etcd/pull/13348)** is included in etcd 3.5.1. 
When users run into this issue, they can manually remove the zombie members. Each time when removing a member, 
it removes the member from both the v2 and v3 store. Similarly, when adding a member, it adds the member to both as well. 

A formal fix is supposed to be available in etcd 3.5.2. 

In etcd 3.6, the v2 store will be completely removed. 

This issue is somewhat similar to another one **[13466](../13466)**.