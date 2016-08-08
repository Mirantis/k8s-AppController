#!/bin/bash

export > /tmp/envvars
exec runsv /etc/sv/ac
