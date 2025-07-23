#!/bin/sh

### vault ###
systemctl daemon-reload
systemctl enable vault
systemctl restart vault
