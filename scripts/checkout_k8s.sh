#!/bin/bash
# Checks out K8S and kubeadm-dind-cluster with correct versions and stores the
## Location in a given file

CURRENT=`pwd`
WORKING=`mktemp -d`
K8S_TAG="${K8S_TAG:-v1.5.1}"

if [ "$#" -ne 1 ]; then
    echo "This script accepts only one argument"
    exit 1
fi

set -e

cd $WORKING
git clone --branch $K8S_TAG --depth 1 --single-branch https://github.com/kubernetes/kubernetes.git
cd kubernetes
git clone --depth 1 --single-branch https://github.com/Mirantis/kubeadm-dind-cluster.git dind
curl https://storage.googleapis.com/kubernetes-release/release/v1.5.1/kubernetes-server-linux-amd64.tar.gz | tar -xz --
mkdir -p _output/dockerized/bin/linux/amd64/
mv kubernetes/server/bin/kubectl _output/dockerized/bin/linux/amd64/
mv kubernetes/server/bin/hyperkube _output/dockerized/bin/linux/amd64/
mv kubernetes/server/bin/kubeadm _output/dockerized/bin/linux/amd64/
mv kubernetes/server/bin/kubelet _output/dockerized/bin/linux/amd64/
ls _output/dockerized/bin/linux/amd64/

cd $CURRENT
echo $WORKING > $1
