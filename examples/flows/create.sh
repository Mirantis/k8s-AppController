#!/bin/bash

source ../common.sh

set -x

$KUBECTL_NAME create -f ../../manifests/appcontroller.yaml
wait-appcontroller

$KUBECTL_NAME create -f deps.yaml

cat flow.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap | $KUBECTL_NAME create -f -
cat job.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap | $KUBECTL_NAME create -f -
cat job2.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap | $KUBECTL_NAME create -f -
cat pod.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap | $KUBECTL_NAME create -f -
cat pod2.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap | $KUBECTL_NAME create -f -

$KUBECTL_NAME exec k8s-appcontroller kubeac run

$KUBECTL_NAME logs -f k8s-appcontroller
