#!/usr/bin/env bash

set -e

# Define versions
# You need to download the binaries beforehand
ETCD_OLD_VERSION="etcd-v3.5.19"
ETCD_NEW_VERSION="etcd-v3.6.0-rc.2"

# Define member names
MEMBER_1="etcd-1"
MEMBER_2="etcd-2"
MEMBER_3="etcd-3"

# Cleanup any existing data
echo "Cleaning up old etcd data..."
rm -rf /tmp/etcd-* 

# Start first etcd member
echo "Starting first etcd member ($MEMBER_1)..."
nohup ${ETCD_OLD_VERSION}/etcd --name $MEMBER_1 \
    --data-dir /tmp/etcd-$MEMBER_1 \
    --initial-advertise-peer-urls http://127.0.0.1:2380 \
    --listen-peer-urls http://127.0.0.1:2380 \
    --advertise-client-urls http://127.0.0.1:2379 \
    --listen-client-urls http://127.0.0.1:2379 \
    --initial-cluster "$MEMBER_1=http://127.0.0.1:2380" \
    --initial-cluster-state new > /tmp/etcd-$MEMBER_1.log 2>&1 &

sleep 5

# Add and promote the first learner
echo "Adding learner ($MEMBER_2)..."
MEMBER2_ID=$(ETCDCTL_API=3 etcdctl --endpoints=http://127.0.0.1:2379 member add $MEMBER_2 --peer-urls=http://127.0.0.1:2382 --learner | grep 'Member' | awk '{print $2}')
nohup ${ETCD_OLD_VERSION}/etcd --name $MEMBER_2 \
    --data-dir /tmp/etcd-$MEMBER_2 \
    --initial-advertise-peer-urls http://127.0.0.1:2382 \
    --listen-peer-urls http://127.0.0.1:2382 \
    --advertise-client-urls http://127.0.0.1:2378 \
    --listen-client-urls http://127.0.0.1:2378 \
    --initial-cluster "$MEMBER_1=http://127.0.0.1:2380,$MEMBER_2=http://127.0.0.1:2382" \
    --initial-cluster-state existing > /tmp/etcd-$MEMBER_2.log 2>&1 &

sleep 5
echo "Promiting learner ($MEMBER_2) with ID: ${MEMBER2_ID}..."
ETCDCTL_API=3 etcdctl --endpoints=http://127.0.0.1:2379 member promote ${MEMBER2_ID}
sleep 5

# Add and promote the second learner
echo "Adding learner ($MEMBER_3)..."
MEMBER3_ID=$(ETCDCTL_API=3 etcdctl --endpoints=http://127.0.0.1:2379 member add $MEMBER_2 --peer-urls=http://127.0.0.1:2384 --learner | grep 'Member' | awk '{print $2}')
nohup ${ETCD_OLD_VERSION}/etcd --name $MEMBER_3 \
    --data-dir /tmp/etcd-$MEMBER_3 \
    --initial-advertise-peer-urls http://127.0.0.1:2384 \
    --listen-peer-urls http://127.0.0.1:2384 \
    --advertise-client-urls http://127.0.0.1:2377 \
    --listen-client-urls http://127.0.0.1:2377 \
    --initial-cluster "$MEMBER_1=http://127.0.0.1:2380,$MEMBER_2=http://127.0.0.1:2382,$MEMBER_3=http://127.0.0.1:2384" \
    --initial-cluster-state existing > /tmp/etcd-$MEMBER_3.log 2>&1 &

sleep 5
echo "Promiting learner ($MEMBER_3) with ID: ${MEMBER3_ID}..."
ETCDCTL_API=3 etcdctl --endpoints=http://127.0.0.1:2379 member promote ${MEMBER3_ID}
sleep 5

# Upgrade members to 3.6.0-rc.2 one by one
echo "Upgrading members to $ETCD_NEW_VERSION one by one..."
for member in $MEMBER_1 $MEMBER_2 $MEMBER_3; do
    echo "Stopping $member..."
    pgrep -f "etcd --name $member" | xargs kill -9
    sleep 2
    
    echo "Starting $member with new version..."
    nohup ${ETCD_NEW_VERSION}/etcd --name $member \
        --data-dir /tmp/etcd-$member \
        --initial-advertise-peer-urls http://127.0.0.1:2380 \
        --listen-peer-urls http://127.0.0.1:2380 \
        --advertise-client-urls http://127.0.0.1:2379 \
        --listen-client-urls http://127.0.0.1:2379 \
        --initial-cluster "$MEMBER_1=http://127.0.0.1:2380,$MEMBER_2=http://127.0.0.1:2382,$MEMBER_3=http://127.0.0.1:2384" \
        --initial-cluster-state existing >> /tmp/etcd-$member.log 2>&1 &
    
    sleep 5
done

echo "Upgrade complete!"
