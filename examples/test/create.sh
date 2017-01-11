#!/bin/bash

source ../common.sh

$KUBECTL_NAME create -f ../../manifests/appcontroller.yaml
wait-appcontroller

$KUBECTL_NAME create -f ./pod_def.yaml

$KUBECTL_NAME exec k8s-appcontroller ac-run
$KUBECTL_NAME logs -f k8s-appcontroller
