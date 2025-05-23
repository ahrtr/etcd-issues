How to debug frequent leader change issue?
======
<span style="color: #808080; font-family: Babas; font-size: 1em;">
ahrtr@github <br>
April 29th, 2025
</span>

# Background
When you see frequent etcd leader changes, usually it's caused by environment issue(s).

This post shows you how to narrow down the cause of the leader change issue.

# `--heartbeat-interval` and `--election-timeout`

`--heartbeat-interval` is the heartbeat interval, and defaults to 100ms. `--election-timeout` is the election timeout, 
which defaults to 1000ms (1 second).

Improper configuration values can cause leader changes. The value of heartbeat interval is recommended to be
around the max of average round-trip time. The election timeout should be at least 5 times the heartbeat interval.

# Network

There are many tools (i.e. [netshoot](https://github.com/nicolaka/netshoot)) available to check network connectivity
between etcd nodes. This section focuses on etcd's built-in metrics and logs related to network indicators.

- Metrics "`peer_round_trip_time_seconds`".
- Log "[lost TCP streaming connection with remote peer](https://github.com/etcd-io/etcd/blob/88bba4d86e3e107a617f2a06f6f879b665d7c7b9/server/etcdserver/api/rafthttp/stream.go#L194)".
- Log "[peer became inactive (message send to peer failed)](https://github.com/etcd-io/etcd/blob/88bba4d86e3e107a617f2a06f6f879b665d7c7b9/server/etcdserver/api/rafthttp/peer_status.go#L66)".

# Resource (CPU and Memory) and disk I/O

Frequent leader changes are usually caused by network issues, but they can also result from other factors such as resource
pressure (CPU or Memory) or disk I/O problems.

High CPU or memory can also lead to frequent leader changes, so please check the resource consumption on each etcd node.

In addition, slow disk I/O may cause etcd to miss heartbeat or delay raft messages. You can use tool like fio to test the disk
performance, or refer to etcd's built-in metrics and logs for I/O related indicators,
- when it takes WAL more than 1 second to sync data to disk, then you will see warning message something like "[slow fdatasync](https://github.com/etcd-io/etcd/blob/16e1fff519eeff66e626dd15fef399ea2b10b9cc/server/storage/wal/wal.go#L816-L820)".
- Metrics "`wal_fsync_duration_seconds`" and "`backend_commit_duration_seconds`". Normally majority values should be less than 32ms or even 16ms.
