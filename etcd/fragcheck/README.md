fragcheck
======
fragcheck is a command-line tool for monitoring the disk usage of an etcd database over periodically compaction operations.

Usage:

```
$ ./fragcheck -h
A simple command line tool to analyze etcd fragmentation

Usage:
  fragcheck [flags]

Flags:
      --cacert string                verify certificates of TLS-enabled secure servers using this CA bundle
      --cert string                  identify secure client using this TLS certificate file
      --cluster                      use all endpoints from the cluster member list
      --command-timeout duration     command timeout (excluding dial timeout) (default 30s)
      --compact-interval duration    the interval to perform compaction (default 30s)
      --dial-timeout duration        dial timeout for client connections (default 2s)
  -d, --discovery-srv string         domain name to query for SRV records describing cluster endpoints
      --discovery-srv-name string    service name to query when using DNS discovery
      --endpoints strings            comma separated etcd endpoints (default [127.0.0.1:2379])
  -h, --help                         help for fragcheck
      --insecure-discovery           accept insecure SRV records describing cluster endpoints (default true)
      --insecure-skip-tls-verify     skip server certificate verification (CAUTION: this option should be enabled only for testing purposes)
      --insecure-transport           disable transport security for client connections (default true)
      --keepalive-time duration      keepalive time for client connections (default 2s)
      --keepalive-timeout duration   keepalive timeout for client connections (default 6s)
      --key string                   identify secure client using this TLS key file
      --key-count int                key count (default 500)
      --max-value-size int           max value size (default 120)
      --min-value-size int           min value size (default 100)
      --password string              password for authentication (if this option is used, --user option shouldn't include password)
      --test-duration duration       how long the test should run (default 5m0s)
      --user string                  username[:password] for authentication (prompt if password is not supplied)
      --version                      print the version and exit
```

See an example below,
```
$ ./fragcheck 
Endpoints: [127.0.0.1:2379]
Validating configuration.
{endpoints:[127.0.0.1:2379] useClusterEndpoints:false dialTimeout:2000000000 commandTimeout:30000000000 keepAliveTime:2000000000 keepAliveTimeout:6000000000 insecure:true insecureDiscovery:true insecureSkepVerify:false certFile: keyFile: caFile: dnsDomain: dnsService: username: password: compactInterval:30000000000 testDuration:300000000000 keyCount:500 minValSize:100 maxValSize:120 printVersion:false}
Getting the initial members status
Endpoint: 127.0.0.1:2379, dbSize: 24576, dbSizeInUse: 24576, used percent: 1.00

wrote key: 3695
Endpoint: 127.0.0.1:2379, dbSize: 737280, dbSizeInUse: 176128, used percent: 0.24

wrote key: 3684
Endpoint: 127.0.0.1:2379, dbSize: 884736, dbSizeInUse: 176128, used percent: 0.20

wrote key: 3646
Endpoint: 127.0.0.1:2379, dbSize: 884736, dbSizeInUse: 172032, used percent: 0.19

wrote key: 3714
Endpoint: 127.0.0.1:2379, dbSize: 888832, dbSizeInUse: 167936, used percent: 0.19

wrote key: 3762
Endpoint: 127.0.0.1:2379, dbSize: 897024, dbSizeInUse: 167936, used percent: 0.19

wrote key: 3712
Endpoint: 127.0.0.1:2379, dbSize: 897024, dbSizeInUse: 172032, used percent: 0.19

wrote key: 3662
Endpoint: 127.0.0.1:2379, dbSize: 897024, dbSizeInUse: 180224, used percent: 0.20

wrote key: 3635
Endpoint: 127.0.0.1:2379, dbSize: 897024, dbSizeInUse: 172032, used percent: 0.19

wrote key: 3731
Endpoint: 127.0.0.1:2379, dbSize: 897024, dbSizeInUse: 167936, used percent: 0.19

Done!
```
