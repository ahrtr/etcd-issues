package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func startEtcd(binPath string, round, idx int) error {
	globalIdx := round*10 + idx
	etcdPath := filepath.Join(binPath, "etcd")
	name := fmt.Sprintf("etcd-%d", round*10+idx)
	dataDir := fmt.Sprintf("data-%s", name)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("failed to create data dir (%s): %v", dataDir, err)
	}
	peerURL := fmt.Sprintf("http://127.0.0.1:%d", basePeerPort+round*10+idx)
	clientURL := fmt.Sprintf("http://127.0.0.1:%d", baseClientPort+round*10+idx)

	args := []string{
		fmt.Sprintf("--name=%s", name),
		fmt.Sprintf("--data-dir=%s", dataDir),
		fmt.Sprintf("--listen-peer-urls=%s", peerURL),
		fmt.Sprintf("--listen-client-urls=%s", clientURL),
		fmt.Sprintf("--advertise-client-urls=%s", clientURL),
		fmt.Sprintf("--initial-advertise-peer-urls=%s", peerURL),
	}

	args = append(args, fmt.Sprintf("--initial-cluster=%s", initialCluster(globalIdx)))
	if globalIdx == 0 {
		args = append(args, "--initial-cluster-state=new")
	} else {
		args = append(args, "--initial-cluster-state=existing")
	}

	var output bytes.Buffer
	if err := executeCmd(etcdPath, args, &output, false); err != nil {
		return err
	}

	testURL := clientURL + "/version"
	log.Printf("Running sanity test on the new etcd instance (%s: %s)\n", name, testURL)
	for i := 0; i < 10; i++ {
		if resp, err := sanityTest(testURL); err == nil {
			log.Printf("Sanity test on new etcd instance (%s: %s) is successful, response: %s\n", name, testURL, resp)
			return nil
		} else {
			log.Println(err)
		}

		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("sanity test on the new etcd instance (%s: %s) failed", name, testURL)

}

func sanityTest(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to GET %s: %v", url, err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}
	return string(body), nil
}

func addMemberAsLearner(ctlPath string, name string, peerURL string) (*clientv3.MemberAddResponse, error) {
	ep := getOneAliveEndpoint()
	args := []string{"--endpoints=" + ep, "member", "add", name, "--peer-urls", peerURL, "--learner", "-w", "json"}
	var output bytes.Buffer
	if err := executeCmd(ctlPath, args, &output, true); err != nil {
		return nil, fmt.Errorf("add learner (%s) failed: %w, output: %s", name, err, output.String())
	}

	var resp clientv3.MemberAddResponse
	if err := json.Unmarshal(output.Bytes(), &resp); err != nil {
		return nil, fmt.Errorf("unmarshal adding learner (%s) response failed: %w, output: %s", name, err, output.String())
	}
	return &resp, nil
}

func promoteLearner(ctlPath string, id uint64) error {
	memberID := fmt.Sprintf("%x", id)
	ep := getOneAliveEndpoint()
	args := []string{"--endpoints=" + ep, "member", "promote", memberID}
	var output bytes.Buffer
	if err := executeCmd(ctlPath, args, &output, true); err != nil {
		return fmt.Errorf("promote learner (%s) failed: %w, output: %s", memberID, err, output.String())
	}
	return nil
}

func removeMember(ctlPath string, id uint64) error {
	memberID := fmt.Sprintf("%x", id)
	ep := getOneAliveEndpoint()
	args := []string{"--endpoints=" + ep, "member", "remove", memberID}
	var output bytes.Buffer
	if err := executeCmd(ctlPath, args, &output, true); err != nil {
		return fmt.Errorf("remove member (%s) failed: %w", memberID, err)
	}
	return nil
}

func memberList(ctlPath string) (*clientv3.MemberListResponse, error) {
	ep := getOneAliveEndpoint()
	args := []string{"--endpoints=" + ep, "member", "list", "-w", "json"}
	var output bytes.Buffer
	if err := executeCmd(ctlPath, args, &output, true); err != nil {
		return nil, fmt.Errorf("list member failed: %w, output: %s", err, output.String())
	}

	var resp clientv3.MemberListResponse
	if err := json.Unmarshal(output.Bytes(), &resp); err != nil {
		return nil, fmt.Errorf("unmarshal list member response failed: %w, output: %s", err, output.String())
	}
	return &resp, nil
}

func initialCluster(idx int) string {
	name := fmt.Sprintf("etcd-%d", idx)
	peerURL := fmt.Sprintf("http://127.0.0.1:%d", basePeerPort+idx)

	if idx == 0 {
		return fmt.Sprintf("%s=%s", name, peerURL)
	}

	s := fmt.Sprintf("%s=%s", name, peerURL)
	for _, m := range members {
		// assume each member has only one peerURL
		s = fmt.Sprintf("%s,%s=%s", s, m.Name, m.PeerURLs[0])
	}

	return s
}

func writeRecord(ctlPath string, key, value string) error {
	ep := getOneAliveEndpoint()
	args := []string{"--endpoints=" + ep, "put", key, value}
	var output bytes.Buffer
	if err := executeCmdWithoutLog(ctlPath, args, &output, true); err != nil {
		return fmt.Errorf("write key (%s) failed: %w", key, err)
	}
	return nil
}

func executeCmd(binPath string, args []string, output io.Writer, wait bool) error {
	log.Printf("Executing %s, args: %s\n", binPath, args)
	return executeCmdWithoutLog(binPath, args, output, wait)
}

func executeCmdWithoutLog(binPath string, args []string, output io.Writer, wait bool) error {
	cmd := exec.Command(binPath, args...)
	cmd.Stdout = output
	cmd.Stderr = output
	if wait {
		return cmd.Run()
	}
	return cmd.Start()
}

func getOneAliveEndpoint() string {
	l := len(members)
	if l == 0 {
		// we just added the very first member
		return "http://127.0.0.1:2379"
	}
	m := members[0]
	if removedMember != nil && m.Name == removedMember.Name {
		m = members[1]
	}
	return m.ClientURLs[0]
}
