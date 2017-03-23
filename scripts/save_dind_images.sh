#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace


K8S_VERSION=${K8S_VERSION:-v1.5}
CACHE=${CACHE:-"/tmp"}

echo "Saving images to $CACHE"

IMAGE_TAR=$CACHE/$K8S_VERSION.tar
mkdir -p $CACHE
docker save mirantis/kubeadm-dind-cluster:${K8S_VERSION} > $IMAGE_TAR
