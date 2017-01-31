#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

IMAGE_REPO=${IMAGE_REPO:-mirantis/k8s-appcontroller}
IMAGE_TAG=${IMAGE_TAG:-latest}
NUM_NODES=${NUM_NODES:-2}
TMP_IMAGE_PATH=${TMP_IMAGE_PATH:-/tmp/image.tar}
MASTER_NAME=${MASTER_NAME=}
SLAVE_PATTERN=${SLAVE_PATTERN:-"dind_node_"}


function import-image {
    docker save ${IMAGE_REPO}:${IMAGE_TAG} -o "${TMP_IMAGE_PATH}"

    if [ -n "$MASTER_NAME" ]; then
        docker cp "${TMP_IMAGE_PATH}" kube-master:/image.tar
        docker exec -ti "${MASTER_NAME}" docker load -i /image.tar
    fi

    for i in `seq 1 "${NUM_NODES}"`;
    do
        docker cp "${TMP_IMAGE_PATH}" "${SLAVE_PATTERN}$i":/image.tar
        docker exec -ti "${SLAVE_PATTERN}$i" docker load -i /image.tar
    done
    set +o xtrace
    echo "Finished copying docker image to dind nodes"
}

import-image
