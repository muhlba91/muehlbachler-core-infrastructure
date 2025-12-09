#!/bin/sh

### vault ###
systemctl daemon-reload
systemctl enable vault
systemctl restart vault

# finalize installation
echo "installed" > /opt/vault.state

# cleanup old images
sleep 90
docker image prune --all --force || true
