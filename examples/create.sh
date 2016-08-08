#!/bin/bash

set -x

kubectl.sh create -f existing_job.yaml
kubectl.sh create -f ../appcontroller.yaml
kubectl.sh create -f ../manifests/dependencies.yaml
kubectl.sh create -f ../manifests/resdefs.yaml

sleep 10

kubectl.sh create -f deps.yaml

cat job.yaml | kubectl.sh exec -i k8s-appcontroller wrap job1 | kubectl.sh create -f -
cat job2.yaml | kubectl.sh exec -i k8s-appcontroller wrap job2 | kubectl.sh create -f -

cat pod.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod1 | kubectl.sh create -f -
cat pod2.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod2 | kubectl.sh create -f -
cat pod3.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod3 | kubectl.sh create -f -

kubectl.sh exec k8s-appcontroller kubeac $KUBERNETES_CLUSTER_URL
