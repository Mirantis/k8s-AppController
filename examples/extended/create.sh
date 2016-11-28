#!/bin/bash


kubectl.sh create -f existing_job.yaml

kubectl.sh create -f ../../manifests/appcontroller.yaml

#wait for appcontroller pod creation
sleep 20

kubectl.sh create -f deps.yaml

cat job.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -
cat job2.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -
cat job3.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -
cat job4.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -

cat pod.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -
cat pod2.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -
cat pod3.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -
cat pod4.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -
cat pod5.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -
cat pod6.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -
cat pod7.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -
cat pod8.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -
cat pod9.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -

cat replicaset.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -

cat service.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -

cat petset.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -

cat daemonset.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -

cat configmap1.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -

kubectl.sh exec k8s-appcontroller ac-run
kubectl.sh logs -f k8s-appcontroller
