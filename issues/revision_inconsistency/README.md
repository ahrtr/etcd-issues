revision inconsistency issue
======

## Table of Contents

- **[Background](#background)**
- **[Current status](#current-status)**
- **[How to reproduce this issue](#how-to-reproduce-this-issue)**
- **[Root cause](#root-cause)**
- **[How is the issue resolved](#how-is-the-issue-resolved)**
- **[How to workaround this issue](#how-to-workaround-this-issue)**

## Background
If etcd crashes during processing defragmentation operation, then the member's revision might be inconsistent
with other members.

This is a regression issue introduced in [pull/12855](https://github.com/etcd-io/etcd/pull/12855), and all
the existing 3.5.x releases (including 3.5.0 ~ 3.5.5) are impacted. Note that previous critical issue 
[issues/13766](https://github.com/etcd-io/etcd/issues/13766) was also caused by the same PR (12855).

etcd 3.4 doesn't have this issue.

It should be very hard to reproduce this issue in production environment, because:
1. Usually users rarely execute the defragmentation operation.
2. It is low possibility for etcd to crash during defragmentation operation. 
3. Even when etcd crashes during defragmentation operation, it isn't guaranteed to reproduce this issue. If there is no traffic when performing defragmentation, then it will not run into this issue.

## Current status
I just delivered a PR [pull/14730](https://github.com/etcd-io/etcd/pull/14730) for main branch (3.6.0) and 
will backport it to release-3.5 later. 

The fix will be included in etcd v3.5.6.

It's really interesting and funny the PR number [14730](https://github.com/etcd-io/etcd/pull/14730) is 
very similar to previous important issue [14370](https://github.com/etcd-io/etcd/issues/14370).

## How to reproduce this issue
Run load test on an etcd cluster, and perform defragmentation operation on one member. Kill the member when the 
defragmentation operation is in progress. Afterwards, start the member again, then the member's revision 
might be inconsistent with other members.

Usually the problematic member's revision will be larger than other members, because etcd re-applies some duplicated entries.

You can also reproduce this issue by executing the E2E test case [TestLinearizability](https://github.com/etcd-io/etcd/blob/2f558ca0dbf1217c1cc5b82a1e2aec428bf6a04f/tests/linearizability/linearizability_test.go#L43).

## Root cause
When etcd processes the defragmentation operation, it commits all pending data into boltDB, but not 
including the consistent index, so the persisted data may not match the consistent index. If etcd crashes
for whatever reason during or immediately after the defragmentation operation, when it starts again it will 
replay the WAL entries starting from the latest snapshot, accordingly it may re-apply some entries which might
have already been applied, eventually the revision isn't consistent with other members.

Specifically, when etcd processes defragmentation operation, it calls
[unsafeCommit](https://github.com/etcd-io/etcd/blob/2f558ca0dbf1217c1cc5b82a1e2aec428bf6a04f/server/storage/backend/batch_tx.go#L342), 
which doesn't call the [OnPreCommitUnsafe](https://github.com/etcd-io/etcd/blob/2f558ca0dbf1217c1cc5b82a1e2aec428bf6a04f/server/storage/backend/batch_tx.go#L332-L334),
so the consistent index isn't persisted.

## How is the issue resolved
It's simple, call the [OnPreCommitUnsafe](https://github.com/etcd-io/etcd/blob/2f558ca0dbf1217c1cc5b82a1e2aec428bf6a04f/server/storage/backend/batch_tx.go#L332-L334)
in method `unsafeCommit` instead of `commit`. Please refer to [pull/14730](https://github.com/etcd-io/etcd/pull/14730).

## How to workaround this issue
If you run into this issue, then you need to remove the problematic member and cleanup its local data. 
Afterwards, add the member into the cluster again, then it will sync data from the leader automatically.
