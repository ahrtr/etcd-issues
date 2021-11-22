etcd-issues 
======
# Overview
This repo contains analysis of various etcd issues from etcd community. The purpose is to provide a central place for others to learn from various practical etcd issues.

# Issue List
| Issue   |      Title      |  Creation Date |  Affected Version | Fixed Version |
|----------|:-------------:|:------:|:------:|:------:|
| **[13340](13340)** |  Provide a better liveness probe for when etcd runs as a Kubernetes pod | 2021-09-10 | 3.5 | 3.6, [pull/13399](https://github.com/etcd-io/etcd/pull/13399)|
| **[13406](13406)** |  SIGBUS on startup in etcd-3.5.0 after filesystem rollback | 2021-10-08 | 3.5 | |
| **[13418](13418)** |  etcd/raft: leader and follower stuck on reject/hint msgs  | 2021-10-15 | 3.5 | |
| **[13415](13415)** |  Kubelet liveness probe restarts etcd when auth is enabled  | 2021-10-13 | 3.5 | |
| **[13466](13466)** |  etcd panic on startup with error message "unexpected removal of unknown remote peer"   | 2021-11-19 | 3.5.0 | |
| **[13480](13480)** |  Unable to specify time.Duration types in etcd config file   | 2021-11-20 | N/A | |
| **[13196](13196)** |  etcd 3.5.0 resurrects ancient (unremovable) members  | 2021-11-22 | 3.5.0 | Workaround fix in 3.5.1, [pull/13348](https://github.com/etcd-io/etcd/pull/13348); <br />Formal fix will be available in 3.5.2; <br />3.6 will deprecate the v2store.|


