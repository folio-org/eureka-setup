#!/bin/bash

set -e

okapi="${TMP_OKAPI_URL:-http://localhost:9130}"
tenant="${TMP_OKAPI_TENANT_ID:-supertenant}"
authToken="${TMP_OKAPI_AUTH_TOKEN}"

while [ $# -gt 0 ]; do
  case "$1" in
  --tenant* | -t*)
    if [[ "$1" != *=* ]]; then shift; fi
    tenant="${1#*=}"
    ;;
  --module* | -m*)
    if [[ "$1" != *=* ]]; then shift; fi
    module="${1#*=}"
    ;;
  --auth-token)
    if [[ "$1" != *=* ]]; then shift; fi
    authToken="${1#*=}"
    ;;
  --okapi* | -o*)
    if [[ "$1" != *=* ]]; then shift; fi
    okapi="${1#*=}"
    ;;
  --version* | -v*)
    if [[ "$1" != *=* ]]; then shift; fi
    moduleVersion="${1#*=}"
    ;;
  *)
    printf >&2 "Error: Invalid argument: %s\n" "$1"
    exit 3
    ;;
  esac
  shift
done

fullModuleName="$module-$moduleVersion"
if [ "$tenant" != 'supertenant' ]; then
  echo "------------------ [$fullModuleName] Creating Tenant ------------------"
  echo
  tenantJson='{ "id": "'$tenant'", "name": "'$tenant'", "description": "Default_tenant" }'
  curl -f -sL -w '\n' -D - -X POST \
    -H "Content-type: application/json" \
    -H "x-okapi-token: $authToken" \
    -d "$tenantJson" \
    "$okapi/_/proxy/tenants"
fi

echo "------------------ [$fullModuleName] Enabling Module for Tenant ------------------"
echo
enableTenantRequest='[ { "id": "'$fullModuleName'", "action": "enable" } ]'
curl -f -m 900 -sL -w '\n' -D - -X POST \
  -H "Content-Type: application/json" \
  -H "x-okapi-token: $authToken" \
  -d "$enableTenantRequest" \
  "$okapi/_/proxy/tenants/$tenant/install"
echo
