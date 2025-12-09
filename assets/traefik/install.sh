#!/bin/sh

### traefik ###
systemctl daemon-reload
systemctl enable traefik
systemctl restart traefik

# finalize installation
echo "installed" > /opt/traefik.state

# cleanup old images
sleep 90
docker image prune --all --force || true
