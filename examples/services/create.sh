#!/bin/bash
set -x

kubectl.sh create -f ../../manifests/appcontroller.yaml

sleep 15

kubectl.sh create -f deps.yaml

cat job.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -

cat service.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -
cat pod.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -
cat pod2.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -
cat pod3.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -
cat pod4.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -

kubectl.sh exec k8s-appcontroller ac-run

kubectl.sh logs -f k8s-appcontroller
