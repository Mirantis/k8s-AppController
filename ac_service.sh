#!/bin/sh

exec 2>&1

exec env -i /usr/bin/kubeac
