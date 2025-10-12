package main

import (
	"encoding/json"
	"fmt"
	"os"
)

/*
Example:

{
  "clusterSize": 3,
  "upgradePath": [
    {
      "version": "v3.5.19",
      "path": "/Users/wachao/software/etcd-v3.5.19-darwin-arm64"
    },
    {
      "version": "v3.5.21",
      "path": "/Users/wachao/software/etcd-v3.5.21-darwin-arm64"
    },
    {
      "version": "v3.6.4",
      "path": "/Users/wachao/software/etcd-v3.6.4-darwin-arm64"
    }
  ]
}
*/

type config struct {
	ClusterSize int               `json:"clusterSize"`
	UpgradePath []versionWithPath `json:"upgradePath"`
}

type versionWithPath struct {
	Version string `json:"version"`
	BinPath string `json:"path"`
}

func loadConfig(path string) (*config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	var cfg config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	return &cfg, nil
}
