#!/bin/bash
# Checks out K8S and kubeadm-dind-cluster with correct versions and stores the
# Location in a given file

CURRENT=`pwd`
WORKING=`mktemp -d`
K8S_TAG="${K8S_TAG:-v1.5.1}"

if [ "$#" -ne 1 ]; then
    echo "Missing argument"
    exit 1
fi

cd $WORKING
git clone --branch $K8S_TAG --depth 1 --single-branch https://github.com/kubernetes/kubernetes.git
cd kubernetes
git clone --depth 1 --single-branch https://github.com/Mirantis/kubeadm-dind-cluster.git dind
cd $CURRENT
echo $WORKING > $1
