package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"time"

	pb "go.etcd.io/etcd/api/v3/etcdserverpb"
)

const (
	basePeerPort   = 12380
	baseClientPort = 2379
)

var (
	members       []*pb.Member
	removedMember *pb.Member

	snapshotCount *int
)

func main() {
	cfgPath := flag.String("f", "config.json", "path to config file")
	snapshotCount = flag.Int("snapshot-count", 50, "umber of committed transactions to trigger a snapshot to disk")
	flag.Parse()

	cfg, err := loadConfig(*cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log.Printf("Start to test, cluster size: %d", cfg.ClusterSize)

	for i, v := range cfg.UpgradePath {
		if i == 0 {
			if err := createInitialCluster(cfg.ClusterSize, v); err != nil {
				log.Fatalf("failed to create initial cluster: %v", err)
			}
		} else {
			if err := rollingUpgrade(cfg.ClusterSize, i, v); err != nil {
				log.Fatalf("failed during upgrade: %v", err)
			}
		}
		if err := writeData(v.BinPath); err != nil {
			log.Fatalf("Failed to write data: %v", err)
		}
		printSeparator()
	}

	log.Println("All upgrades completed")
	log.Println("Sleeping for 1 hour, press ctrl+c to exit")
	time.Sleep(1 * time.Hour)
}

func createInitialCluster(clusterSize int, v versionWithPath) error {
	log.Printf("Creating initial cluster with version %s, cluster size %d", v.Version, clusterSize)
	for i := 0; i < clusterSize; i++ {
		if i == 0 {
			// Start the very first member
			log.Printf("Starting the very first member with version: %s\n", v.Version)
			if err := startEtcd(v.BinPath, 0, i); err != nil {
				return fmt.Errorf("failed to start very first etcd: %v", err)
			}
			mustSaveMemberList(v.BinPath)
		} else {
			if err := bootNewMember(0, i, v); err != nil {
				return err
			}
			mustSaveMemberList(v.BinPath)
		}
	}
	log.Printf("Successfully created initial cluster with version %s of size %d", v.Version, clusterSize)
	return nil
}

func rollingUpgrade(clusterSize int, round int, v versionWithPath) error {
	ctl := filepath.Join(v.BinPath, "etcdctl")
	members2Remove := append([]*pb.Member{}, members...)
	log.Printf("Upgrading cluster to version %s, round: %d\n", v.Version, round)
	for i := 0; i < clusterSize; i++ {
		newMemberName := fmt.Sprintf("etcd-%d", round*10+i)
		log.Printf("rolling replace member (%s: %x) to %s with version %s", members[0].Name, members[0].ID, newMemberName, v.Version)

		// boot a new member
		if err := bootNewMember(round, i, v); err != nil {
			return fmt.Errorf("failed to boot a new member (%d, %d): %v", round, i, err)
		}

		// remove the old member
		member2Remove := members2Remove[i]
		removedMember = member2Remove
		log.Printf("Removing member (%d, %d) %s: %x", round, i, member2Remove.Name, member2Remove.ID)
		if err := removeMember(ctl, member2Remove.ID); err != nil {
			return fmt.Errorf("failed to remove a member (%d, %d): %v", round, i, err)
		}
		time.Sleep(5 * time.Second)

		mustSaveMemberList(v.BinPath)
	}
	log.Printf("Successfully upgraded cluster to version %s, round: %d\n", v.Version, round)
	return nil
}

func bootNewMember(round int, idx int, v versionWithPath) error {
	globalIdx := round*10 + idx

	name := fmt.Sprintf("etcd-%d", globalIdx)
	peerURL := fmt.Sprintf("http://127.0.0.1:%d", basePeerPort+globalIdx)

	ctl := filepath.Join(v.BinPath, "etcdctl")

	// add learner
	log.Printf("Adding learner (%d, %d) %s:%s\n", round, idx, name, peerURL)
	resp, err := addMemberAsLearner(ctl, name, peerURL)
	if err != nil {
		return fmt.Errorf("failed to add learner: %w", err)
	}

	// start learner
	log.Printf("Starting learner (%d, %d) %s", round, idx, name)
	if err := startEtcd(v.BinPath, round, idx); err != nil {
		return fmt.Errorf("failed to start learner (%d, %d): %v", round, idx, err)
	}
	time.Sleep(5 * time.Second)

	// promote learner
	log.Printf("Promoting learner (%d, %d) %s", round, idx, name)
	if err := promoteLearner(ctl, resp.Member.ID); err != nil {
		return fmt.Errorf("failed to promote learner (%d, %d): %v", round, idx, err)
	}
	log.Printf("member %s promoted", name)
	time.Sleep(5 * time.Second)

	return nil
}

func mustSaveMemberList(binPath string) {
	ctl := filepath.Join(binPath, "etcdctl")

	resp, err := memberList(ctl)
	if err != nil {
		log.Fatalf("Failed to get member list: %v\n", err)
	}
	members = resp.Members
	removedMember = nil
}

func printSeparator() {
	fmt.Println("--------------------------------------------------")
	fmt.Println()
}

func writeData(binPath string) error {
	log.Println("Writing some data")
	ctl := filepath.Join(binPath, "etcdctl")
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key-%04d", i)
		value := fmt.Sprintf("value-%04d", i)
		if err := writeRecord(ctl, key, value); err != nil {
			return fmt.Errorf("write data failed, %d: %w", i, err)
		}
	}
	return nil
}
