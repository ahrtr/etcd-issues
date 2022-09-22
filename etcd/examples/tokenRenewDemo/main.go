package main

import (
	"context"
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

/*
Please see the discussion on the issue https://github.com/etcd-io/etcd/issues/13413.

Test step:
start the etcd server using command "etcd --auth-token simple --auth-token-ttl 5"
etcdctl role add root
etcdctl user add root
etcdctl user  grant-role  root root
etcdctl auth enable
etcdctl --user root:changeme put foo bar
etcdctl --user root:changeme get foo

In another terminal run this program:
go run main.go

Output:
[key:"foo" create_revision:2 mod_revision:2 version:1 value:"bar" ]
{"level":"warn","ts":"2021-10-27T14:00:59.531+0800","logger":"etcd-client","caller":"v3@v3.5.1/retry_interceptor.go:62","msg":"retrying of unary invoker failed","target":"etcd-endpoints://0xc00028b500/127.0.0.1:2379","attempt":0,"error":"rpc error: code = Unauthenticated desc = etcdserver: invalid auth token"}
[key:"foo" create_revision:2 mod_revision:2 version:1 value:"bar" ]
*/
func main() {
	client, err := clientv3.New(clientv3.Config{
		Context:   context.Background(),
		Endpoints: []string{"127.0.0.1:2379"},
		Username:  "root",
		Password:  "root",
		DialOptions: []grpc.DialOption{
			grpc.WithBlock(),
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Get(context.Background(), "test")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.Kvs)

	// The --auth-token-ttl is 5 seconds, so the token will be expired after 6 seconds sleep.
	// The clientv3 is supposed to automatically refresh the token.
	time.Sleep(6 * time.Second)

	resp2, err := client.Get(context.Background(), "test")
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(resp2.Kvs)
	}
}
