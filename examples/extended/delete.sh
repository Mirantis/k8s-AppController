#!/bin/bash

source ../common.sh

$KUBECTL_NAME delete -f existing_job.yaml

$KUBECTL_NAME delete -f deps.yaml

cat job.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap job1 | $KUBECTL_NAME delete -f -
cat job2.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap job2 | $KUBECTL_NAME delete -f -
cat job3.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap job3 | $KUBECTL_NAME delete -f -
cat job4.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap job4 | $KUBECTL_NAME delete -f -

cat pod.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap pod1 | $KUBECTL_NAME delete -f -
cat pod2.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap pod2 | $KUBECTL_NAME delete -f -
cat pod3.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap pod3 | $KUBECTL_NAME delete -f -
cat pod4.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap pod4 | $KUBECTL_NAME delete -f -
cat pod5.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap pod5 | $KUBECTL_NAME delete -f -
cat pod6.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap pod6 | $KUBECTL_NAME delete -f -
cat pod7.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap pod7 | $KUBECTL_NAME delete -f -
cat pod8.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap pod8 | $KUBECTL_NAME delete -f -
cat pod9.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap pod9 | $KUBECTL_NAME delete -f -

cat replicaset.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap frontend | $KUBECTL_NAME delete -f -

cat daemonset.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap daemonset | $KUBECTL_NAME delete -f -

cat secret.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap secret | $KUBECTL_NAME delete -f -

cat service.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap | $KUBECTL_NAME delete -f -

cat statefulset.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap | $KUBECTL_NAME delete -f -

cat configmap1.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap | $KUBECTL_NAME delete -f -

cat deployment.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap | $KUBECTL_NAME delete -f -

cat pvc.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap | $KUBECTL_NAME delete -f -

$KUBECTL_NAME delete -f ../../manifests/appcontroller.yaml
