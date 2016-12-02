DaemonSet on DaemonSet dependency should be resolved per-node rather then per k8s DaemonSet object.

Rationale for above is that DaemonSets are special objects and if one DaemonSet depends on the other, then the dependency is actually on a per-node basis, not for the whole cluster.

We have several options:

- Don't do anything, treat DaemonSet as other k8s objects. This will take the least amount of time, but it will hinder DaemonSet functionality as part of dependency graph.
- Implement advanced DaemonSet pods orchestration on top of taint/tolerance/node affinity mechanisms as part of AppController pod. This will require hacking DaemonSet node scheduling to make child DaemonSet pods run on only on nodes which have ready parent DaemonSet pods. This can be done inside AppController process. This require to expand our scheduler logic extensively. The idea is to make a proxy for DaemonSet status checker which will retrieve DaemonSet state from k8s storage and manipulate nodes/DS objects to orchestrate them properly.
- We can do above outside of AppController, but we need to make sure that the information won't be scattered over several components.
