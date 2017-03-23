#!/bin/bash
# Copyright 2017 Mirantis
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o xtrace
set -o pipefail
set -o errexit
set -o nounset


IMAGE_REPO=${IMAGE_REPO:-mirantis/k8s-appcontroller}
IMAGE_TAG=${IMAGE_TAG:-latest}
NUM_NODES=${NUM_NODES:-2}
TMP_IMAGE_PATH=${TMP_IMAGE_PATH:-/tmp/image.tar}
MASTER_NAME=${MASTER_NAME="kube-master"}
SLAVE_PATTERN=${SLAVE_PATTERN:-"kube-node-"}


function import-image {
  docker save -o "${TMP_IMAGE_PATH}" "${IMAGE_REPO}":"${IMAGE_TAG}"

  if [ -n "${MASTER_NAME}" ]; then
    docker cp "${TMP_IMAGE_PATH}" "${MASTER_NAME}":/image.tar
    docker exec -ti "${MASTER_NAME}" docker load -i /image.tar
  fi

  for node in $(seq 1 "${NUM_NODES}"); do
    docker cp "${TMP_IMAGE_PATH}" "${SLAVE_PATTERN}""${node}":/image.tar
    docker exec -ti "${SLAVE_PATTERN}""${node}" docker load -i /image.tar
  done
  echo "Finished copying docker image to dind nodes"
}

import-image
