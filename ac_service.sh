#!/bin/sh

exec 2>&1

source /tmp/envvars

exec env -i /usr/bin/kubeac
