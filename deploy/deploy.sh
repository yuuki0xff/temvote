#!/bin/bash
set -euv

install --compare --owner=root --group=root --mode=600 ./version /etc/deploy/version
install --compare --owner=root --group=root --mode=700 ./manage /usr/local/sbin
install --compare --owner=root --group=root --mode=600 ./cron.d/deploy /etc/cron.d/deploy
