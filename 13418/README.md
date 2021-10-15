issue https://github.com/etcd-io/etcd/issues/13418
======
# Issue 
When programmatically constructing an etcd cluster of two members, the leader receives a rejected MsgAppResp from the follower. 

The error log from the follower side is: 
```
2021/10/15 01:08:27 DEBUG: 2 [logterm: 1, index: 2] rejected MsgApp [logterm: 2, index: 2] from 1
```
The error log from the leader side is:
```
2021/10/15 01:08:27 DEBUG: 1 received MsgAppResp(rejected, hint: (index 2, term 1)) from 2 for index 2
2021/10/15 01:08:27 DEBUG: 1 decreased progress of 2 to [StateProbe match=0 next=2]
```

The steps are roughly as below:
```
    start a node A with one peer object define node A
    wait until A become leader
    send ProposeConfChange to add node B
    start node B with 2 peers objects defines node A & B
    results -> reject/hint inf loop
```

# Root cause & Solution
The reason of this issue is the second member should be joining an existing cluster instead of creating a new cluster, which means it should be in the same 
way as adding a flag "--initial-cluster-state existing" from the CLI. From programming perspective, it should call
**[raft.RestartNode](https://github.com/etcd-io/etcd/blob/519f62b269cbc5f0438587cdcd9e3d4653c6515b/raft/node.go#L241)** instead of **[raft.StartNode](https://github.com/etcd-io/etcd/blob/519f62b269cbc5f0438587cdcd9e3d4653c6515b/raft/node.go#L218)**.

Please see the complete example code in main.go, which is updated on top of the code in the issue [issues/13418](https://github.com/etcd-io/etcd/issues/13418). 

```go
go run main.go 
```