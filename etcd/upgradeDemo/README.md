UpgradeDemo
======

## How it works
The upgrade demo follows exactly the same way as how cluster-api upgrades Kubernetes clusters.

Just three steps to run the demo.
- Step 1: configure the config.json, see example below,
```
{
  "clusterSize": 3,
  "upgradePath": [
    {
      "version": "v3.5.19",
      "path": "/usr/local/bin/etcd-v3.5.19"
    },
    {
      "version": "v3.5.21",
      "path": "/usr/local/bin/etcd-v3.5.21"
    },
    {
      "version": "v3.6.4",
      "path": "/usr/local/bin/etcd-v3.6.4"
    }
  ]
}

```

- Step 2: download the etcd versions from [etcd/releases](https://github.com/etcd-io/etcd/releases),
  and put them to the directories configured in above `config.json`.
- Step 3: start the demo,
```
./upgradeDemo
```

## History issues

1. Too many learner members

Refer to https://etcd.io/blog/2025/upgrade_from_3.5_to_3.6_issue_followup/

Run command below to reproduce the issue,
```
./upgradeDemo -f config.json --snapshot-count=50
```

2. v3.7.x panicking on bootstrap

refer to https://github.com/etcd-io/etcd/issues/22633

Run command below to reproduce the issue,
```
./upgradeDemo -f config2.json --key-count=10
```
