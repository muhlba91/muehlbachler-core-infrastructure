#!/bin/sh

### cron ###
chmod +x /bin/tailscale-backup
systemctl daemon-reload
systemctl restart cron
