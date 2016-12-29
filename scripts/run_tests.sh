#!/bin/bash
K8S_IMAGE="zefciu/k8s-test"

docker run \
    -d --privileged \
    --name "dind-test" \
    --hostname "dind-test" \
    -l kubeadm-dind \
    $K8S_IMAGE


