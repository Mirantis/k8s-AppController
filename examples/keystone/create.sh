#/bin/bash

set -x

#set up cluster requirements - thirdparty resources and wrap pod
kubectl.sh create -f ../../appcontroller-wrap.yaml
kubectl.sh create -f ../../manifests/dependencies.yaml
kubectl.sh create -f ../../manifests/resdefs.yaml

sleep 10

#insert dependencies definitions
kubectl.sh create -f deps.yaml

#mariadb static stuff
kubectl.sh create -f mariadb-service.yaml
kubectl.sh create -f mariadb-pv.yaml
kubectl.sh create -f mariadb-pvc.yaml
kubectl.sh create -f mariadb-configmap.yaml

#dependency graph nodes for mariadb
cat mariadb-bootstrap-job.yaml | kubectl.sh exec -i k8s-appcontroller-wrap wrap mariadbjob1 | kubectl.sh create -f -
cat mariadb-pod.yaml | kubectl.sh exec -i k8s-appcontroller-wrap wrap mariadbpod | kubectl.sh create -f -

#execute dependency graph
kubectl.sh create -f ../../appcontroller.yaml
watch kubectl.sh logs k8s-appcontroller
