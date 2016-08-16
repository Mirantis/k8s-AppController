# Demo
[![asciicast](https://asciinema.org/a/c4ujuq2f8mv1cl16h0u5x0sl1.png)](https://asciinema.org/a/c4ujuq2f8mv1cl16h0u5x0sl1)

# Usage

Clone repo:

`git clone https://github.com/Mirantis/k8s-AppController.git`
`cd k8s-AppController`

Create third party resource kinds:

`kubectl create -f manifests/dependencies.yaml`

`kubectl create -f manifests/resdefs.yaml`

Suppose you have some yaml files with single k8s object definitions (pod and jobs are supported right now). Create AppController ResourceDefintions for them:

`cat path_to_your_pod.yaml | kubectl exec -i k8s-appcontroller wrap <resource_name> | kubectl create -f -`

Create file with dependencies:
```yaml
apiVersion: appcontroller.k8s1/v1alpha1
kind: Dependency
metadata:
  name: dependency-1
parent: pod/<pod_resource_name_1>
child: job/<job_resource_name_2>
---
apiVersion: appcontroller.k8s1/v1alpha1
kind: Dependency
metadata:
  name: dependency-2
parent: pod/<pod_resource_name_2>
child: pod/<pod_resource_name_3>
---
apiVersion: appcontroller.k8s1/v1alpha1
kind: Dependency
metadata:
  name: dependency-3
parent: job/<job_resource_name_1>
child: job/<job_resource_name_1>
```
Load it to k8s:

`kubectl create -f dependencies_file.yaml`

Run AppController pod:

`kubectl create -f appcontroller.yaml`

Start appcontroller process:

`kubectl exec k8s-appcontroller ac-run`

You can stop appcontroller process by:

`kubectl exec k8s-appcontroller ac-stop`
