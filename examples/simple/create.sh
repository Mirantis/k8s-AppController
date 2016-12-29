#!/bin/bash

source ../common.sh

echo "Let's create the job that existed before (not being created by AppController)"

echo "$KUBECTL_NAME create -f existing_job.yaml"
$KUBECTL_NAME create -f existing_job.yaml

echo "Creating pod with AppController binary. This is going to be our entry point."
echo "$KUBECTL_NAME create -f ../../appcontroller.yaml"
$KUBECTL_NAME create -f ../../manifests/appcontroller.yaml
wait-appcontroller

echo "Let's create dependencies. Please refer to https://github.com/Mirantis/k8s-AppController/tree/demo/examples/simple/graph.svg to see the graph composition."
echo "$KUBECTL_NAME create -f deps.yaml"
$KUBECTL_NAME create -f deps.yaml

echo "Let's create resource definitions. We are wrapping existing job and pod definitions in ResourceDefinitions. We are not creating the pods and jobs themselves - yet!"
echo "cat job.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap job1 | $KUBECTL_NAME create -f -"
cat job.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap | $KUBECTL_NAME create -f -
echo "cat job2.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap job2 | $KUBECTL_NAME create -f -"
cat job2.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap | $KUBECTL_NAME create -f -

echo "cat pod.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap pod1 | $KUBECTL_NAME create -f -"
cat pod.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap | $KUBECTL_NAME create -f -
echo "cat pod2.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap pod2 | $KUBECTL_NAME create -f -"
cat pod2.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap | $KUBECTL_NAME create -f -
echo "cat pod3.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap pod3 | $KUBECTL_NAME create -f -"
cat pod3.yaml | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap | $KUBECTL_NAME create -f -

echo "Here we are running appcontroller binary itself. As the log will say, it retrieves dependencies and resource definitions from the k8s cluster and creates underlying objects accordingly."
echo "$KUBECTL_NAME exec k8s-appcontroller ac-run"
$KUBECTL_NAME exec k8s-appcontroller ac-run
