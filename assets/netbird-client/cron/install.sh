#!/bin/sh

### cron ###
chmod +x /bin/netbird-client-backup
systemctl daemon-reload
systemctl restart cron
