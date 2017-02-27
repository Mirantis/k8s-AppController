#!/bin/bash

source ../common.sh

$KUBECTL_NAME delete -f deps.yaml

cat timedout-pod.yaml | $KUBECTL_NAME delete -f -
cat pod.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap | $KUBECTL_NAME delete -f -

$KUBECTL_NAME delete -f ../../manifests/appcontroller.yaml
