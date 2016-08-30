#!/bin/bash
set -x

kubectl.sh create -f ../../manifests/appcontroller.yaml

sleep 15

kubectl.sh create -f deps.yaml

cat job.yaml | kubectl.sh exec -i k8s-appcontroller wrap job | kubectl.sh create -f -

cat service.yaml | kubectl.sh exec -i k8s-appcontroller wrap service | kubectl.sh create -f -
cat pod.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod1 | kubectl.sh create -f -
cat pod2.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod2 | kubectl.sh create -f -
cat pod3.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod3 | kubectl.sh create -f -
cat pod4.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod4 | kubectl.sh create -f -

kubectl.sh exec k8s-appcontroller ac-run

kubectl.sh logs -f k8s-appcontroller
