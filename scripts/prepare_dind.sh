#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

function ensure-build {
  local name=$1
  if [ ! -f _output/bin/$name ]; then
    make WHAT="cmd/$name"
  fi
}

function prepare-dind {
  pushd $WORKING/kubernetes
  ensure-build hyperkube
  ensure-build kubectl
  ./dind/dind-down-cluster.sh
  ./dind/dind-up-cluster.sh
  popd
}

prepare-dind
