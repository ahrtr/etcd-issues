Highlights 2023 Q1 - etcd
======
<span style="color: #808080; font-family: Babas; font-size: 1em;">
ahrtr@github <br>
April 5, 2023
</span>

# Table of Contents
- **[Background](#background)**
- **[Robustness test](#robustness-test)**
- **[Two critical data inconsistency issues](#two-critical-data-inconsistency-issues)**
- **[Decouple raft from etcd](#decouple-raft-from-etcd)**
- **[Separate http and gRPC servers](#separate-http-and-grpc-servers)**
- **[Added one more role "Member"](#added-one-more-role-member)**

# Background
This post briefly summarizes the big changes in etcd community in 2023 Q1. Some items were actually completed late last year.

# Robustness test
The [robustness](https://github.com/etcd-io/etcd/tree/main/tests/robustness) test is introduced to verify correctness and consistency of data. 
Previously we had a functional test, but it's over designed and wasn't well maintained. It's also flaky. So we replaced the functional test 
with the new robustness test, and it reuses the existing e2e test framework.

# Two critical data inconsistency issues
The two critical issues were officially announced [here](https://docs.google.com/document/d/1q6PausGMsj-ZyqN2Zx0W8426KsB5GEh3XA801JxBCiE/edit#heading=h.vfialzpu94tr).
The first data inconsistency issue may happen when etcd crashes during processing defragmentation operation. When the etcd instance starts again,
it might reapply some entries which have already been applied. Accordingly it might result in the member's data becoming inconsistent
with the other members.

The second issue is for a case when auth is enabled and a new member added to the cluster. The new added member might fail to appy 
due to permission denied, and eventually become data inconsistent with other members.

# Decouple raft from etcd
Previously raft is a module included in the etcd repository. But raft is a standalone protocol to maintain a replicated state machine, 
and is also widely used outside etcd; it may have different evolution pace. So we moved raft into a separate repository [etcd-io/raft](https://github.com/etcd-io/raft)
under the etcd-io organization, and renamed the module name to `go.etcd.io/raft/v3`. Please get more detailed info in [issues/14713](https://github.com/etcd-io/etcd/issues/14713).

# Separate http and gRPC servers
Previously clients access both the gRPC and http services using the same port, e.g. 2379. Now users can configure a separate endpoint/port
to serve only http requests. The reason for this change is when a etcd client is generating high read response load, it can result in 
watch response stream in the same connection being starved. Please get more detailed info in [ssues/15402](https://github.com/etcd-io/etcd/issues/15402).

# Added one more role "Member"
Previously there were only two roles in etcd community, which are reviewer and maintainer. But there is a high bar to be a reviewer, let alone
to be a maintainer. So we decided to add one more role to show early appreciation to active contributors.

Please refer to [pull/15593](https://github.com/etcd-io/etcd/pull/15593).

