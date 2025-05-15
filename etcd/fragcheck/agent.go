package main

import (
	"fmt"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func memberList(gcfg globalConfig) (*clientv3.MemberListResponse, error) {
	cfgSpec := clientConfigWithoutEndpoints(gcfg)
	eps, err := endpointsFromCmd(gcfg)
	if err != nil {
		return nil, err
	}
	cfgSpec.Endpoints = eps

	c, err := createClient(cfgSpec)
	if err != nil {
		return nil, err
	}

	ctx, cancel := commandCtx(gcfg.commandTimeout)
	defer func() {
		c.Close()
		cancel()
	}()

	members, err := c.MemberList(ctx)
	if err != nil {
		return nil, err
	}

	return members, nil
}

type epStatus struct {
	Ep   string                   `json:"Endpoint"`
	Resp *clientv3.StatusResponse `json:"Status"`
}

func (es epStatus) String() string {
	return fmt.Sprintf("endpoint: %s, dbSize: %d, dbSizeInUse: %d, memberId: %x, leader: %x, revision: %d, term: %d, index: %d",
		es.Ep, es.Resp.DbSize, es.Resp.DbSizeInUse, es.Resp.Header.MemberId, es.Resp.Leader, es.Resp.Header.Revision, es.Resp.RaftTerm, es.Resp.RaftIndex)
}

func membersStatus(gcfg globalConfig, eps []string) ([]epStatus, error) {
	var statusList []epStatus
	for _, ep := range eps {
		status, err := memberStatus(gcfg, ep)
		if err != nil {
			return nil, fmt.Errorf("failed to get member(%q) status: %w", ep, err)
		}
		statusList = append(statusList, status)
	}

	return statusList, nil
}

func memberStatus(gcfg globalConfig, ep string) (epStatus, error) {
	cfgSpec := clientConfigWithoutEndpoints(gcfg)
	cfgSpec.Endpoints = []string{ep}
	c, err := createClient(cfgSpec)
	if err != nil {
		return epStatus{}, fmt.Errorf("failed to createClient: %w", err)
	}

	ctx, cancel := commandCtx(gcfg.commandTimeout)
	defer func() {
		c.Close()
		cancel()
	}()
	resp, err := c.Status(ctx, ep)

	return epStatus{Ep: ep, Resp: resp}, err
}

func compact(gcfg globalConfig, rev int64, eps []string) error {
	cfgSpec := clientConfigWithoutEndpoints(gcfg)
	cfgSpec.Endpoints = eps
	c, err := createClient(cfgSpec)
	if err != nil {
		return err
	}

	ctx, cancel := commandCtx(gcfg.commandTimeout)
	defer func() {
		c.Close()
		cancel()
	}()

	if rev == 0 {
		resp, rerr := readData(gcfg, eps, "foo")
		if rerr != nil {
			return rerr
		}

		rev = resp.Header.Revision
	}

	_, err = c.Compact(ctx, rev, []clientv3.CompactOption{clientv3.WithCompactPhysical()}...)
	return err
}

func ensureMembersSynced(gcfg globalConfig, eps []string) error {
	for _, ep := range eps {
		if _, rerr := readData(gcfg, []string{ep}, "foo"); rerr != nil {
			return rerr
		}
	}

	return nil
}

func readData(gcfg globalConfig, eps []string, key string) (*clientv3.GetResponse, error) {
	cfgSpec := clientConfigWithoutEndpoints(gcfg)
	cfgSpec.Endpoints = eps
	c, err := createClient(cfgSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to createClient: %w", err)
	}

	ctx, cancel := commandCtx(gcfg.commandTimeout)
	defer func() {
		c.Close()
		cancel()
	}()

	resp, rerr := c.Get(ctx, key)
	return resp, rerr
}

func writeData(gcfg globalConfig, eps []string, key, value string) (*clientv3.PutResponse, error) {
	cfgSpec := clientConfigWithoutEndpoints(gcfg)
	cfgSpec.Endpoints = eps
	c, err := createClient(cfgSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to createClient: %w", err)
	}

	ctx, cancel := commandCtx(gcfg.commandTimeout)
	defer func() {
		c.Close()
		cancel()
	}()

	resp, werr := c.Put(ctx, key, value)
	return resp, werr
}
