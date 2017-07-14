#!/bin/bash
set -eu

install --compare --owner=root --group=root --mode=200 ./version /etc/deploy/version
install --compare --owner=root --group=root --mode=700 ./manage /usr/local/sbin
