package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"go.etcd.io/etcd/client/pkg/v3/srv"
	"golang.org/x/exp/slices"
)

var errBadScheme = errors.New("url scheme must be http or https")

func endpoints(gcfg globalConfig) ([]string, error) {
	if !gcfg.useClusterEndpoints {
		if len(gcfg.endpoints) == 0 {
			return nil, errors.New("no endpoints provided")
		}
		return gcfg.endpoints, nil
	}

	return endpointsFromCluster(gcfg)
}

func endpointsFromCluster(gcfg globalConfig) ([]string, error) {
	memberlistResp, err := memberList(gcfg)
	if err != nil {
		return nil, err
	}

	var eps []string
	for _, m := range memberlistResp.Members {
		// learner member only serves Status and SerializableRead requests, just ignore it
		if !m.GetIsLearner() {
			for _, ep := range m.ClientURLs {
				eps = append(eps, ep)
			}
		}
	}

	slices.Sort(eps)
	eps = slices.Compact(eps)

	return eps, nil
}

func endpointsFromCmd(gcfg globalConfig) ([]string, error) {
	eps, err := endpointsFromDNSDiscovery(gcfg)
	if err != nil {
		return nil, err
	}

	if len(eps) == 0 {
		eps = gcfg.endpoints
	}

	if len(eps) == 0 {
		return nil, errors.New("no endpoints provided")
	}

	return eps, nil
}

func endpointsFromDNSDiscovery(gcfg globalConfig) ([]string, error) {
	if gcfg.dnsDomain == "" {
		return nil, nil
	}

	srvs, err := srv.GetClient("etcd-client", gcfg.dnsDomain, gcfg.dnsService)
	if err != nil {
		return nil, err
	}

	eps := srvs.Endpoints
	if gcfg.insecureDiscovery {
		return eps, nil
	}

	// strip insecure connections
	var ret []string
	for _, ep := range eps {
		if strings.HasPrefix(ep, "http://") {
			fmt.Fprintf(os.Stderr, "ignoring discovered insecure endpoint %q\n", ep)
			continue
		}
		ret = append(ret, ep)
	}
	return ret, nil
}
