#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

IMAGE_REPO=${IMAGE_REPO:-mirantis/k8s-appcontroller}
IMAGE_TAG=${IMAGE_TAG:-latest}
NUM_NODES=${NUM_NODES:-2}
TMP_IMAGE_PATH=${TMP_IMAGE_PATH:-/tmp/image.tar}

function import-image {
        # echo "Export docker image and import it on a dind node dind_node_1"
        # CONTAINERID="$(docker create ${IMAGE_REPO}:${IMAGE_TAG} bash)"
        # set -o xtrace
        # docker export "${CONTAINERID}" > "${TMP_IMAGE_PATH}"
        # # TODO implement it as a provider (e.g source functions)

        docker save ${IMAGE_REPO}:${IMAGE_TAG} -o "${TMP_IMAGE_PATH}"

        docker cp "${TMP_IMAGE_PATH}" kube-master:/image.tar
        docker exec -ti kube-master docker load -i /image.tar

        for i in `seq 1 "${NUM_NODES}"`;
        do
            docker cp "${TMP_IMAGE_PATH}" kube-node-$i:/image.tar
            docker exec -ti kube-node-$i docker load -i /image.tar
        done
        set +o xtrace
        echo "Finished copying docker image to dind nodes"
}

import-image