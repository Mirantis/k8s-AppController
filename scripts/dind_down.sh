#!/bin/bash

set -o errexit

pushd $WORKING/kubernetes/
./dind/dind-down-cluster.sh
popd
