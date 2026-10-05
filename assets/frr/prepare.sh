#!/bin/sh

### frr ###
# create directories
mkdir -p /opt/frr/config || true

# kernel modules for EVPN (VRF + VXLAN)
apt-get update
DEBIAN_FRONTEND=noninteractive apt-get install --yes --no-install-recommends linux-image-extra-virtual
modprobe vrf
modprobe vxlan
cat <<EOF > /etc/modules-load.d/evpn.conf
vrf
vxlan
EOF

# enable forwarding
cat <<EOF > /etc/sysctl.d/99-enable-forwarding.conf
net.ipv4.conf.all.forwarding = 1
net.ipv6.conf.all.forwarding = 1
EOF

sysctl --system
