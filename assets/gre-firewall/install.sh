#!/bin/sh

### gre firewall ###
chmod 0755 /opt/gre-firewall/firewall.sh
systemctl daemon-reload
systemctl enable gre-firewall
systemctl restart gre-firewall
