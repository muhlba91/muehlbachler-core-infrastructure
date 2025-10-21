#!/bin/sh

### cron ###
chmod +x /bin/wireguard-backup
systemctl daemon-reload
systemctl restart cron
