Sanity check on etcd cluster and collect basic info
======
<span style="color: #808080; font-family: Babas; font-size: 1em;">
ahrtr@github <br>
January 5th, 2023
</span>

# Table of Contents

- **[Sanity check](#sanity-check)**
  - [Member list](#member-list)
    - [Not necessary to execute `etcdctl member list`?](#not-necessary-to-execute-etcdctl-member-list)
  - [Endpoint status](#endpoint-status)
  - [Endpoint health](#endpoint-health)
- **[Collect basic info](#collect-basic-info)**

# Sanity check
You can do sanity check on running etcd cluster whenever you want.

## Member list
The first thing is to figure out how many members are in the cluster by executing `etcdctl member list` command. See example below,
```
# etcdctl --cacert /etc/kubernetes/pki/etcd/ca.crt --cert /etc/kubernetes/pki/etcd/server.crt --key /etc/kubernetes/pki/etcd/server.key --endpoints https://master-0.etcd.cfcr.internal:2379 member list -w table
+------------------+---------+--------------------------------------+------------------------------------------+------------------------------------------+------------+
|        ID        | STATUS  |                 NAME                 |                PEER ADDRS                |               CLIENT ADDRS               | IS LEARNER |
+------------------+---------+--------------------------------------+------------------------------------------+------------------------------------------+------------+
| 17f206fd866fdab2 | started | f7edad70-ed1f-4de1-9410-8047f8eb363e | https://master-0.etcd.cfcr.internal:2380 | https://master-0.etcd.cfcr.internal:2379 |      false |
| 604ea1193b383592 | started | 144199a1-9bbe-4835-bf70-cd43d2752a1b | https://master-2.etcd.cfcr.internal:2380 | https://master-2.etcd.cfcr.internal:2379 |      false |
| 9dccb73515ee278f | started | d8a7737a-143e-42ae-adad-dc209f8ad82a | https://master-1.etcd.cfcr.internal:2380 | https://master-1.etcd.cfcr.internal:2379 |      false |
+------------------+---------+--------------------------------------+------------------------------------------+------------------------------------------+------------+
```

### Not necessary to execute `etcdctl member list`?
Some people might think the `etcdctl endpoint status` output can cover almost all the info returned by `etcdctl member list`, so 
no need to execute `etcdctl member list` at all. Usually it's true as long as you don't care about the `PEER ADDRS`. 

But there is a corner case, if any member has multiple values (comma separated) for the argument `--advertise-client-urls`, then the two commands above may display different number of endpoints. 
See an example below. But of course, the 2nd and 3rd endpoints in the `etcdctl endpoint status` output have the same ID, you can still tell the correct number of members.

**It's recommended to always execute `etcdctl member list`** in order to avoid any confusion.

```
$ etcdctl member list -w table
+------------------+---------+--------+------------------------+--------------------------------------------------+------------+
|        ID        | STATUS  |  NAME  |       PEER ADDRS       |                   CLIENT ADDRS                   | IS LEARNER |
+------------------+---------+--------+------------------------+--------------------------------------------------+------------+
| 8211f1d0f64f3269 | started | infra1 | http://127.0.0.1:12380 |                            http://127.0.0.1:2379 |      false |
| 91bc3c398fb3c146 | started | infra2 | http://127.0.0.1:22380 | http://127.0.0.1:22379,http://192.168.2.10:22379 |      false |
| fd422379fda50e48 | started | infra3 | http://127.0.0.1:32380 |                           http://127.0.0.1:32379 |      false |
+------------------+---------+--------+------------------------+--------------------------------------------------+------------+

$ etcdctl  endpoint status -w table --cluster
+---------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
|         ENDPOINT          |        ID        | VERSION | DB SIZE | IS LEADER | IS LEARNER | RAFT TERM | RAFT INDEX | RAFT APPLIED INDEX | ERRORS |
+---------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
|     http://127.0.0.1:2379 | 8211f1d0f64f3269 |   3.5.5 |   25 kB |     false |      false |         2 |          8 |                  8 |        |
|    http://127.0.0.1:22379 | 91bc3c398fb3c146 |   3.5.5 |   25 kB |     false |      false |         2 |          8 |                  8 |        |
| http://192.168.2.10:22379 | 91bc3c398fb3c146 |   3.5.5 |   25 kB |     false |      false |         2 |          8 |                  8 |        |
|    http://127.0.0.1:32379 | fd422379fda50e48 |   3.5.5 |   25 kB |      true |      false |         2 |          8 |                  8 |        |
+---------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
```

## Endpoint status
Get endpoints status by executing command `etcdct endpoint status`. See example below,
```
# etcdctl --cacert /etc/kubernetes/pki/etcd/ca.crt --cert /etc/kubernetes/pki/etcd/server.crt --key /etc/kubernetes/pki/etcd/server.key --endpoints https://master-0.etcd.cfcr.internal:2379 endpoint status --cluster -w table
+------------------------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
|                 ENDPOINT                 |        ID        | VERSION | DB SIZE | IS LEADER | IS LEARNER | RAFT TERM | RAFT INDEX | RAFT APPLIED INDEX | ERRORS |
+------------------------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
| https://master-0.etcd.cfcr.internal:2379 | 17f206fd866fdab2 |   3.5.4 |  5.2 MB |     false |      false |        65 |   23058805 |           23058805 |        |
| https://master-2.etcd.cfcr.internal:2379 | 604ea1193b383592 |   3.5.4 |  5.1 MB |      true |      false |        65 |   23058805 |           23058805 |        |
| https://master-1.etcd.cfcr.internal:2379 | 9dccb73515ee278f |   3.5.4 |  5.1 MB |     false |      false |        65 |   23058805 |           23058805 |        |
+------------------------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
```

Note if you get all endpoints included in the flag `--endpoints`, then no need to use flag `--cluster`,
```
--endpoints https://master-0.etcd.cfcr.internal:2379,https://master-1.etcd.cfcr.internal:2379,https://master-2.etcd.cfcr.internal:2379
```

How to read the above table?
- Is the etcd version out of support? The etcd community fixed some critical data inconsistency issues in the year of 2022, so make sure you are not running one of the problematic versions.
  - If it's 3.5.x, make sure it's >= 3.5.6.
- Do all etcd member have similar DB size? Usually minor difference is accepted, but big difference is unexpected unless you perform defragmentation on only part of the members.
- Is there only one leader? In some extreme scenario, when there is leader changing exactly when you are executing the command, you may see two leaders, but when you execute the command again, you should see only one leader.
- Do all member have exactly the same "RAFT TERM", "RAFT INDEX" and "RAFT APPLIED INDEX" values? If not, they should be very close, and converge to the same values when you execute the command multiple times; otherwise it means data inconsistency.
- Is there any alarm displayed in the "ERRORS" column? If yes, then it means something wrong happened.

If you output the endpoints status in JSON format, please format the output to make it readable. See an example below. Note some tools (including `jq` and some online JSON formatters) may replace the last 3 digits 
of `cluster_id` and `member_id` with 0. Please make sure you are using a correct JSON formatter tool. 
```
# etcdctl --cacert /etc/kubernetes/pki/etcd/ca.crt --cert /etc/kubernetes/pki/etcd/server.crt --key /etc/kubernetes/pki/etcd/server.key --endpoints https://master-0.etcd.cfcr.internal:2379 endpoint status --cluster -w json
[
  {
    "Endpoint": "https://master-0.etcd.cfcr.internal:2379",
    "Status": {
      "header": {
        "cluster_id": 7895810959607866176,
        "member_id": 1725449293188291250,
        "revision": 17109610,
        "raft_term": 65
      },
      "version": "3.5.4",
      "dbSize": 5177344,
      "leader": 6939661205564306834,
      "raftIndex": 23061469,
      "raftTerm": 65,
      "raftAppliedIndex": 23061469,
      "dbSizeInUse": 2269184
    }
  },
  {
    "Endpoint": "https://master-2.etcd.cfcr.internal:2379",
    "Status": {
      "header": {
        "cluster_id": 7895810959607866176,
        "member_id": 6939661205564306834,
        "revision": 17109610,
        "raft_term": 65
      },
      "version": "3.5.4",
      "dbSize": 5144576,
      "leader": 6939661205564306834,
      "raftIndex": 23061469,
      "raftTerm": 65,
      "raftAppliedIndex": 23061469,
      "dbSizeInUse": 2265088
    }
  },
  {
    "Endpoint": "https://master-1.etcd.cfcr.internal:2379",
    "Status": {
      "header": {
        "cluster_id": 7895810959607866176,
        "member_id": 11370664597832738703,
        "revision": 17109610,
        "raft_term": 65
      },
      "version": "3.5.4",
      "dbSize": 5144576,
      "leader": 6939661205564306834,
      "raftIndex": 23061469,
      "raftTerm": 65,
      "raftAppliedIndex": 23061469,
      "dbSizeInUse": 2289664
    }
  }
]
```

Obviously, the JSON output has more details than the table output. Points:
- You can get both the `dbSize` and `dbSizeInUse`, accordingly you can calculate how many free space each member has, and tell whether you should perform defragmentation operation.
- `revision` of each member is also returned, make sure all members have the same value or converge to the same value soon, otherwise it also means data inconsistency.

The JSON output also has a couple of drawbacks against the table output. Points:
- Obviously table output is more human-readable than the JSON output.
- You need to figure out which member is the leader by comparing the `member_id` and `leader` fields.
- The field [IsLearner](https://github.com/etcd-io/etcd/blob/6200b22f79abb6e9c4e6126f72ded616d85546c4/api/etcdserverpb/rpc.pb.go#L4248) isn't displayed by default. It's only displayed when its value is `true`. Note that gRPC doesn't transport fields which have golang zero values at all. This might be confusing some people that why the JSON output doesn't display this field.

**The recommendation is to provide both JSON and TABLE outputs**.

## Endpoint health
The `etcdctl endpoint health` issues a read request to each endpoint, and the duration displayed in the column `TOOK` is the duration the read request took. 
If there is any alarm, it will be displayed in the column `ERROR`.
```
etcdctl --cacert /etc/kubernetes/pki/etcd/ca.crt --cert /etc/kubernetes/pki/etcd/server.crt --key /etc/kubernetes/pki/etcd/server.key --endpoints https://master-0.etcd.cfcr.internal:2379 endpoint health --cluster -w table
+------------------------------------------+--------+-------------+-------+
|                 ENDPOINT                 | HEALTH |    TOOK     | ERROR |
+------------------------------------------+--------+-------------+-------+
| https://master-2.etcd.cfcr.internal:2379 |   true | 26.973035ms |       |
| https://master-1.etcd.cfcr.internal:2379 |   true | 39.959273ms |       |
| https://master-0.etcd.cfcr.internal:2379 |   true | 59.082637ms |       |
+------------------------------------------+--------+-------------+-------+
```

# Collect basic info
No matter what issue you are running into, please always provide at least the following info as possible as you can. 
More info might be requested based on the issue details.
- etcd version.
  - Usually the sanity check output contains this info. But it's important, so list it separately in case you have any trouble executing `etcdctl` command.
- The output of all the sanity checks in above section.
  - If the IP addresses or FQDNs in the endpoint are sensitive to you, feel free to obfuscate them.
- Each etcd member's configuration (the CLI arguments and env variables settings).
  - If the IP addresses or FQDNs in the endpoint are sensitive to you, feel free to obfuscate them.
- All members' complete log.
- Each member's metrics if you run into performance issue. See [How to debug performance issue](how_to_debug_performance_issue.md).
- Steps to reproduce the issue (as minimally and precisely as possible). 
- All member's data directories. 
  - Please do not remove the data and keep them at a safe place, we might guide you to analyze the data later.
