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


TRAVIS_PULL_REQUEST_BRANCH=${TRAVIS_PULL_REQUEST_BRANCH:-}
TRAVIS_TEST_RESULT=${TRAVIS_TEST_RESULT:-}
TRAVIS_BRANCH=${TRAVIS_BRANCH:-}
IMAGE_REPO=${IMAGE_REPO:-mirantis/k8s-appcontroller}
TRAVIS_TAG=${TRAVIS_TAG:-}
PUBLISH=${PUBLISH:-}


function push-to-docker {
  if [ -z "${PUBLISH}" ]; then
    echo "Publish is disabled."
    exit 0
  fi

  if [ -z "${TRAVIS_TEST_RESULT}" ]; then
    echo "TRAVIS_TEST_RESULT is not set!"
    exit 1
  else
    if [ "${TRAVIS_TEST_RESULT}" -ne 0 ]; then
      echo "Some of the previous steps ended with an errors! The build is broken!"
      exit 1
    fi
  fi

  if [ -n "${TRAVIS_PULL_REQUEST_BRANCH}" ]; then
    echo "Processing PR ${TRAVIS_PULL_REQUEST_BRANCH}"
    exit 0
  else
    set +o xtrace
    docker login -u="${DOCKER_USERNAME}" -p="${DOCKER_PASSWORD}"
    set -o xtrace
  fi

  if [ -n "${TRAVIS_TAG}" ]; then
    echo "Pushing with tag - ${TRAVIS_TAG}"
    docker tag "${IMAGE_REPO}" "${IMAGE_REPO}":"${TRAVIS_TAG}"
    docker push "${IMAGE_REPO}":"${TRAVIS_TAG}"
    exit
  fi

  if [ "${TRAVIS_BRANCH}" == "master" ]; then
    echo "Pushing with tag - latest"
    docker push "${IMAGE_REPO}":latest
    exit
  fi

  echo "Pushing with tag - ${TRAVIS_BRANCH}"
  docker tag "${IMAGE_REPO}" "${IMAGE_REPO}":"${TRAVIS_BRANCH}"
  docker push "${IMAGE_REPO}":"${TRAVIS_BRANCH}"
}

push-to-docker
