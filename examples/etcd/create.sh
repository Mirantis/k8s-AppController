# /bin/sh

#!/bin/bash

source ../common.sh

set -x

for file in resdefs/*.yaml
do 
  cat $file | $KUBECTL_NAME exec -i k8s-appcontroller kubeac wrap | $KUBECTL_NAME create -f-
done

for file in dependencies/*.yaml
do 
  cat $file | $KUBECTL_NAME create -f-
done
