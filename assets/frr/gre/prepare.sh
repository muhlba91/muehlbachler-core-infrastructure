#!/bin/sh

### gre ###
# disable autoconf for ipv6
cat <<EOF > /etc/sysctl.d/99-disable-ipv6-autoconf.conf
net.ipv6.conf.default.autoconf = 0
net.ipv6.conf.default.accept_ra = 0
net.ipv6.conf.all.autoconf = 0
net.ipv6.conf.all.accept_ra = 0
EOF
