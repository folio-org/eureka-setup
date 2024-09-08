#!/bin/sh

set -e

/usr/local/bin/folio/scripts/init.sh &

vault server --config "/usr/local/bin/folio/config/vault-server.json"
