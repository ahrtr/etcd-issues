package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.etcd.io/etcd/pkg/v3/pbutil"
	"go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

/*
Issue: https://github.com/etcd-io/etcd/issues/13418

The source is coming from the above issue, but it has issue on the scenario1.
I fixed the issue in the original source code.

The reason of this issue is the second member should be joining an existing cluster instead of creating a new cluster, which means it should be in the same
way as adding a flag "--initial-cluster-state existing" from the CLI. From programming perspective, it should call raft.RestartNode instead of raft.StartNode.

Original author: Sanad Haj Yahya (shaj13)

Updated by Benjamin Wang
Date: 2021-10-15
*/

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	scenario1()
	//scenario2()

	<-sigs
}

/*
Scenario 1
    start a node A with one peer object define node A
    wait until A become leader
    send ProposeConfChange to add node B
    start node B with 2 peers objects defines node A & B
    results -> reject/hint inf loop
*/
func scenario1() {
	n1 := newNode(true, 1)
	n2 := newNode(false, 2, 1)

	n1.tr.remote = n2
	n2.tr.remote = n1

	go n1.run()

	// wait until it become leader
	for n1.rn.Status().Lead == raft.None {
	}

	fmt.Println("The first node is leader now")

	err := n1.rn.ProposeConfChange(context.Background(), raftpb.ConfChange{
		NodeID: 2,
	})

	if err != nil {
		panic(err)
	}

	// start n2
	go n2.run()
}

/*
Scenario 2
    start node A with 2 peer object define node A & B
    start node B with 2 peers objects defines node A & B
    results -> working as expected

*/
func scenario2() {
	n1 := newNode(true, 1, 2)
	n2 := newNode(true, 2, 1)

	n1.tr.remote = n2
	n2.tr.remote = n1

	go n1.run()

	// start n2
	go n2.run()
}

func newNode(newCluster bool, ids ...uint64) *node {
	lg := &raft.DefaultLogger{Logger: log.New(os.Stderr, "", log.LstdFlags)}
	lg.EnableDebug()

	mem := raft.NewMemoryStorage()
	cfg := &raft.Config{
		ElectionTick:              10,
		ID:                        ids[0],
		HeartbeatTick:             1,
		MaxSizePerMsg:             1024 * 1024,
		MaxInflightMsgs:           256,
		MaxUncommittedEntriesSize: 1 << 30,
		Storage:                   mem,
		Logger:                    lg,
	}

	peers := make([]raft.Peer, len(ids))
	for i, id := range ids {
		peers[i] = raft.Peer{ID: id}
	}

	node := new(node)
	node.mem = mem
	if newCluster {
		// create a new cluster
		node.rn = raft.StartNode(cfg, peers)
	} else {
		// join an existing cluster
		node.rn = raft.RestartNode(cfg)
	}
	node.tr = new(transport)

	return node
}

type node struct {
	rn  raft.Node
	mem *raft.MemoryStorage
	tr  *transport
}

func (n *node) run() {
	ticker := time.NewTicker(time.Millisecond * 100)
	for {
		select {
		case <-ticker.C:
			n.rn.Tick()
		case rd := <-n.rn.Ready():
			n.mem.Append(rd.Entries)

			for _, ent := range rd.CommittedEntries {
				fmt.Println(ent.Type)
				if ent.Type == raftpb.EntryConfChange {
					cc := new(raftpb.ConfChange)
					pbutil.MustUnmarshal(cc, ent.Data)
					n.rn.ApplyConfChange(cc)
				}
			}

			for _, msg := range rd.Messages {
				n.tr.send(msg)
			}

			n.rn.Advance()
		}
	}
}

type transport struct {
	remote *node
}

func (tr transport) send(msg raftpb.Message) {
	if tr.remote != nil {
		tr.remote.rn.Step(context.TODO(), msg)
	}
}
