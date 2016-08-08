#!/bin/bash

set -x

kubectl.sh create -f ../appcontroller.yaml
kubectl.sh create -f ../manifests/dependencies.yaml
kubectl.sh create -f ../manifests/resdefs.yaml

sleep 10

kubectl.sh create -f deps.yaml

cat job.yaml | kubectl.sh exec -i k8s-appcontroller wrap job1 | kubectl.sh create -f -
cat pod.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod1 | kubectl.sh create -f -

kubectl.sh exec k8s-appcontroller ac-run
