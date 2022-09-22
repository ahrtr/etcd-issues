package main

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"time"

	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv2 "go.etcd.io/etcd/client/v2"
)

type v2discovery struct {
	token string
	c     clientv2.KeysAPI
}

// Refer to https://github.com/etcd-io/etcd/issues/14447
// The program below is working well, and the result is:
//    size: 3
//
// The issue(14447) is due to ca-certificates package isn't included in
// debian base image.
// Refer to https://github.com/debuerreotype/docker-debian-artifacts/issues/15
func main() {
	durl := "https://discovery.etcd.io/d6db9ed5ff85dac2466be83973194203"

	disc, err := newV2Discovery(durl)
	if err != nil {
		panic(err)
	}

	// https://discovery.etcd.io/d6db9ed5ff85dac2466be83973194203/_config/size
	configKey := path.Join("/", disc.token, "_config")
	resp, err := disc.c.Get(context.TODO(), path.Join(configKey, "size"), nil)
	if err != nil {
		panic(err)
	}

	size, err := strconv.ParseUint(resp.Node.Value, 10, 0)
	if err != nil {
		panic(err)
	}
	fmt.Printf("size: %d\n", size)
}

func newV2Discovery(durl string) (*v2discovery, error) {
	u, err := url.Parse(durl)
	if err != nil {
		return nil, err
	}
	token := u.Path
	u.Path = ""

	tr, err := transport.NewTransport(transport.TLSInfo{}, 30*time.Second)

	cfg := clientv2.Config{
		Transport: tr,
		Endpoints: []string{u.String()},
	}

	c, err := clientv2.New(cfg)
	if err != nil {
		return nil, err
	}
	dc := clientv2.NewKeysAPIWithPrefix(c, "")

	return &v2discovery{
		c:     dc,
		token: token,
	}, nil
}
