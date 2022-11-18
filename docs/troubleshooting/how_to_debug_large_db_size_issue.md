How to debug large db size issue?
======
<span style="color: #808080; font-family: Babas; font-size: 1em;">
ahrtr@github <br>
November 17th, 2022
</span>

# Background
Users can configure the quota of the backend db size using flag `--quota-backend-bytes`. It's the max number of bytes 
the etcd db file may consume, namely the ${etcd-data-dir}/member/snap/db file. Its default value is 2GB, and the 
suggested max value is 8GB.

2GB is usually sufficient for most use cases. If you run out of the db quota, it would be better to figure out 
whether it's expected. It's exactly the reason why I provide this guide.

# How to check db size
The easiest way is to execute `ls -lrt ${etcd-data-dir}/member/snap` command directly. For example, the db file is
5148672 bytes in the following example,
```
# ls -lrt /path-2-db-data-dir/member/snap/
total 5108
-rw-r--r-- 1 vcap vcap   13503 Nov 15 00:48 000000000000001f-0000000000c0433e.snap
-rw-r--r-- 1 vcap vcap   13503 Nov 15 12:33 000000000000001f-0000000000c1c9df.snap
-rw-r--r-- 1 vcap vcap   13502 Nov 16 00:17 000000000000001f-0000000000c35080.snap
-rw-r--r-- 1 vcap vcap   13503 Nov 16 12:03 000000000000001f-0000000000c4d721.snap
-rw-r--r-- 1 vcap vcap   13502 Nov 16 23:52 0000000000000023-0000000000c65dc2.snap
-rw------- 1 vcap vcap 5148672 Nov 17 02:44 db
```

The second way to check db size is to execute `etcdctl endpoint status` command. For example, it's about 5.1MB in the following example. 
Note that if you do not set the `--cluster` flag, then you need to get all members' endpoints included in `--endpoints`, e.g, 
`--endpoints https://etcd-0:2379,https://etcd-1:2379,https://etcd-2:2379`.
```
# etcdctl --cacert /etc/kubernetes/pki/etcd/ca.crt --cert /etc/kubernetes/pki/etcd/server.crt --key /etc/kubernetes/pki/etcd/server.key --endpoints https://etcd-0:2379 endpoint status -w table --cluster
+------------------------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
|                 ENDPOINT                 |        ID        | VERSION | DB SIZE | IS LEADER | IS LEARNER | RAFT TERM | RAFT INDEX | RAFT APPLIED INDEX | ERRORS |
+------------------------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
| https://master-0.etcd.cfcr.internal:2379 | 17f206fd866fdab2 |   3.5.4 |  5.2 MB |     false |      false |        35 |   13024257 |           13024257 |        |
| https://master-2.etcd.cfcr.internal:2379 | 604ea1193b383592 |   3.5.4 |  5.1 MB |     false |      false |        35 |   13024257 |           13024257 |        |
| https://master-1.etcd.cfcr.internal:2379 | 9dccb73515ee278f |   3.5.4 |  5.1 MB |      true |      false |        35 |   13024257 |           13024257 |        |
+------------------------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
```

You can also display the output using json format (See example below). The benefit is that you can get the both `dbSize` 
and `dbSizeInUse`. Let's use the first endpoint as an example, its `dbSize` is 5177344 bytes, and `dbSizeInUse` is
2039808 bytes. **It means that there are (5177344-2039808) bytes of free space, which can be reclaimed by executing the defragmentation operation**.
```
# etcdctl --cacert /etc/kubernetes/pki/etcd/ca.crt --cert /etc/kubernetes/pki/etcd/server.crt --key /etc/kubernetes/pki/etcd/server.key --endpoints https://etcd-0:2379 endpoint status -w json --cluster
[
  {
    "Endpoint": "https://master-0.etcd.cfcr.internal:2379",
    "Status": {
      "header": {
        "cluster_id": 7895810959607866000,
        "member_id": 1725449293188291300,
        "revision": 9662972,
        "raft_term": 35
      },
      "version": "3.5.4",
      "dbSize": 5177344,
      "leader": 11370664597832739000,
      "raftIndex": 13024242,
      "raftTerm": 35,
      "raftAppliedIndex": 13024242,
      "dbSizeInUse": 2039808
    }
  },
  {
    "Endpoint": "https://master-2.etcd.cfcr.internal:2379",
    "Status": {
      "header": {
        "cluster_id": 7895810959607866000,
        "member_id": 6939661205564306000,
        "revision": 9662972,
        "raft_term": 35
      },
      "version": "3.5.4",
      "dbSize": 5148672,
      "leader": 11370664597832739000,
      "raftIndex": 13024242,
      "raftTerm": 35,
      "raftAppliedIndex": 13024242,
      "dbSizeInUse": 2031616
    }
  },
  {
    "Endpoint": "https://master-1.etcd.cfcr.internal:2379",
    "Status": {
      "header": {
        "cluster_id": 7895810959607866000,
        "member_id": 11370664597832739000,
        "revision": 9662972,
        "raft_term": 35
      },
      "version": "3.5.4",
      "dbSize": 5144576,
      "leader": 11370664597832739000,
      "raftIndex": 13024242,
      "raftTerm": 35,
      "raftAppliedIndex": 13024242,
      "dbSizeInUse": 2039808
    }
  }
]
```

# Compaction & Defragmentation
etcd supports MVCC(Multi-Version Concurrent Control), and it keeps an exact history of its key spaces.
The compaction operation is the only way to purge history. But the free space will not be reclaimed automatically,
instead you should perform defragmentation to reclaim the free space after the compaction operation. 
So usually Compaction + Defragmentation is the right way to purge history and reclaim unused storage space.

Please follow guide [maintenance/#space-quota](https://etcd.io/docs/v3.5/op-guide/maintenance/#space-quota)
to manually perform compaction and defragmentation operations.

Notes:
1. Compaction is a cluster-wide operation, so you only need to execute compaction once on whichever etcd member. 
Of course, it will not do any harm if executing it multiple times.
2. Defragmentation is a time-consuming task, so it's recommended to do it for each member one by one.
3. **There is a known issue that etcd might run into data inconsistency issue if it crashes in the middle of an online 
defragmentation operation using `etcdctl` or clientv3 API. All the existing v3.5 releases are affected, including 3.5.0 ~ 3.5.5. 
So please use `etcdutl` to offline perform defragmentation operation**, but this requires taking each member offline one at a time. 
It means that you need to stop each etcd instance firstly, then perform defragmentation using `etcdutl`, start the instance at last.
Please refer to the issue 1 in [public statement](https://docs.google.com/document/d/1q6PausGMsj-ZyqN2Zx0W8426KsB5GEh3XA801JxBCiE/edit).

The following example shows you how to execute defragmention using `etcdutl`,
```
$ etcdutl defrag --data-dir  ~/tmp/etcd/infra1.etcd/
```

# What data occupies most of the storage space
If the compaction + defragmentation doesn't help; in other words, if the db size is still exceeding
or close to the quota, then you need to figure out what data is consuming most of the storage space.

The straightforward way is to count the number of each kind of object. If the etcd cluster is supporting
Kubernetes apiserver, then execute command below to do the statistics. 

In the following example, there are about 971K events, so obvious the events occupy most of the storage space. 
**The next step is to use your best judgement to figure out the root cause**. Is it expected? Have you or your user
change the value of `--event-ttl` (apiserver flag, defaults to 1h)? Is it running into a known issue (e.g. [107170](https://github.com/kubernetes/kubernetes/issues/107170)  )? 
```
# etcdctl --cacert /etc/kubernetes/pki/etcd/ca.crt --cert /etc/kubernetes/pki/etcd/server.crt --key /etc/kubernetes/pki/etcd/server.key --endpoints https://etcd-0:2379 get /registry --prefix --keys-only | grep -v ^$ | awk -F '/'  '{ h[$3]++ } END {for (k in h) print h[k], k}' | sort -nr
971223 events
51011 pods
10008 replicasets
10008 deployments
3886 services
2555 secrets
2549 serviceaccounts
2514 configmaps
2507 namespaces
1947 endpointslices
506 leases
500 minions
500 csinodes
500 cns.vmware.com
77 clusterroles
64 clusterrolebindings
34 apiregistration.k8s.io
13 podsecuritypolicy
12 rolebindings
12 flowschemas
11 apiextensions.k8s.io
10 roles
8 prioritylevelconfigurations
3 masterleases
2 vmware.com
2 ranges
2 priorityclasses
2 daemonsets
2 controllerrevisions
1 validatingwebhookconfigurations
1 nsx.vmware.com
1 jobs
1 csidrivers
```

If the etcd instance is not running, then you can use [etcd-dump-db](https://github.com/etcd-io/etcd/tree/main/tools/etcd-dump-db)
to do the similar analysis(see example below). Note that you can build the binary `etcd-dump-db` using command `go build` with a golang version 1.19.x.
```
./etcd-dump-db iterate-bucket  /path-2-etcd-data-dir/member/snap/db  key --decode | egrep -o '"/registry.*' | cut -d'|' -f1 | grep -v ^$ | awk -F '/'  '{ h[$3]++ } END {for (k in h) print h[k], k}' | sort -nr
630 leases
150 masterleases
82 nsx.vmware.com
77 clusterroles
64 clusterrolebindings
54 secrets
48 serviceaccounts
33 apiregistration.k8s.io
18 pods
13 podsecuritypolicy
12 rolebindings
12 flowschemas
12 configmaps
10 services
10 roles
9 apiextensions.k8s.io
8 replicasets
8 prioritylevelconfigurations
8 minions
8 deployments
6 namespaces
5 endpointslices
3 csinodes
2 ranges
2 priorityclasses
2 daemonsets
2 controllerrevisions
1 validatingwebhookconfigurations
1 jobs
```

# Solutions
If the behavior (db size exceeds the quota) is expected, then you can set a bigger value for `--quota-backend-bytes`. 
You need to make sure your cloud provider supports this, otherwise the manual update might not survive across cluster 
upgrading or VM recreating. **Note that etcd (actually boltDB) maps the db file into memory directly, so a larger value 
also means more memory usage**. Just I mentioned in the beginning of this post, the suggested max value is 8GB. Of course,
If your VM has big memory (e.g. 64GB), it's OK to set a value > 8GB.

The other solution is to set per-resource etcd servers overrides using apiserver flag `--etcd-servers-overrides`. 
In the following example, there are two etcd clusters; one for the normal usage, and the other dedicated to events.
```
--etcd-servers="https://etcd-0:2379,https://etcd-1:2379,https://etcd-2:2379" \
--etcd-servers-overrides="/events#https://etcd-3:2379,https://etcd-4:2379,https://etcd-5:2379"
```

If the behavior (db size exceeds the quota) isn't expected, then you'd better figure out the root cause and resolve
it firstly. If you insist on apply solutions mentioned above, it can just mitigate the issue instead of resolving it.
