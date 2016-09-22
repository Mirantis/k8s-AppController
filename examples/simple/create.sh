#!/bin/bash


echo "Let's create the job that existed before (not being created by AppController)"

echo "kubectl.sh create -f existing_job.yaml"
kubectl.sh create -f existing_job.yaml

echo "Creating pod with AppController binary. This is going to be our entry point."
echo "kubectl.sh create -f ../../appcontroller.yaml"
kubectl.sh create -f ../../manifests/appcontroller.yaml

echo "Wait for cluster to register new endpoints and run appcontroller container"
sleep 10

echo "Let's create dependencies. Please refer to https://github.com/Mirantis/k8s-AppController/tree/demo/examples/simple/graph.svg to see the graph composition."
echo "kubectl.sh create -f deps.yaml"
kubectl.sh create -f deps.yaml

echo "Let's create resource definitions. We are wrapping existing job and pod definitions in ResourceDefinitions. We are not creating the pods and jobs themselves - yet!"
echo "cat job.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap job1 | kubectl.sh create -f -"
cat job.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -
echo "cat job2.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap job2 | kubectl.sh create -f -"
cat job2.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -

echo "cat pod.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap pod1 | kubectl.sh create -f -"
cat pod.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -
echo "cat pod2.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap pod2 | kubectl.sh create -f -"
cat pod2.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -
echo "cat pod3.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap pod3 | kubectl.sh create -f -"
cat pod3.yaml | kubectl.sh exec -i k8s-appcontroller kubeac wrap | kubectl.sh create -f -

echo "Here we are running appcontroller binary itself. As the log will say, it retrieves dependencies and resource definitions from the k8s cluster and creates underlying objects accordingly."
echo "kubectl.sh exec k8s-appcontroller ac-run"
kubectl.sh exec k8s-appcontroller ac-run
