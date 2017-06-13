# Flows

Flows is a way to designate part of AppController dependency graph (i.e. its subgraph) and then use
it as a vertex in other parts of the graph. With flows it becomes possible to compose complex deployment
topologies from smaller reusable units which may be designed and maintained separately from each other.

Examples, where flows are especially useful:
* Flow may correspond to a single application deployment. Then we can design complex software systems
  as a dependency graph of applications (higher level abstracts) rather than internal parts of those applications
  (pods, services, volumes, etc). Such graphs are not just much easier to understand and visualize, but also
  makes much easier later maintenance of such system.
* Flows may represent units within applications - nodes in a cluster, worker nodes in backend system, frontend
  servers in a web farm. Especially, this is helpful if such units are made of several Kubernetes resources.
  Encapsulating such deployment blocks in a flow allows to build multi-node systems by repeating (replicating)
  flow.
* Moreover, flows can represent an operation that can be performed with application: scale, migrate, heal, backup
  and so on, if such operations can be expressed by Kubernetes resources. Considering that number of then can
  contain arbitrary bash scripts, this is often the case.

Most important properties of flows are:
1. **Name**. Since flow is named part of dependency graph, it has a name. When one wants to use the flow in another
   graph part, he just makes a dependency on the flow vertex.
2. **Scope**. This is a label selector than can be applied to dependencies. All graph vertexes, reachable from the
   flow vertex using only edges (dependencies) that match the selector are said to belong to the flow. There can
   be one or two scopes for a flow: construction scope and (optional) destruction scope. Construction scope defines,
   what is going to be deployed for each flow replica. Destruction scope defines what needs to be performed before
   deleting replica resource.
3. **Parameters**. Flows can have parameters, which can be used to generalize the topology for which flow stands for.
   Resource definition get special syntax for how to substitute parameter values into various string fields of
   resources. Flow consumer can provide parameters values.
4. **Replication**. Strictly saying it is not a property of flows, but rather how that get deployed. Flow replication
   means that there can be several copies (replicas) of the flow subgraph merged together into single graph. Before
   flows it did not make any sense because several copies of the same graph will still produce the same resources in
   Kubernetes. But with parametrization, parameter value may be used as part of resource names. So with different
   arguments the same flow can produce different resources. Replication creates specified number of flow copies, each
   of them gets unique name which can be used like if it was a parameter.

## Flow definition

In AppController dependency graph there are resource definitions and dependencies between them. Flows do not bring
new entities into this picture. Instead, flow is implemented as a yet another resource type that can be used as a graph
vertex. Flow resource is where all flow properties can be specified.

To use flow, i.e. deploy resources that make the flow, one just need to place it in dependency graph, and the flow
vertex creation will trigger creation of the subgraph.

### Flow resource

Below is a sample Flow resource that has all possible attributes:

```YAML
apiVersion: appcontroller.k8s/v1alpha1
kind: Flow
metadata:
  name: flow-name

construction:
  key1: value1
  key2: value2
destruction:
  key1: value1
  key2: value2

replicaSpace: optional-name
exported: true
sequential: true

parameters:
  parameterName1:
    description: optional parameter description
    default: optional parameter default value
  parameterName2:
```

Flow resource must have `Flow` kind and `appcontroller.k8s/v1alpha1` API version. However, despite looking similar
to other Kubernetes resources, Flows are are not real resource. One cannot upload them to Kubernetes as is.
Instead, they must be wrapped inside `Definition` resources. This is consistent with how other graph vertexes
are represented. However, for other resource types it is still possible to create them in Kubernetes with
`kubectl create -f resource.yaml`. But for flows it is always `cat flow.yaml | kubeac wrap | kubectl create -f-`.

Sections below explain each of the flow properties and how they affect graph deployment.

Note: Since it is possible to run `kubeac` both inside and outside of the cluster, here and below I refer to
`kubeac` binary by its simple name. However, in most cases this is going to be something like
`kubectl exec k8s-appcontroller kubeac` to run the binary remotely, in AppController pod.

### Flow name

Each flow must have a `name` in its `metadata`. This name is used to refer to the flow in order to run it.
There are two method to run a flow:
1) Call it from within a dependency graph. This is done by placing dependency on the `flow/flowName` resource,
   where `flowName` is the name of the flow.
2) Explicitly call the flow from command line: `kubeac run flowName`.

The later method is only possible for exported flows. Exportable flows are the flows that have `exported: true`
in their definition. Such flows are explicitly designed to be used by the user, rather than being of an internal use.

When running flow from commandline, flow name may be omitted. In this case `DEFAULT` is used, which is the name of
default flow - the main dependency graph.

### Flow scope

Since flow is a dependency graph on its own, having flow resource alone is not enough. There must be a way to
identify, which resources belong to the flow. This is achieved by traversing the graph starting from the flow
vertex. All vertices that are reachable from this vertex using dependencies with labels that match `construction`
or `destruction` selectors of the flow are said to belong to the flow.

Both `construction` and optional `destruction` fields of the `Flow` resource are dictionaries that state what
labels dependencies must have in order for child resources of those dependencies to be included in the flow.

Flows may consume other flows by making dependencies of the consumed flow resource. Moreover, there might be
other resources in the consuming flow that depends on the flow, being consumed. In this case consumed flow is
going to be a parent in both dependencies that form the flow and those that are part of consuming flow. So it is
important to choose label selectors that do not overlap. If during graph traversal from a flow vertex AppController
encounters another flow vertex, it will not go over dependencies that match the second flow even if they match
the original flow selector. This is especially important for `DEFAULT` flow that has empty selector that matches
any dependency.

`DEFAULT` flow is the one that corresponds to the main dependency graph. It is created implicitly by AppController.
All non-flow resources that do not depend on anything automatically become dependent on `flow/DEFAULT` flow.
`kubeac run` is a shortcut for `kubeac run DEFAULT`. By-default, `DEFAULT` flow is exported, has no arguments and
empty selector for `construction` scope that matches any label. But it can also be explicitly declared with
different settings. Moreover, if there is a dependency that has `flow/DEFAULT` either as a parent or as a child,
the flow must be created explicitly (by creating flow resource definition with the name `DEFAULT`).

For most flows, only the `construction` scope is required, because this is how AppController knows what to deploy
for the flow. `destruction` is only required to specify actions that must be performed before flow resources get
deleted. However, if both `construction` and `destruction` scopes are present the latter will have priority over
the former for dependencies that match both selectors at the same time.

### Flow parameters

In order for flows to be reusable, there should be a way to generalize them. This is exactly what flow parameters
are for. Parameters are arbitrary key-value pairs that can be provided by flow consumers and then used somewhere in
resource definitions.

Only the declared parameter can be used in the flow, unless `kubeac run` is followed by `--undeclared-args` switch.
The advantages of declared parameters are:
* Such declaration serves as a documentation. It is easy to see what parameters can be passed to the flow. Each
  parameter can be accompanied with description text.
* Declaration may have default value for the parameter. If there is a default value, the parameter becomes optional.
  Values for parameter that do not have default must be provided by the flow consumer.

Parameters are declared in the `parameters` section of the `Flow` resource. It is a dictionary where keys are
parameter names and values are structures with two optional fields: `description` and `default`. If none of them
present, then the value may just remain empty, as shown for `parameterName2` in example above. Parameter names
may be any combination of alpha-numeric characters and underscores (`[0-9A-Za-z_]+` regexp).

When creating resource from its `Definition`, AppController looks for `$parameterName` strings in selected fields
of the definition and then substitutes it with parameter value. Each resource has its own list of fields, where
such substitution can take place. Usually it contains all the fields, where parametrization is relevant with
notable exception for fields that contain shell scripts. In order to access parameter values in scripts,
they must be propagated through environment variables, which are parametrized. Parameter references may be used
in any portion of the fields. For example, `a$arg$yet_another_arg-d` is a valid field value that will be turned
into `abc-d` if the flow gets `b` as an `arg` parameter value and `c` for `yet_another_arg`.

Note that in dependency graph resources are identified by their names before substitution takes place. For example,
in dependencies, the pod with name `$arg` will still be `pod/$arg`.

There are two ways to pass arguments to flow:
1. Through the CLI
2. Through dependencies

With the first method it is possible to provide parameter values to the flow being started:\
`kubeac run flowName --arg arg1=value1 --arg arg2=value2`

The second method is used when one flow calls another to pass parameters between them. It can be also used
to pass parameters between two arbitrary resources. In order to do it, `args` field of dependencies is used:
```YAML
apiVersion: appcontroller.k8s/v1alpha1
kind: Dependency
metadata:
  generateName: dependency-
  labels:
    flow: my-flow
parent: pod/my-pod
child: flow/another-flow
args:
  arg1: $myFlowParameter
  arg2: hardcoded value
```

If resource depends on several other resources, each of those dependencies may pass its own arguments. In this
case they all get merged into a single dictionary. If the argument is passed twice with different values, the result
value of this parameter is undefined. However, if this parameter is used in resource name, there is going to be
created as many resources as the number of passed values for the parameter.
For example, the graph
```
             {arg=a}
            ---------
[parent]  /           \   [child]
my-flow ->             -> pod/pod-$arg
          \           /
           ---------
            {arg=b}
```
will create two pods: `pod-a` and `pod-b`.

## Replication

Flow replication is an AppController feature that makes specified number of flow graph copies, each one with
a unique name and then merges them into a single graph. Because each replica name may be used in some of resource
names, there can be resources that belong just to that replica, but there also can be resources that are shared
across replicas.

Replication is very handy where there is a need to deploy several mostly identical entities. For example, it may
be cluster nodes or servers in the web farm. Kubernetes has a built-in replication capabilities for pods. With
AppController flows the entire topology can be replicated.

Replication works as following:
1. User specifies desired number of flow replicas, either in absolute number or relative to the current
   replica count.
2. AppController allocates requested number of replicas. For each replica, special third party Kubernetes resource
   of type `appcontroller.k8s.Replica` is created. The resource name has an auto-generated part which becomes a
   replica name.
3. If new replicas were created by the adjustment:
    1. For each new replica, build flow dependency graph with special **`AC_NAME`** parameter being set to the
       replica name (alongside other flow arguments).
    2. All generated replica graphs are merged into a single graph. Vertices that have the same name are merged
       into one vertex.
4. If new replica count is less than it was before (i.e. there are replicas that should be deleted):
    1. For each extra replica build flow dependency graph using flow `destruction` scope. Pass **`AC_NAME`**
       argument to each replica.
    2. Merge all such replicas into one graph.
    3. After deployment, delete all resources that belong exclusively, to replicas being deleted, including
       the `Replica` object and resources from both `construction` and `destruction` flow scopes.
5. If the number of replicas does not change, build (and merge) dependency graph, for existing replicas.

The replica count is specified using `-n` (or longer `--replicas`) commandline switch:

`kubec run my-flow -n3` creates deploys 3 replicas of `my-flow`. If there were 1, 2 replicas would be created.
If there were 7 of them, 4 replicas would be deleted.\
`kubeac run my-flow -n+1` increments replica count by 1\
`kubeac run my-flow -n-2` decreases replica count by 2\
`kubeac run my-flow` if there are no replicas exist, create one, otherwise validate status of
resources of existing replicas.

### Replication of dependencies

With commandline parameters one can create number of flow replicas. But sometimes there is a need to have flow
that creates several replicas of another flow, or just several resources with the same specification that differ
only in name.

One possible solution is to utilize technique shown above: make parameter value be part of resource name and
then duplicate the dependency that leads to this resource and pass different parameter value along each of
dependencies. This works well for small and fixed number of replicas. But if the number goes big, it becomes hard
to manage such number of dependency objects. Moreover if the number itself is not fixed but rather passed as a
parameter replicating resource by manual replication of dependencies becomes impossible.

Luckily, the dependencies can be automatically replicated. This is done through the `generateFor` field of the
`Dependency` object. `generateFor` is a map where keys are argument names and values are list expressions. Each list
expression is comma-separated list of values. If the value has a form of `number..number`, it is expended into a
list of integers in the given range. For example `"1..3, 10..11, abc"` will turn into `["1", "2", "3", "10", "11", "abc"]`. 
Then the dependency is going to be replicated automatically with each replica getting on of the list values as an
additional argument. There can be several `generateFor` arguments. In this case there is going to be one dependency
for each combination of the list values. For example,

```YAML
apiVersion: appcontroller.k8s/v1alpha1
kind: Dependency
metadata:
  name: dependency
parent: pod/podName
child: flow/flowName-$x-$y
generateFor:
  x: 1..2
  y: a, b
```

has the same effect as

```YAML
apiVersion: appcontroller.k8s/v1alpha1
kind: Dependency
metadata:
  name: dependency1
parent: pod/podName
child: flow/flowName-$x-$y
args:
  x: 1
  y: a
---
apiVersion: appcontroller.k8s/v1alpha1
kind: Dependency
metadata:
  name: dependency2
parent: pod/podName
child: flow/flowName-$x-$y
args:
  x: 2
  y: a
---
apiVersion: appcontroller.k8s/v1alpha1
kind: Dependency
metadata:
  name: dependency3
parent: pod/podName
child: flow/flowName-$x-$y
args:
  x: 1
  y: b
---
apiVersion: appcontroller.k8s/v1alpha1
kind: Dependency
metadata:
  name: dependency4
parent: pod/podName
child: flow/flowName-$x-$y
args:
  x: 2
  y: b
```

Besides simplifying the dependency graph, dependency replication makes possible to have dynamic number of replicas
by using parameter value right inside the list expressions:

```YAML
apiVersion: appcontroller.k8s/v1alpha1
kind: Dependency
metadata:
  name: dependency
parent: pod/podName
child: flow/flowName-$index
generateFor:
  index: 1..$replicaCount
```

### Replica-spaces and contexts

Replica-space, is a tag that all replicas of the flow share. When new `Replica` object for the flow is created,
it gets `replicaspace=name` label. When AppController needs to adjust replica count, first thing it does is selects
all `Replica` objects in the flow replica-space.

Usually, replica-space name is the same as flow name so that replicas from different flows do not interfere.
However, replica-space name of the flow can be specified explicitly using `replicaSpace` attribute of the `Flow`.
If different flows put the same name in the `replicaSpace` fields, they will get shared replica-space. This is
useful for cases, where there are several alternate ways to create entities.

However, there is a case when counting replicas based on their replica-space alone is not enough. Consider there
are two flows: `A` and `B` and flow `A` calls `B` as part of its dependency graph. User deployed 5 replicas of `B`
and then wants to have 2 replicas of `A`. How many replicas of `B` should be created? What should happen if user
opts to delete one `A` replica then?

The problem here is that replicas of flows, created in context of another flow are indistinguishable from replicas,
created explicitly. This is where `contexts` come into play. When flow is run from another flow, it gets unique
name that is a combination of flow name, replica name and the dependency name, which triggered the flow. In other
words, this name is unique for each usage of the flow, but its not random and thus remain the same on subsequent
deployments. This name is called `context`. When AppController looks for replicas of flow that has such context,
it queries Kubernetes for `Replica` objects that have composite label `replicaspace=name;context=context`. When
new replicas need to be created, they get this composite label as well. As a result, each flow occurrence within
another flow will "see" only its own replicas so the `Flow` resource can always adjust replica count to 1.
However, when the flow is run independently, it will not have any context and thus query replicas based on
replica-space alone, which means it will get all the replicas from all contexts.

### Sequential flows

By default, if flow has more than one replica, generated dependency graph would have each replica subgraph attached
to the graph root vertex (the `Flow` vertex). When deployed, resources of all replicas are going to be created in
parallel. However, in some cases it is desired that replicas be deployed sequentially, one by one. This can be achieved
by setting `sequential` attribute of the `Flow` to `true`. For sequential flows each replica roots get attached to the
leaf vertices of previous one.

## Scheduling flow deployments

When user runs `kubeac run something` the deployment does not happen immediately (unless there is also a `--deploy`
switch present in the commandline). Instead it schedules flow deployment that another process will pick and deploy.
This another process is also `kubeac` which is run as `kubeac deploy`. Usually that process is run in a restartable
Kubernetes pod. If the deployment process fails for some reason, the pod will be rescheduled and the deployment
process will be restarted. Then it will pick the same flow and continue deployment. This is what makes AppController
be highly available.

Flow deployments are scheduled by creating a `ConfigMap` object with special label `AppController=FlowDeployment`.
Deployment process watches for config-maps with such label create and delete events, sorts new config-maps in order
they were created and deploys one by one. Thus it is possible to schedule flow deployment using `kubectl` or even
Kubernetes API alone. This is especially useful from within containers where `kubeac` is not available.

Below is a config-map format:
```YAML
apiVersion: v1
kind: ConfigMap
metadata:
  generateName: flow-deployment-          # Any name will work. With generateName it is easier to generate unique names
  labels:
    AppController: FlowDeployment         # Magic label

data:
  selector: ""                            # AppController will only look for definitions and dependencies with
                                          # matching this label selector

  concurrency: "0"                        # How many resources can be deployed in parallel. "0" = no limit
  flowName: ""                            # Flow name. Empty string is the same as DEFAULT
  exportedOnly: "false"                   # When set to "false" any flow can be run.
                                          # Otherwise, only the flows that are marked as exported

  allowUndeclaredArgs: "false"            # Allow flow to use parameters that were not declared
  replicaCount: "0"                       # replica count
  fixedNumberOfReplicas: "false"          # "true" means that `replicaCount` is an absolute number,
                                          # otherwise it is relative to the current count

  minReplicaCount: "0"                    # Minimum number of replicas, "0" = no minimum
  maxReplicaCount: "0"                    # Maximum number of replicas, "0" = no maximum
  allowDeleteExternalResources: "false"   # By default, when replica is deleted, AppController will not delete
                                          # resources that it did not create. This can be overridden by this setting

  arg.arg1: "value1"                      # argument values. "arg." is prepend to each parameter name
  arg.arg2: "value2"
```

All settings in this config-map are optional. The example above is filled with default values.
Once the deployment is done, AppController automatically deletes the config-map for it.
