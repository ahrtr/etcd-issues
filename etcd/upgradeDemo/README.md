UpgradeDemo
======

The upgrade demo follows exactly the same way as how cluster-api upgrades Kubernetes clusters.

Just three steps to run the demo.
- Step 1: configure the config.json, see example below,
```
{
  "size": 3,
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
- Start the demo,
```
./upgradeDemo
```
