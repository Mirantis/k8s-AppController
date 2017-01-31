#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

WORKING=${WORKING:-`mktemp -d`}
K8S_TAG=${K8S_TAG:-v1.5.2}
DIND_COMPATIBLE_COMMIT=${DIND_COMPATIBLE_COMMIT:-897ad95a8e0e1fe674ff81533d4198a3cecee41e}
mkdir -p $WORKING

pushd $WORKING &> /dev/null
if [ ! -d kubernetes ]; then
  git clone --branch $K8S_TAG --depth 1 --single-branch https://github.com/kubernetes/kubernetes.git
fi
pushd kubernetes &> /dev/null
if [ ! -d dind ]; then 
  git clone --single-branch https://github.com/sttts/kubernetes-dind-cluster.git dind
fi
go get -u github.com/jteeuwen/go-bindata/go-bindata
pushd dind &> /dev/null
git checkout $DIND_COMPATIBLE_COMMIT
popd &> /dev/null
popd &> /dev/null
popd &> /dev/null

echo $WORKING
