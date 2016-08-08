#!/bin/sh

export > /tmp/envvars
exec runsv /etc/sv/ac
