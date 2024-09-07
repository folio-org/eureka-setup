#!/bin/sh

set -e

/usr/local/bin/ebsco/scripts/init.sh &

vault server --config "/usr/local/bin/ebsco/config/vault-server.json"
