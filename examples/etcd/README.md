etcd cluster application
------------------------

1. Prepare AppController:

    1. Start AppController Pod. Suppose its name is k8s-appcontroller
    2. Upload etcd graph definition resources to k8s by running create.sh
    3. This application uses custom etcd docker image which is the original etcd image plus kubectl binary.
       Use build-image.sh script to build the image `etcd-custom` and then publish it to the docker repository
       used by your k8s environment. 

2. Deploy the cluster:

`kubectl exec k8s-appcontroller kubeac run etcd-bootstrap -n 3 --arg clusterName=my-cluster`

`-n 3` - initial number of etcd nodes (can be any number)

`--arg clusterName=my-cluster` - cluster name. This allow to deploy several independent cluster in one k8s namespace. 
If omitted, `etcd` name is used by default.

3. Scale the cluster:

`kubectl exec k8s-appcontroller kubeac run etcd-scale -n +1 --arg clusterName=my-cluster`

`-n +1` - adds one node to the cluster. Use `-n -1` to scale the cluster down by one node. In this case the last 
added node is going to be deleted. At the moment it is only possible to scale cluster up by one node at a time. 
However, any number of nodes can be removed. Note, that this can also remove nodes created upon initial deployment. 

`--arg clusterName=my-cluster` - the name of the cluster to scale (`etcd` if not specified).
