issue https://github.com/etcd-io/etcd/issues/13553
======
If applications use clientv3, then this issue will never happen.

But If the client application sends data with invalid client-api-version directly to etcdserver via tcp connection, 
then the etcd server may be panic. Accordingly, there is a security concern that the malicious program may take down 
the etcd server. 

The program [app_send_invalid_client_api_version.c](app_send_invalid_client_api_version.c) is the demo malicious application.
It sends invalid client-api-version, which isn't a valid UTF-8 string.

The issue is fixed in PR [pull/13560](https://github.com/etcd-io/etcd/pull/13560), which will be included in 3.6.
