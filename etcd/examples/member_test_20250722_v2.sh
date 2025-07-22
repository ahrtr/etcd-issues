#!/usr/bin/env bash

# reproduce https://github.com/etcd-io/etcd/issues/20340

set -eEuo pipefail

# Define versions
VERSION="v3.6.2"

export PATH=./${VERSION}:${PATH}

# Define member names
MEMBER_1="etcd-1"
MEMBER_2="etcd-2"
MEMBER_3="etcd-3"

# Cleanup any existing data
echo "- [*] Cleaning up old etcd data..."
rm -rf ./tmp/etcd-*
sleep 2
pkill -f "etcd --name" || true
sleep 2
pkill -f "etcd --name" || true
sleep 2
mkdir -p ./tmp/

############################################################
# Helpers
############################################################

function check_logs_for_panic() {
    echo "- [*] Checking panic..."
    if grep -i panic ./tmp/*.log; then
        echo "=== Panic found in logs!"
        exit 1
    else
        echo "  - No panic found in logs."
    fi


    echo "- [*] Checking failure log..."
    if grep -i 'failed to nodeToMember' ./tmp/*.log; then
        echo "=== 'failed to nodeToMember' found in logs!"
        exit 1
    else
        echo "  - No 'failed to nodeToMember' found in logs."
    fi
}

function healthcheck() {
    check_logs_for_panic

    echo "- [*] Performing healthchecks..."
    echo "  - $MEMBER_1"
    etcdctl --endpoints=http://127.0.0.1:2379 endpoint status
    etcdctl --endpoints=http://127.0.0.1:2379 endpoint health

    echo "  - $MEMBER_2"
    etcdctl --endpoints=http://127.0.0.1:2378 endpoint status
    etcdctl --endpoints=http://127.0.0.1:2378 endpoint health

    echo "  - $MEMBER_3"
    etcdctl --endpoints=http://127.0.0.1:2377 endpoint status
    etcdctl --endpoints=http://127.0.0.1:2377 endpoint health
}

############################################################
# Starting Cluster
############################################################

function phase_prep() {
    # Start first etcd member
    echo "- [*] Starting first etcd member ($MEMBER_1)..."
    nohup etcd --name $MEMBER_1 \
        --data-dir ./tmp/etcd-$MEMBER_1 \
        --initial-advertise-peer-urls http://127.0.0.1:2380 \
        --listen-peer-urls http://127.0.0.1:2380 \
        --advertise-client-urls http://127.0.0.1:2379 \
        --listen-client-urls http://127.0.0.1:2379 \
        --initial-cluster "$MEMBER_1=http://127.0.0.1:2380" \
        --initial-cluster-state new > ./tmp/etcd-$MEMBER_1.log 2>&1 &

    sleep 5

    echo "- [*] add members"
    # Add and promote the first learner
    echo "  - Adding learner ($MEMBER_2)..."
    MEMBER2_ID=$(etcdctl --endpoints=http://127.0.0.1:2379 member add $MEMBER_2 --peer-urls=http://127.0.0.1:2382 --learner | grep 'Member' | awk '{print $2}')
    nohup etcd --name $MEMBER_2 \
        --data-dir ./tmp/etcd-$MEMBER_2 \
        --initial-advertise-peer-urls http://127.0.0.1:2382 \
        --listen-peer-urls http://127.0.0.1:2382 \
        --advertise-client-urls http://127.0.0.1:2378 \
        --listen-client-urls http://127.0.0.1:2378 \
        --initial-cluster "$MEMBER_1=http://127.0.0.1:2380,$MEMBER_2=http://127.0.0.1:2382" \
        --initial-cluster-state existing > ./tmp/etcd-$MEMBER_2.log 2>&1 &

    sleep 5
    echo "  - Promoting learner ($MEMBER_2) with ID: ${MEMBER2_ID}..."
    etcdctl --endpoints=http://127.0.0.1:2379 member promote "${MEMBER2_ID}"

    # Add and promote the second learner
    sleep 5
    echo "  - Adding learner ($MEMBER_3)..."
    MEMBER3_ID=$(etcdctl --endpoints=http://127.0.0.1:2379 member add $MEMBER_3 --peer-urls=http://127.0.0.1:2384 --learner | grep 'Member' | awk '{print $2}')
    nohup etcd --name $MEMBER_3 \
        --data-dir ./tmp/etcd-$MEMBER_3 \
        --initial-advertise-peer-urls http://127.0.0.1:2384 \
        --listen-peer-urls http://127.0.0.1:2384 \
        --advertise-client-urls http://127.0.0.1:2377 \
        --listen-client-urls http://127.0.0.1:2377 \
        --initial-cluster "$MEMBER_1=http://127.0.0.1:2380,$MEMBER_2=http://127.0.0.1:2382,$MEMBER_3=http://127.0.0.1:2384" \
        --initial-cluster-state existing > ./tmp/etcd-$MEMBER_3.log 2>&1 &

    sleep 5
    echo "  - Promoting learner ($MEMBER_3) with ID: ${MEMBER3_ID}..."
    etcdctl --endpoints=http://127.0.0.1:2379 member promote "${MEMBER3_ID}"

    # read -p "Continue [y/N]? " -r confirm
    # [[ "${confirm,,}" == "y" ]] || exit 1
}

############################################################
# Removing Members
############################################################

function phase_remove() {
    # Remove etcd-2
    echo "- [*] Removing member ($MEMBER_2)..."
    etcdctl member remove ${MEMBER2_ID}
    sleep 5
    rm -rf ./tmp/etcd-$MEMBER_2

    echo "- [*] Adding member ($MEMBER_2)..."
    echo "  - Adding learner ($MEMBER_2)..."
    MEMBER2_ID=$(etcdctl --endpoints=http://127.0.0.1:2379 member add $MEMBER_2 --peer-urls=http://127.0.0.1:2382 --learner | grep 'Member' | awk '{print $2}')
    nohup etcd --name $MEMBER_2 \
        --data-dir ./tmp/etcd-$MEMBER_2 \
        --initial-advertise-peer-urls http://127.0.0.1:2382 \
        --listen-peer-urls http://127.0.0.1:2382 \
        --advertise-client-urls http://127.0.0.1:2378 \
        --listen-client-urls http://127.0.0.1:2378 \
        --initial-cluster "$MEMBER_1=http://127.0.0.1:2380,$MEMBER_2=http://127.0.0.1:2382,$MEMBER_3=http://127.0.0.1:2384" \
        --initial-cluster-state existing > ./tmp/etcd-$MEMBER_2.log 2>&1 &

    sleep 5
    echo "  - Promoting learner ($MEMBER_2) with ID: ${MEMBER2_ID}..."
    etcdctl --endpoints=http://127.0.0.1:2379 member promote "${MEMBER2_ID}"
}

############################################################
# Cluster reboot
############################################################

function phase_reboot() {
    # Stop all members
    echo "- [*] Stop all members"
    for member in $MEMBER_1 $MEMBER_2 $MEMBER_3; do
        echo "  - Stopping $member..."
        pgrep -f "etcd --name $member" | xargs kill -9
        rm -f ./tmp/etcd-${member}.log
        sleep 2
    done

    # read -p "Continue [y/N]? " -r confirm
    # [[ "${confirm,,}" == "y" ]] || exit 1

    # Start all members again
    echo "- [*] start all members again"

    echo "  - start $MEMBER_1"
    nohup etcd --name $MEMBER_1 \
        --data-dir ./tmp/etcd-$MEMBER_1 \
        --initial-advertise-peer-urls http://127.0.0.1:2380 \
        --listen-peer-urls http://127.0.0.1:2380 \
        --advertise-client-urls http://127.0.0.1:2379 \
        --listen-client-urls http://127.0.0.1:2379 \
        --initial-cluster "$MEMBER_1=http://127.0.0.1:2380,$MEMBER_2=http://127.0.0.1:2382,$MEMBER_3=http://127.0.0.1:2384" \
        --initial-cluster-state new > ./tmp/etcd-$MEMBER_1.log 2>&1 &

    echo "  - start $MEMBER_2"
    nohup etcd --name $MEMBER_2 \
        --data-dir ./tmp/etcd-$MEMBER_2 \
        --initial-advertise-peer-urls http://127.0.0.1:2382 \
        --listen-peer-urls http://127.0.0.1:2382 \
        --advertise-client-urls http://127.0.0.1:2378 \
        --listen-client-urls http://127.0.0.1:2378 \
        --initial-cluster "$MEMBER_1=http://127.0.0.1:2380,$MEMBER_2=http://127.0.0.1:2382,$MEMBER_3=http://127.0.0.1:2384" \
        --initial-cluster-state existing > ./tmp/etcd-$MEMBER_2.log 2>&1 &

    echo "  - start $MEMBER_3"
    nohup etcd --name $MEMBER_3 \
        --data-dir ./tmp/etcd-$MEMBER_3 \
        --initial-advertise-peer-urls http://127.0.0.1:2384 \
        --listen-peer-urls http://127.0.0.1:2384 \
        --advertise-client-urls http://127.0.0.1:2377 \
        --listen-client-urls http://127.0.0.1:2377 \
        --initial-cluster "$MEMBER_1=http://127.0.0.1:2380,$MEMBER_2=http://127.0.0.1:2382,$MEMBER_3=http://127.0.0.1:2384" \
        --initial-cluster-state existing > ./tmp/etcd-$MEMBER_3.log 2>&1 &


    sleep 5
    # read -p "Continue [y/N]? " -r confirm
    # [[ "${confirm,,}" == "y" ]] || exit 1
}

############################################################
# Cleanup
############################################################

function phase_cleanup() {
    # Stop all members
    echo "- [*] Stop all members"
    for member in $MEMBER_1 $MEMBER_2 $MEMBER_3; do
        echo "  - Stopping $member..."
        pgrep -f "etcd --name $member" | xargs kill -9
        sleep 2
    done

    echo "- [*] Done"
}

phase_prep
healthcheck

ATTEMPT=1
#while true; do
    echo "===*** Starting new round === (attempt $ATTEMPT)"

    phase_remove
    healthcheck

    phase_reboot
    healthcheck

    echo "=== Round completed successfully ==="
    ((ATTEMPT++))
#done

phase_cleanup
