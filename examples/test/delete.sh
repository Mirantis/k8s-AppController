#!/bin/bash

source ../common.sh

$KUBECTL_NAME delete -f ./pod_def.yaml

$KUBECTL_NAME delete -f ../../manifests/appcontroller.yaml
