AppController [![Build Status](https://travis-ci.org/Mirantis/k8s-AppController.svg?branch=master)](https://travis-ci.org/Mirantis/k8s-AppController) [![Stories in Progress](https://badge.waffle.io/Mirantis/k8s-AppController.png?label=in%20progress&title=In%20Progress)](http://waffle.io/Mirantis/k8s-AppController)
=============
AppController is a pod that you can spawn in your Kubernetes cluster which will take care of your complex deployments for you.

In ideal world of Kubernetes you don't need to care about dependencies. Unfortunately reality can be different: your applications most like will have some sort of dependencies and your components will require to be deployed in certain order. One of the examples is the database that needs to be installed before all other components can be launched. Best way is to refactor your apps in such a way so they do not fail when database connection is not yet available. For complex applications it can take a lot of time, blocking you from running your workloads on Kubernetes. AppController enables you to use Kubernetes for your complex applications out of hand.

Despite managing complex dependencies and ordered execution, AppController is completely Kubernetes native, using k8s objects and features. This way once you refactor your applications to be full blown micro services, you don't need to change anything. Zero vendor lock-in.

AppController team works closely with other Kubernetes projects to provide wide and seamless integrations.

## Basic concepts

AppController uses three basic concepts:

### K8s Objects

AppController interacts with bare Kubernetes objects by creating them (if they are needed by deployment and do not exist yet) and reading their state. The state is used by AppController to ensure that dependencies for other objects are met.

### Dependencies

Dependencies are objects that represent vertices in your deployment graph. You can define them and easily create them with kubectl. Dependencies are ThirdPartyResource which is API extension provided by AppController. It's worth mentioning, that Dependencies can represent dependency between pre-existing K8s object (not orchestrated by AppController) and Resource Definitions, so parts of your deployment graph can depend on objects that were created in your cluster before you even started AppController-aided-deployment. Dependency could have metadata which can contain additional informations about how to determine if it's fulfilled.

Dependency on Replica Set accepts `success_factor` key with stringified percentage integer value of how many replicas should be ready to fulfill the status check.

### Resource Definitions

Resource Definitions are objects that represent Kubernetes Objects that are not yet created, but are part of deployment graph. They store manifests of underlying objects. AppController supports most of Kubernetes Objects, if some object type is missing please create github issue about it.

Resource Definitions are (the same as Dependencies) ThirdPartyResource API extension.

# Demo


## General AppController demos

[![asciicast](https://asciinema.org/a/c4ujuq2f8mv1cl16h0u5x0sl1.png)](https://asciinema.org/a/c4ujuq2f8mv1cl16h0u5x0sl1)

[Voice demo from sig-apps meeting](https://youtu.be/BXRToNV4Rdw?t=178)

[Voice demo from kubernetes community meeting](https://youtu.be/NzkoocVeFMQ?t=31)

## Helm integration demo

[![asciicast](https://asciinema.org/a/0pzyc1tuyk9ohabohnzenvpbe.png)](https://asciinema.org/a/0pzyc1tuyk9ohabohnzenvpbe)

# Usage

Run AppController pod:

`kubectl create -f https://raw.githubusercontent.com/Mirantis/k8s-AppController/master/manifests/appcontroller.yaml`

Suppose you have some yaml files with single k8s object definitions (pod and jobs are supported right now). Create AppController ResourceDefintions for them:

`cat path_to_your_pod.yaml | kubectl exec -i k8s-appcontroller kubeac wrap | kubectl create -f -`

Create file with dependencies:
```yaml
apiVersion: appcontroller.k8s/v1alpha1
kind: Dependency
metadata:
  name: dependency-1
parent: pod/<pod_resource_name_1>
child: job/<job_resource_name_2>
---
apiVersion: appcontroller.k8s/v1alpha1
kind: Dependency
metadata:
  name: dependency-2
parent: pod/<pod_resource_name_2>
child: pod/<pod_resource_name_3>
---
apiVersion: appcontroller.k8s/v1alpha1
kind: Dependency
metadata:
  name: dependency-3
parent: job/<job_resource_name_1>
child: job/<job_resource_name_3>
---
apiVersion: appcontroller.k8s/v1alpha1
kind: Dependency
metadata:
  name: dependency-4
parent: replicaset/<replicaset_resource_name_1>
child: job/<job_resource_name_1>
meta:
  success_factor: "80"
```
Load it to k8s:

`kubectl create -f dependencies_file.yaml`

Start appcontroller process:

`kubectl exec k8s-appcontroller ac-run`

You can stop appcontroller process by:

`kubectl exec k8s-appcontroller ac-stop`

## Reporting

Use the following command:

`kubectl exec k8s-appcontroller kubeac -- get-status`

to get current status of deployment. You might use `-r` to get detailed report
or `-j` to get a JSON representation of status.

## Building

In order to build, issue::

`make TAG=some-docker-tag`

This would vendor the dependencies with glide and build the container with a
given tag.  The default tag is `mirantis/k8s-appcontroller`

# Multiple AppControllers

You can have multiple AppController pods running in your Kubernetes cluster. You can separate your workloads by labeling your Dependencies and Definitions.

Your AppController objects will be retrieved by AppController for processing based on the selector you provide inside pod environment variable `KUBERNETES_AC_LABEL_SELECTOR`. You can pass this variable to pod using Kubernetes environment variable passing mechanism (empty environment variable is already in `manifests/appcontroller.yaml` file for you to fill).

## Example

Example value of this variable could be `app=app1`. AppController pod with this value in `KUBERNETES_AC_LABEL_SELECTOR` variable will work only with Dependencies and Definitions that contain `app: app1` key-value pair in their `metadata.labels` section.


## Special cases
If the selector is empty, the AppController pod will use all Dependencies and Definitions available in cluster.

You can also override this behaviour by using `-l` flag for `kubeac deploy` command available on AppController pod, but this should be only done for testing purposes and is not encouraged in production.



# Roadmap

Here is the brief list of our mid term Roadmap:

* Joining Kubernetes Incubator
* Cooperation with [Helm](https://github.com/kubernetes/helm) project
* [Failure handling](https://github.com/Mirantis/k8s-AppController/blob/master/docs/research/failure-handling.md)
* Implementation of [AppController Mysql Multi Slave research](https://github.com/Mirantis/k8s-AppController/blob/master/docs/research/lcm.md)
* Real life examples
* HA for AppController Pod
* Documentation improvements
