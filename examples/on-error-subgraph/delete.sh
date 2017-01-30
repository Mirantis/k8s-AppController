#!/bin/bash

source ../common.sh

$KUBECTL_NAME delete -f deps.yaml

cat pod.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap pod1 | $KUBECTL_NAME delete -f -
cat pod2.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap pod2 | $KUBECTL_NAME delete -f -

$KUBECTL_NAME delete -f ../../appcontroller.yaml
