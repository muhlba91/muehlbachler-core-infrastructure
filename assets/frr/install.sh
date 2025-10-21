#!/bin/sh

### frr ###
systemctl daemon-reload
systemctl enable frr
systemctl restart frr

# cleanup old images
sleep 90
docker image prune --all --force || true
