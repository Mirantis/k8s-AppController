#!/bin/bash

source ../common.sh

set -x

AC_POD_NAME=${AC_POD_NAME:-k8s-appcontroller}

for file in resdefs/*.yaml
do 
  cat $file | $KUBECTL_NAME exec -i $AC_POD_NAME kubeac wrap | $KUBECTL_NAME create -f-
done

$KUBECTL_NAME create -f dependencies/
$KUBECTL_NAME create -f resources/
