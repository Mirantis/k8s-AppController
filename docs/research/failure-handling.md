================
Failure Handling
================

As currently implemented, AppController works well if all requested K8s objects get created properly.
If any of them failed to create, AppController quits immediately. We need to improve failure handling in 
AppController.

Generally we should allow user to provide desired reaction of AppController on possible failure of creating
Kubernetes object. Possible reactions include:

 1. **Retry**: AC tries to delete failed K8s resource and then re-create it. We should be able to set maximum retries and delay before the next retry.
 2. **Rollback**: AC deletes all resources that were created during current run.
 3. **Abort**: Immediate quit, keep all created resources.
 4. **Ignore current**: AC proceeds with creating resources as if current one was created successfully.
 5. **Ignore children**: AC skips creation of all resources depending on the current one and all cascade dependencies, but proceeds with creating the rest of resources.

Also during discussion in the community several additional features related  to the topic were requested:

 - **Debug**: In case of failure one might want to understand the reason and to fix it. This could imply two possible scenarios:
    - **Fix and restart**: In case of failure AC immediately quits (failure policy **Abort**), then user inspects failed resource and restarts execution.
    - **Pause and resume**: In case of failure AC pauses creation of current dependency graph branch, allows user to inspect failed resource, possibly update resource definition, and then resume execution with retrying, ignoring this or all dependent resources. Unfortunately to implement such feature we need to execute AC interactively, but currently AC executes as a batch.
 - **Per-resource failure handling policy**: User might want to set per-resource failure handling policy as well as general (default) one and per resource type. We should take into account that failure handling policy could not be _per-dependency_.
 - **Timeout**: If resource does not get ready during given timeout, it should be considered as failed. It is not clear, should we set timeout per-resource, per-dependency or both.
 - **Alternative dependency graph branch**: In case of resource creation failure we might want to execute alternative branch/path of dependency graph. This could be achieved via two additional features:
    - **Anti-dependency**: An attribute in dependency's metadata that requires creation of the child only if creation of the parent failed.
    - **Alternative dependency** or **Optional dependency**: To be able to join _direct_ and _alternative_ branches of dependency graph.

Implementation details
----------------------

1. All resources should have _delete_ method.
2. **Abort** should be default failure policy. _Alternatively_ **Rollback** should be default failure policy.
3. It should be possible to set up default failure policy via CLI flags or environment variables.
4. Resource could be in _Ignored_ state.
5. Dependency metadata can include _timeout_ attribute and _inverse_ flag.
6. 3rdPartyResource that contains resource definition can include _failurePolicy_ attribute. 
7. TBD syntax for _alternative dependency_.
8. Corresponding changes should be made to scheduler.
