issue https://github.com/etcd-io/etcd/issues/13480
======
## Issue
etcd version: N/A.

While performing testing with etcd, the reporter found he is unable to set time.Duration types if configuring etcd via config file,
which is something like:

```
advertise-client-urls: https://172.16.135.67:2379
client-transport-security:
  cert-file: /var/lib/rancher/k3s/server/tls/etcd/server-client.crt
  client-cert-auth: true
  key-file: /var/lib/rancher/k3s/server/tls/etcd/server-client.key
  trusted-ca-file: /var/lib/rancher/k3s/server/tls/etcd/server-ca.crt
data-dir: /var/lib/rancher/k3s/server/db/etcd
election-timeout: 5000
heartbeat-interval: 500
listen-client-urls: https://172.16.135.67:2379,https://127.0.0.1:2379
listen-metrics-urls: http://127.0.0.1:2381
listen-peer-urls: https://172.16.135.67:2380
log-outputs:
- stderr
logger: zap
name: ck-c7-1-cbad9bf0
peer-transport-security:
  cert-file: /var/lib/rancher/k3s/server/tls/etcd/peer-server-client.crt
  client-cert-auth: true
  key-file: /var/lib/rancher/k3s/server/tls/etcd/peer-server-client.key
  trusted-ca-file: /var/lib/rancher/k3s/server/tls/etcd/peer-ca.crt
grpc-keepalive-timeout: 40s
```

The error message was:
```
error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go struct field configYAML.grpc-keepalive-timeout of type time.Duration
```

The reporter just parses the configuration file via json.Unmarshal, and then starts etcd via the parameters defined in the config file. 
His point is he can set the time something like "40s" on the command line, so he should can set the same format of time in the config file as well.

## Analysis
Actually the issue has nothing do with etcd, and it isn't an issue at all. It is just golang's expected behaviour. 

The flag can parse the time, see [flag.go#L269](https://github.com/golang/go/blob/e8cda0a6c925668972ada40602ada08468fa90dc/src/flag/flag.go#L269), 
but json doesn't support time format something like "40s", see [decode.go#L315-L351](https://github.com/golang/go/blob/master/src/encoding/json/decode.go#L315-L351).

The workaround is to replace the "40s" with "40000000000", it should be parsed as nanoseconds.

See the example in main.go, the output is: 
```
{"name":"ahrtr","age":13,"time":40000000000}
{ahrtr 13 40s}
```
