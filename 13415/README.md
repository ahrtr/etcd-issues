issue https://github.com/etcd-io/etcd/issues/13415
======
# Issue 
When authentication is enabled, then the etcd POD kept being restarted. The health check request on /health failed with etcd 3.5, but it works for 3.4. 

etcd 3.4
```
$ etcdctl auth enable
Authentication Enabled

$ curl http://127.0.0.1:2381/health
{"health":"true"}
```

etcd 3.5
```
$ etcdctl auth enable
Authentication Enabled

$ curl http://127.0.0.1:2381/health
{"health":"false","reason":"RANGE ERROR:auth: user name is empty"}
```

# Root cause & Solution
Firstly, when the auth is enabled, then all requests will & should be authenticated, including the /health request. So I think the behavior you observed should be expected. 
Please note that there are big improvements in 3.5 against 3.4. I agree that the user experience changes on the /health request between the two versions when auth is enabled, 
but it shouldn't be a problem, because etcd can get the autoInfo from certificate automatically (see below).

Secondly, etcd gets the autoInfo from the context firstly; if it doesn't exist, then it will get the autoInfo from the certificate. Please see **[v3_server.go#L896](https://github.com/etcd-io/etcd/blob/519f62b269cbc5f0438587cdcd9e3d4653c6515b/server/etcdserver/v3_server.go#L896)**.

So the solution is to set the scheme as "HTTPS" in the livenessProbe and startupProbe, so that etcd can get the autoInfo from the certificate.
