issue https://github.com/etcd-io/etcd/issues/14370
======

## Table of Contents

- **[Background](#background)**
- **[How to reproduce this issue](#how-to-reproduce-this-issue)**
- **[Root cause](#root-cause)**
- **[Why multi-member cluster doesn't have this issue](#why-multi-member-cluster-doesnt-have-this-issue)**
- **[How to fix this issue](#how-to-fix-this-issue)**
- **[How to workaround this issue](#how-to-workaround-this-issue)**

## Background
This is a legacy issue in all releases, including 3.4, 3.5 and main. We will 
resolve this issue in 3.4.21, 3.5.5 and main respectively.

The issue can only happen in one-member cluster. The etcd instance might run
into a situation that a client gets a success response to the write request, 
but the data is actually lost when the etcd instance crashes immediately
after it responds to the client but before it successfully persists the data in
both WAL file and BoltDB db.

## How to reproduce this issue
Steps:
1. Get the gofail runtime package for both etcdserver and etcdutl. Specifically, run command below in both `server` and `etcdutl`,
```
$ go get go.etcd.io/gofail/runtime   # execute this command for both server and etcdutl
```
2. Build etcd with failpoint and start etcd
```
$ FAILPOINTS=enable make
$ GOFAIL_HTTP="127.0.0.1:22381" ./bin/etcd 
```
3. Trigger the failpoints <br>
Note: run this step and the following steps in a separate terminal,
```
curl http://127.0.0.1:22381/etcdserver/raftBeforeLeaderSend -XPUT -d'sleep(100)'
curl http://127.0.0.1:22381/etcdserver/raftBeforeSave -XPUT -d'panic'
curl http://127.0.0.1:22381/backend/beforeCommit -XPUT -d'sleep(200)'
```
4. Send a put request to etcdserver
```
$ ./bin/etcdctl  put k1 v1
```
Please note that the client(etcdctl) will get an "OK" response, and the etcd crashes.

5. Start etcd again and get the data <br>
```
$ ./bin/etcd
$ ./bin/etcdctl  get k1   ## no data because the data was lost
```
You will get nothing because the data has already been lost. 
<br>
Note:
1. You can refer to the original's reporters' [steps](https://github.com/etcd-io/etcd/issues/14370#issue-1346593247) to reproduce this issue as well.
2. Please read [gofail/design](https://github.com/etcd-io/gofail/blob/master/doc/design.md) to get more detailed info on the gofail project.

## Root cause
The leader updates the [pr.Match](https://github.com/etcd-io/etcd/blob/4d57eb8d3b2190bfdba3c65f8eb93c0349fc6dcc/raft/raft.go#L638) 
and marks an entry as [committed](https://github.com/etcd-io/etcd/blob/4d57eb8d3b2190bfdba3c65f8eb93c0349fc6dcc/raft/raft.go#L640) 
immediately after appending it to the unstable logs for one-member cluster.

Afterwards, it sends identical [Entries](https://github.com/etcd-io/etcd/blob/4d57eb8d3b2190bfdba3c65f8eb93c0349fc6dcc/raft/node.go#L71) 
and [CommittedEntries](https://github.com/etcd-io/etcd/blob/4d57eb8d3b2190bfdba3c65f8eb93c0349fc6dcc/raft/node.go#L79)
to etcdserver via the Ready channel. After receiving the ready data, etcdserver persists the `Entries` to the WAL file 
concurrently with applying the `CommittedEntries`, and responds to the client immediately after it finishes the applying workflow.

Unfortunately, it doesn't mean that the data has been successfully saved in both WAL file and BoltDB db 
when the client receives the response, see reasons below. If the etcd instance crashes after the client gets the response but before etcdserver 
successfully persists the `Entries` in WAL file and `CommittedEntries` in boltDB, then the data loss is lost. 
1. etcdserver persists the `Entries` concurrently with the applying workflow.
2. etcdserver commits the boltDB transaction periodically instead of on each request.

## Why multi-member cluster doesn't have this issue
It's a little complicated. In short, it's possible for the leader to mark an entry as committed before it successfully being 
persisted in the WAL files of majority members, but there is no chance for the leader to broadcast the new commitId to any 
follower before the entry is successfully persisted in its local WAL file and of course WAL files of majority members. 

Please also refer to [14370#issuecomment-1232584558](https://github.com/etcd-io/etcd/issues/14370#issuecomment-1232584558) 
to get deeper understanding if you are interested.

## How to fix this issue
I delivered 4 solutions/PRs for this issue,
1. [etcd/pull/14394](https://github.com/etcd-io/etcd/pull/14394)
2. [etcd/pull/14400](https://github.com/etcd-io/etcd/pull/14400)
3. [etcd/pull/14407](https://github.com/etcd-io/etcd/pull/14407)
4. [etcd/pull/14411](https://github.com/etcd-io/etcd/pull/14411)

[etcd/pull/14411](https://github.com/etcd-io/etcd/pull/14411) is the best one among all the four solutions. But I suggested to fix the issue 
for release-3.5 and 3.4 using [etcd/pull/14400](https://github.com/etcd-io/etcd/pull/14400) for safety. It isn't finalized yet.
Please refer to [issues/14370#issuecomment-1232560340](https://github.com/etcd-io/etcd/issues/14370#issuecomment-1232560340) to get detailed info.

Piotr Tabor (ptabor@) and Tobias Grieger (tbg@) also raised two draft PRs for this issue. 
1. [etcd/pull/14406](https://github.com/etcd-io/etcd/pull/14406), it's ptabor's PR and already closed.
2. [etcd/pull/14413](https://github.com/etcd-io/etcd/pull/14413), it's tbg's draft PR which is still in progress of development.

Please see [14370#issuecomment-1235091312](https://github.com/etcd-io/etcd/issues/14370#issuecomment-1235091312) 
to get my personal analysis & comparison between [etcd/pull/14411](https://github.com/etcd-io/etcd/pull/14411) 
and [etcd/pull/14413](https://github.com/etcd-io/etcd/pull/14413). We are still waiting for the feedback from Tobias Grieger and Ben Darnell, who 
are the maintainers of the etcd raft package. Please also refer to [14370#issuecomment-1235790496](https://github.com/etcd-io/etcd/issues/14370#issuecomment-1235790496)

## How to workaround this issue
1. Please try to set up multi-member cluster in production environment.
2. We haven't received any real issue coming from production environment for single-member cluster yet. You have to upgrade to 3.5.5 (to be released) or 3.4.21 (to be released) if you have to set up one-member cluster.