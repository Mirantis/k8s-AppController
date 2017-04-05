#!/bin/bash

source ../common.sh

$KUBECTL_NAME create -f ../../manifests/appcontroller.yaml
wait-appcontroller

$KUBECTL_NAME create -f deps.yaml

cat pod.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap | $KUBECTL_NAME create -f -
cat pod2.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap | $KUBECTL_NAME create -f -

$KUBECTL_NAME exec k8s-appcontroller run
$KUBECTL_NAME logs -f k8s-appcontroller
