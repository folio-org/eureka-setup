#!/bin/sh

set -e

vaultInitFile=/vault/data/initialization.txt

sleep 5

printRootToken() {
  vaultToken=$(jq -r '.root_token' "$vaultInitFile")
  export VAULT_TOKEN="$vaultToken"
  echo "$(date -u +%FT%T.%3NZ) [INFO] init.sh: Root VAULT TOKEN is: $VAULT_TOKEN"
}

if [ ! -f "$vaultInitFile" ]; then
  echo "$(date -u +%FT%T.%3NZ) [INFO] init.sh: Initializing vault..."
  vault operator init -format=json -n 1 -t 1 > ${vaultInitFile}
  printRootToken
  /usr/local/bin/folio/scripts/unseal.sh "$vaultInitFile"
  echo "$(date -u +%FT%T.%3NZ) [INFO] init.sh: Adding predefined secrets..."
  . /usr/local/bin/folio/scripts/add-secrets.sh
else
  printRootToken
  /usr/local/bin/folio/scripts/unseal.sh "$vaultInitFile"
fi


