# AppController Mysql Multi Slave research

This research is meant to provide a guidance on how to achieve ability of AppController-deployed Mysql Multi Slave cluster to cope with losing the master node by promoting one of slave nodes to master.
This is an example which will allow us to use AC for complicated lcm cases.

MySQL Masters should not be restarted. If the master node crashes, the pod eligible for promotion is transformed into master. Each pod eligible for promotion should have an executable which will perform it's promotion.

AC will need additional process traversing the deployment graph and checking the state of vertices. If the vertex is not ready and has deployed children, we need to take action, which will be defined in an annotation in the pod OR in the AC ResourceDefinition Metadata. We need a DSL for defining these actions, which will allow AC to read specification of these action and run them. This could be the implementation of `failurePolicy` from failure-handling research. This DSL can be used for both LCM and error-handling purposes.

Suggested annotation format:
```
{
  "onFail": [
    {"type": "exec", "cmd": "promote_to_master.sh", "oneOf": "role:slave"}, # promote one of slaves to master
    {"type": "create", "template": "slave pod template"}, # create new slave pod which will replicate new master
    {"type": "rerun", "root": "subgraph-vertice-id"} # redeploy subgraph
  ]
}
```
 - `onFail` is a list of actions which need to be performed by ac.
 - `type` is a type of action which should be performed on detected failure. Two types are proposed, exec and create.
   - `exec` type executes the `cmd` in a pod. We need to specify the target pod on which the `cmd` should be executed. In this case, `oneOf` is K8s selector which will be used to retrieve slaves that are eligible for the promotion, one of which will be selected.
   - `create` type creates new k8s object (or AC resource definition) using the `template`. The template will be either a name of k8s object which template we should use, or plain-text object template. We need to see which is preferrable when we implement this.
   - `rerun` type causes part of the graph that starts in `root` to be redeployed.

The AC run will be as follows:
 - Check if all graph vertices are created (don't check their status, just their existence). If not, go to 2. if Yes, go to 3.
 - Run deployment (this is what happens when you run ac process right now).
 - Run new monitoring process of AC.

Above flow allows ac to remain stateless.
