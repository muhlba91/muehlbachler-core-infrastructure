#!/bin/sh

### tailscale ###
systemctl daemon-reload
systemctl enable tailscale
systemctl restart tailscale
