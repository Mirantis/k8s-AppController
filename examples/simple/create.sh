#!/bin/bash


echo "Let's create the job that existed before (not being created by AppController)"

echo "kubectl.sh create -f existing_job.yaml"
kubectl.sh create -f existing_job.yaml

echo "Creating pod with AppController binary. This is going to be our entry point."
echo "kubectl.sh create -f ../../appcontroller.yaml"
kubectl.sh create -f ../../appcontroller.yaml
echo "Creating ThirdPartyResources - basically we are creating new k8s API endpoints so that we can store Resource Definitions and Dependencies in Kubernetes cluster."
echo "kubectl.sh create -f ../../manifests/dependencies.yaml"
kubectl.sh create -f ../../manifests/dependencies.yaml
echo "kubectl.sh create -f ../../manifests/resdefs.yaml"
kubectl.sh create -f ../../manifests/resdefs.yaml

echo "Wait for cluster to register new endpoints and run appcontroller container"
sleep 10

echo "Let's create dependencies. Please refer to https://github.com/Mirantis/k8s-AppController/tree/demo/examples/simple/graph.svg to see the graph composition."
echo "kubectl.sh create -f deps.yaml"
kubectl.sh create -f deps.yaml

echo "Let's create resource definitions. We are wrapping existing job and pod definitions in ResourceDefinitions. We are not creating the pods and jobs themselves - yet!"
echo "cat job.yaml | kubectl.sh exec -i k8s-appcontroller wrap job1 | kubectl.sh create -f -"
cat job.yaml | kubectl.sh exec -i k8s-appcontroller wrap job1 | kubectl.sh create -f -
echo "cat job2.yaml | kubectl.sh exec -i k8s-appcontroller wrap job2 | kubectl.sh create -f -"
cat job2.yaml | kubectl.sh exec -i k8s-appcontroller wrap job2 | kubectl.sh create -f -

echo "cat pod.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod1 | kubectl.sh create -f -"
cat pod.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod1 | kubectl.sh create -f -
echo "cat pod2.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod2 | kubectl.sh create -f -"
cat pod2.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod2 | kubectl.sh create -f -
echo "cat pod3.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod3 | kubectl.sh create -f -"
cat pod3.yaml | kubectl.sh exec -i k8s-appcontroller wrap pod3 | kubectl.sh create -f -

echo "Here we are running appcontroller binary itself. As the log will say, it retrieves dependencies and resource definitions from the k8s cluster and creates underlying objects accordingly."
echo "kubectl.sh exec k8s-appcontroller kubeac \$KUBERNETES_CLUSTER_URL"
kubectl.sh exec k8s-appcontroller kubeac $KUBERNETES_CLUSTER_URL
