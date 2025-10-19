#!/bin/sh

### wireguard ###
systemctl daemon-reload
systemctl enable wireguard
systemctl restart wireguard
