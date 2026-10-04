#!/bin/sh

### frr ###
# create directories
mkdir -p /opt/frr/config || true

# enable forwarding
cat <<EOF > /etc/sysctl.d/99-enable-forwarding.conf
net.ipv4.conf.all.forwarding = 1
net.ipv6.conf.all.forwarding = 1
EOF

sysctl --system
