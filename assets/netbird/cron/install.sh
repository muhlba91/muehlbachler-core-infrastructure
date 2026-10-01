#!/bin/sh

### cron ###
chmod +x /bin/netbird-backup
systemctl daemon-reload
systemctl restart cron
