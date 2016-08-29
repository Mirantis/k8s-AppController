#!/bin/bash


kubectl.sh create -f existing_job.yaml

kubectl.sh create -f ../../manifests/appcontroller.yaml

#wait for appcontroller pod creation
sleep 20

kubectl.sh create -f deps.yaml

cat job.yaml | kubectl.sh exec -i k8s-appcontroller wrap job1 | kubectl.sh create -f -
cat job2.yaml | kubectl.sh exec -i k8s-appcontroller wrap job2 | kubectl.sh create -f -
cat job3.yaml | kubectl.sh exec -i k8s-appcontroller wrap job3 | kubectl.sh create -f -
cat job4.yaml | kubectl.sh exec -i k8s-appcontroller wrap job4 | kubectl.sh create -f -

cat pod.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod1 | kubectl.sh create -f -
cat pod2.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod2 | kubectl.sh create -f -
cat pod3.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod3 | kubectl.sh create -f -
cat pod4.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod4 | kubectl.sh create -f -
cat pod5.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod5 | kubectl.sh create -f -
cat pod6.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod6 | kubectl.sh create -f -
cat pod7.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod7 | kubectl.sh create -f -
cat pod8.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod8 | kubectl.sh create -f -
cat pod9.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod9 | kubectl.sh create -f -

cat replicaset.yaml | kubectl.sh exec -i k8s-appcontroller wrap frontend | kubectl.sh create -f -

cat service.yaml | kubectl.sh exec -i k8s-appcontroller wrap service | kubectl.sh create -f -

kubectl.sh exec k8s-appcontroller ac-run
kubectl.sh logs -f k8s-appcontroller
