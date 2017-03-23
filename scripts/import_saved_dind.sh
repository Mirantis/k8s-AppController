#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace


K8S_VERSION=${K8S_VERSION:-v1.5}
CACHE=${CACHE:-"/tmp"}

echo "Loading dind images from directory $CACHE"

mkdir -p $CACHE
IMAGE_TAR=$CACHE/$K8S_VERSION.tar
if [ -f $IMAGE_TAR ]; then
  docker load < $IMAGE_TAR
fi
