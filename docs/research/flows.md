Flows
=====

Flow is a way to designate a part of AC resource graph (i.e. its subgraph) and then use
it as a atomic in other parts of a graph. Thus the deployment of a bigger system may be
split into smaller reusable components and then compose the overall deployment from these
components.

Major advantage of this approach is that the resources constituting the component are
encapsulated in that component. If, say, one want to have a deployment of a database
and an application, that depends on it, both database and the application become such
components and the main graph has just these two nodes, whereas there are separate
subgraphs for each of them. If then one want to change the database subgraph he is free
to do say without a need to make any changes to the application subgraph since it doesn't
depend on any of the resources constituting the database, but rather on the component as a
whole.

Flows are a means to define such components, their scope within the graph and their
properties.

Flows:
* Have a name
* Can be exported so that the user may explicitly deploy particular flow using 
  AppController CLI
* Can be replicated, i.e. produce more than one deployment of the subgraph
* Each flow replica has a unique name which can be substituted into dependent resource
  names or anywhere in their definition. Thus each flow replica might produce both
  different resources (if the replica name is used as part of its name) and share
  common resources
* Can be parametrized. Parameter values can be used in resource defiitions the same
  way as replica name
* May be stable (idempotent) or not. Stable flows produce the same replica names on
  each run, which result in the same resource graph
* Define order in which resources are deleted as well as some custom cleanup steps

Detailed proposal of the flows might be found here: https://docs.google.com/document/d/1UH9r9X3AaOND_KZO0Qslp6iPkt526PMDpG29B4L1sVQ/edit?usp=sharing

Example application, that provides LCM capabilities by utilizing the flows can be found at https://github.com/istalker2/ac-etcd

Note, that both are work in progress and are not final.
