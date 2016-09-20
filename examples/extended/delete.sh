#!/bin/bash


kubectl.sh delete -f existing_job.yaml

kubectl.sh delete -f deps.yaml

cat job.yaml | kubectl.sh exec -i k8s-appcontroller wrap job1 | kubectl.sh delete -f -
cat job2.yaml | kubectl.sh exec -i k8s-appcontroller wrap job2 | kubectl.sh delete -f -
cat job3.yaml | kubectl.sh exec -i k8s-appcontroller wrap job3 | kubectl.sh delete -f -
cat job4.yaml | kubectl.sh exec -i k8s-appcontroller wrap job4 | kubectl.sh delete -f -

cat pod.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod1 | kubectl.sh delete -f -
cat pod2.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod2 | kubectl.sh delete -f -
cat pod3.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod3 | kubectl.sh delete -f -
cat pod4.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod4 | kubectl.sh delete -f -
cat pod5.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod5 | kubectl.sh delete -f -
cat pod6.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod6 | kubectl.sh delete -f -
cat pod7.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod7 | kubectl.sh delete -f -
cat pod8.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod8 | kubectl.sh delete -f -
cat pod9.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod9 | kubectl.sh delete -f -

cat replicaset.yaml | kubectl.sh exec -i k8s-appcontroller wrap frontend | kubectl.sh delete -f -

cat daemonset.yaml | kubectl.sh exec -i k8s-appcontroller wrap daemonset | kubectl.sh delete -f -
kubectl.sh delete -f ../../manifests/appcontroller.yaml
