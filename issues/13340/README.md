issue https://github.com/etcd-io/etcd/issues/13340
======
etcd version: 3.5.

Currently, etcd performs a linearizable read when processing the `/health` requests, so it needs to get the consensus from all members. 
The behaviour may not be correct, because kubelet may restart an etcd POD when the cluster isn't healthy, such as no leader. 

Kubernetes Probes (i.e. livenessProbe) use "/health" endpoint to make a decision whether to restart a specific container.
In this case, it should only check local etcd member's health instead of etcd cluster's health. When the cluster isn't healthy,
such as no raft leader, restarting the local etcd member cannot help, and it could even make the situation worse. So the endpoint
should provide an option to let users choose to check the etcd cluster's health or local etcd member's health.
The default behaviour is to check the etcd cluster's health so as to be backward compatible.
When the query parameter "serializable=true" is provided, then etcdserver should only do serializable read, which means
it checks local etcd member's health instead of the etcd cluster's health.

I submitted a PR **[pull/13399](https://github.com/etcd-io/etcd/pull/13399)** to fix this issue. The solution is to add a query parameter "serializable=true" to the `/health` endpoint. 
Please get more detailed info in the PR. 

