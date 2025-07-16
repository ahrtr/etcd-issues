#!/usr/bin/env bash

# reproduce https://github.com/etcd-io/etcd/issues/20340

set -e

# Define versions
# You need to download the binaries beforehand
ETCD_OLD_VERSION="etcd-v3.5.21"
ETCD_NEW_VERSION="etcd-v3.6.2"

# Define member names
MEMBER_1="etcd-1"
MEMBER_2="etcd-2"
MEMBER_3="etcd-3"

function add_members() {
    # Add and promote the first learner
    echo "Adding learner ($MEMBER_2)..."
    MEMBER2_ID=$(etcdctl --endpoints=http://127.0.0.1:2379 member add $MEMBER_2 --peer-urls=http://127.0.0.1:2382 --learner | grep 'Member' | awk '{print $2}')
    nohup ${ETCD_NEW_VERSION}/etcd --name $MEMBER_2 \
        --data-dir /tmp/etcd-$MEMBER_2 \
        --initial-advertise-peer-urls http://127.0.0.1:2382 \
        --listen-peer-urls http://127.0.0.1:2382 \
        --advertise-client-urls http://127.0.0.1:2378 \
        --listen-client-urls http://127.0.0.1:2378 \
        --initial-cluster "$MEMBER_1=http://127.0.0.1:2380,$MEMBER_2=http://127.0.0.1:2382" \
        --initial-cluster-state existing > /tmp/etcd-$MEMBER_2.log 2>&1 &

    sleep 5
    echo "Promiting learner ($MEMBER_2) with ID: ${MEMBER2_ID}..."
    etcdctl --endpoints=http://127.0.0.1:2379 member promote ${MEMBER2_ID}
    sleep 5

    # Add and promote the second learner
    echo "Adding learner ($MEMBER_3)..."
    MEMBER3_ID=$(etcdctl --endpoints=http://127.0.0.1:2379 member add $MEMBER_2 --peer-urls=http://127.0.0.1:2384 --learner | grep 'Member' | awk '{print $2}')
    nohup ${ETCD_NEW_VERSION}/etcd --name $MEMBER_3 \
        --data-dir /tmp/etcd-$MEMBER_3 \
        --initial-advertise-peer-urls http://127.0.0.1:2384 \
        --listen-peer-urls http://127.0.0.1:2384 \
        --advertise-client-urls http://127.0.0.1:2377 \
        --listen-client-urls http://127.0.0.1:2377 \
        --initial-cluster "$MEMBER_1=http://127.0.0.1:2380,$MEMBER_2=http://127.0.0.1:2382,$MEMBER_3=http://127.0.0.1:2384" \
        --initial-cluster-state existing > /tmp/etcd-$MEMBER_3.log 2>&1 &

    sleep 5
    echo "Promiting learner ($MEMBER_3) with ID: ${MEMBER3_ID}..."
    etcdctl --endpoints=http://127.0.0.1:2379 member promote ${MEMBER3_ID}
    sleep 5
}



# Cleanup any existing data
echo "Cleaning up old etcd data..."
rm -rf /tmp/etcd-* 

# Start first etcd member
echo "Step 1: Starting first etcd member ($MEMBER_1)..."
nohup ${ETCD_NEW_VERSION}/etcd --name $MEMBER_1 \
    --data-dir /tmp/etcd-$MEMBER_1 \
    --initial-advertise-peer-urls http://127.0.0.1:2380 \
    --listen-peer-urls http://127.0.0.1:2380 \
    --advertise-client-urls http://127.0.0.1:2379 \
    --listen-client-urls http://127.0.0.1:2379 \
    --initial-cluster "$MEMBER_1=http://127.0.0.1:2380" \
    --initial-cluster-state new > /tmp/etcd-$MEMBER_1.log 2>&1 &

sleep 5

echo "Step 2: add members"
add_members

read -p "Continue [y/N]? " -r confirm
[[ "${confirm,,}" == "y" ]] || exit 1

# Create a snapshot
echo "Step 3: Creating a snapshot"
rm -f /tmp/snapshot.db
etcdctl snapshot save /tmp/snapshot.db

# Stop all members and cleanup data
echo "Step 4: Stop all members and cleanup data"
for member in $MEMBER_1 $MEMBER_2 $MEMBER_3; do
    echo "Stopping $member..."
    pgrep -f "etcd --name $member" | xargs kill -9
    sleep 2
done

rm -rf /tmp/etcd-* 

# Restore single-node cluster
echo "Step 5: Restoring single-node cluster"
etcdutl snapshot restore /tmp/snapshot.db --data-dir=/tmp/etcd-etcd-1 --name="etcd-1" --initial-cluster="etcd-1=http://127.0.0.1:2380" --initial-advertise-peer-urls=http://127.0.0.1:2380

echo "Step 6: Starting single-node etcd cluster: ($MEMBER_1)..."
nohup ${ETCD_NEW_VERSION}/etcd --name $MEMBER_1 \
    --data-dir /tmp/etcd-$MEMBER_1 \
    --initial-advertise-peer-urls http://127.0.0.1:2380 \
    --listen-peer-urls http://127.0.0.1:2380 \
    --advertise-client-urls http://127.0.0.1:2379 \
    --listen-client-urls http://127.0.0.1:2379 \
    --initial-cluster "$MEMBER_1=http://127.0.0.1:2380" \
    --initial-cluster-state new > /tmp/etcd-$MEMBER_1.log 2>&1 &

read -p "Continue [y/N]? " -r confirm
[[ "${confirm,,}" == "y" ]] || exit 1

echo "Step 7: add members"
add_members

read -p "Continue [y/N]? " -r confirm
[[ "${confirm,,}" == "y" ]] || exit 1

# Remove etcd-2
echo "Step 8: Removing member ($MEMBER_2)..."
etcdctl member remove ${MEMBER2_ID}
sleep 5
rm -rf /tmp/etcd-$MEMBER_2

# Remove etcd-3
echo "Step 9: Removing member ($MEMBER_3)..."
etcdctl member remove ${MEMBER3_ID}
sleep 5
rm -rf /tmp/etcd-$MEMBER_3

read -p "Continue [y/N]? " -r confirm
[[ "${confirm,,}" == "y" ]] || exit 1

echo "Step 10: add members"
add_members

read -p "Continue [y/N]? " -r confirm
[[ "${confirm,,}" == "y" ]] || exit 1

# Stop all members
echo "Step 11: Stop all members"
for member in $MEMBER_1 $MEMBER_2 $MEMBER_3; do
    echo "Stopping $member..."
    pgrep -f "etcd --name $member" | xargs kill -9
    rm -f /tmp/etcd-${member}.log
    sleep 2
done

read -p "Continue [y/N]? " -r confirm
[[ "${confirm,,}" == "y" ]] || exit 1

# Start all members again
echo "Step 12: start all members again"

echo "start $MEMBER_1"
nohup ${ETCD_NEW_VERSION}/etcd --name $MEMBER_1 \
    --data-dir /tmp/etcd-$MEMBER_1 \
    --initial-advertise-peer-urls http://127.0.0.1:2380 \
    --listen-peer-urls http://127.0.0.1:2380 \
    --advertise-client-urls http://127.0.0.1:2379 \
    --listen-client-urls http://127.0.0.1:2379 \
    --initial-cluster "$MEMBER_1=http://127.0.0.1:2380,$MEMBER_2=http://127.0.0.1:2382,$MEMBER_3=http://127.0.0.1:2384" \
    --initial-cluster-state new > /tmp/etcd-$MEMBER_1.log 2>&1 &

echo "start $MEMBER_2"
nohup ${ETCD_NEW_VERSION}/etcd --name $MEMBER_2 \
    --data-dir /tmp/etcd-$MEMBER_2 \
    --initial-advertise-peer-urls http://127.0.0.1:2382 \
    --listen-peer-urls http://127.0.0.1:2382 \
    --advertise-client-urls http://127.0.0.1:2378 \
    --listen-client-urls http://127.0.0.1:2378 \
    --initial-cluster "$MEMBER_1=http://127.0.0.1:2380,$MEMBER_2=http://127.0.0.1:2382,$MEMBER_3=http://127.0.0.1:2384" \
    --initial-cluster-state existing > /tmp/etcd-$MEMBER_2.log 2>&1 &

echo "start $MEMBER_3"
nohup ${ETCD_NEW_VERSION}/etcd --name $MEMBER_3 \
    --data-dir /tmp/etcd-$MEMBER_3 \
    --initial-advertise-peer-urls http://127.0.0.1:2384 \
    --listen-peer-urls http://127.0.0.1:2384 \
    --advertise-client-urls http://127.0.0.1:2377 \
    --listen-client-urls http://127.0.0.1:2377 \
    --initial-cluster "$MEMBER_1=http://127.0.0.1:2380,$MEMBER_2=http://127.0.0.1:2382,$MEMBER_3=http://127.0.0.1:2384" \
    --initial-cluster-state existing > /tmp/etcd-$MEMBER_3.log 2>&1 &

read -p "Continue [y/N]? " -r confirm
[[ "${confirm,,}" == "y" ]] || exit 1

# Stop all members
echo "Step 13: Stop all members"
for member in $MEMBER_1 $MEMBER_2 $MEMBER_3; do
    echo "Stopping $member..."
    pgrep -f "etcd --name $member" | xargs kill -9
    sleep 2
done

echo "Done"
