#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

CURDIR=$(dirname "${BASH_SOURCE}")


IMAGE_REPO=${IMAGE_REPO:-mirantis/k8s-appcontroller}
IMAGE_TAG=${IMAGE_TAG:-latest}
NUM_NODES=${NUM_NODES:-2}
TMP_IMAGE_PATH=${TMP_IMAGE_PATH:-/tmp/appcontroller.tar}

function import-image {
        echo "Export docker image and import it on a dind node dind_node_1"
        CONTAINERID="$(docker create ${IMAGE_REPO}:${IMAGE_TAG} bash)"
        set -o xtrace
        docker export "${CONTAINERID}" > "${TMP_IMAGE_PATH}"
        # TODO implement it as a provider (e.g source functions)

        for i in `seq 1 "${NUM_NODES}"`;
        do
            docker cp "${TMP_IMAGE_PATH}" dind_node_$i:/tmp
            docker exec -ti dind_node_$i docker import "${TMP_IMAGE_PATH}" ${IMAGE_REPO}:${IMAGE_TAG}
        done
        set +o xtrace
        echo "Finished copying docker image to dind nodes"
}

import-image
