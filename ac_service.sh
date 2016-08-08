#!/bin/sh

exec 2>&1

source /tmp/envvars

exec /usr/bin/kubeac
