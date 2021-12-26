# issue https://github.com/etcd-io/etcd/issues/13554 

When a malicious client sends data to etcdserver with invalid SortTarget, then etcd server will crash due to nil pointer error.

PR [pull/13555](https://github.com/etcd-io/etcd/pull/13555) fixed this issue. The fix will be included in etcd 3.6. 

The client demo is app_send_invalid_sortTarget.c. 
