#!/bin/bash

set -o errexit

IMAGE_REPO=${IMAGE_REPO:-mirantis/k8s-appcontroller}
PUBLISH=${PUBLISH:-}

function push-to-docker {
    if [ -z "$PUBLISH" ]; then
        echo "Publish is disabled"
        exit 0
    fi
    if [ "$TRAVIS_PULL_REQUEST_BRANCH" != "" ]; then
         echo "Processing PR $TRAVIS_PULL_REQUEST_BRANCH"
         exit 0
    fi
    docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"
    if [ ! -z "$TRAVIS_TAG" ]; then
        echo "Pushing with tag - $TRAVIS_TAG"
        docker push $IMAGE_REPO:$TRAVIS_TAG
        exit 0
    fi
    if [ $TRAVIS_BRANCH == "master" ]; then
        echo "Pushing with tag - latest"
        docker push $IMAGE_REPO:latest
        exit 0
    fi
    echo "Pushing with tag - $TRAVIS_BRANCH"
    docker push $IMAGE_REPO:$TRAVIS_BRANCH
}

push-to-docker

