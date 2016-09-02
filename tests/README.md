# AppController test cases

Each test case is confined within its directory.

You need a cluster with AppController pod running to test it (or you need to manually create ThirdPartyResource API extensions in your cluster and provide AppController process with your cluster connection data.

To create objects from this test case:

```
kubectl create -f dependencies.yaml
kubectl create -f definitions.yaml
```

After that run AppController process. After deployment is done, verify that order of creation of K8s objects matches order defined in expected_order.yaml file.
