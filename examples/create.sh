#!/bin/bash

set -x

kubectl.sh create -f ../appcontroller-wrap.yaml
kubectl.sh create -f ../manifests/dependencies.yaml
kubectl.sh create -f ../manifests/resdefs.yaml

sleep 10

kubectl.sh create -f deps.yaml

cat job.yaml | kubectl.sh exec -i k8s-appcontroller-wrap wrap job1 | kubectl.sh create -f -
cat pod.yaml | kubectl.sh exec -i k8s-appcontroller-wrap wrap pod1 | kubectl.sh create -f -
