#!/bin/sh

### frr ###
systemctl daemon-reload
systemctl enable frr
systemctl restart frr
