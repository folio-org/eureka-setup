#!/bin/bash

set -e

cd "$(dirname "$0")" || exit

okapi="http://localhost:9130"
baseMgrsPath=""
descriptorsPath=""
authToken="${TMP_OKAPI_AUTH_TOKEN}"
mgrApplicationBaseUrl="http://localhost:9008"

while [ $# -gt 0 ]; do
  case "$1" in
  --okapi* | -o*)
    if [[ "$1" != *=* ]]; then shift; fi
    okapi="${1#*=}"
    ;;
  --tenant* | -t*)
    if [[ "$1" != *=* ]]; then shift; fi
    tenant="${1#*=}"
    ;;
  --username* | -u*)
    if [[ "$1" != *=* ]]; then shift; fi
    username="${1#*=}"
    ;;
  --password* | -p*)
    if [[ "$1" != *=* ]]; then shift; fi
    password="${1#*=}"
    ;;
  --base-mgrs-folder-path* | -bf*)
    if [[ "$1" != *=* ]]; then shift; fi
    baseMgrsPath="${1#*=}"
    ;;
  --descriptors-path* | -bf*)
    if [[ "$1" != *=* ]]; then shift; fi
    descriptorsPath="${1#*=}"
    ;;
  --mgr-applications-url* | -mau*)
    if [[ "$1" != *=* ]]; then shift; fi
    mgrApplicationBaseUrl="${1#*=}"
    ;;
  *)
    printf >&2 "Error: Invalid argument: %s\n" "$1"
    exit 3
    ;;
  esac
  shift
done

echo "------------------ Input values ------------------"
echo "* okapi: $okapi"
echo "* username: $username"
echo "* password: $password"
echo "* okapi: $okapi"
echo "* tenant: $tenant"
echo "* admin username: superuser"
echo "* base mgr-components path: $baseMgrsPath"
echo "* base descriptors path: $descriptorsPath"
echo

export TMP_TENANT_ID="$tenant"
export TMP_OKAPI_URL="$okapi"

echo "------------------ Performing user login [username: $username] ------------------"
loginRequest='{ "username": "'$username'", "password":"'$password'" }'
authToken=$(curl -fSsL -w '\n' -XPOST \
  -H "content-type: application/json" \
  -H "x-okapi-tenant: $tenant" \
  -d "$loginRequest" \
  "$okapi/authn/login" | jq -r '.okapiToken')
echo

if [[ -z "$authToken" ]]; then
  echo "Auth token is null or empty, exiting"
  exit 1
fi

export TMP_OKAPI_AUTH_TOKEN="$authToken"

tenant="supertenant"

echo "------------------ Registering mgr-components in okapi ------------------"
curl -w '\n' -D - -X POST \
  -H "x-okapi-token:$authToken" \
  -H "Content-type: application/json" \
  -d "@$baseMgrsPath/mgr-tenants/src/main/resources/descriptors/ModuleDescriptor.json" \
  http://localhost:9130/_/proxy/modules

curl -w '\n' -D - -X POST \
  -H "x-okapi-token:$authToken" \
  -H "Content-type: application/json" \
  -d "@$baseMgrsPath/mgr-applications/src/main/resources/descriptors/ModuleDescriptor.json" \
  http://localhost:9130/_/proxy/modules

curl -w '\n' -D - -X POST \
  -H "x-okapi-token:$authToken" \
  -H "Content-type: application/json" \
  -d "@$baseMgrsPath/mgr-tenant-entitlements/src/main/resources/descriptors/ModuleDescriptor.json" \
  http://localhost:9130/_/proxy/modules

# enable modules
curl -w '\n' -D - -X POST \
  -H "x-okapi-token:$authToken" \
  -H "Content-type: application/json" \
  -d'{"id":"mgr-tenants"}' \
  http://localhost:9130/_/proxy/tenants/supertenant/modules

curl -w '\n' -D - -X POST \
  -H "x-okapi-token:$authToken" \
  -H "Content-type: application/json" \
  -d'{"id":"mgr-applications"}' \
  http://localhost:9130/_/proxy/tenants/supertenant/modules

curl -w '\n' -D - -X POST \
  -H "x-okapi-token:$authToken" \
  -H "Content-type: application/json" \
  -d'{"id":"mgr-tenant-entitlements"}' \
  http://localhost:9130/_/proxy/tenants/supertenant/modules
# 
# verify enabled modules
curl -w '\n' -D - "$okapi/_/proxy/modules" \
  -H "x-okapi-token:$authToken"

# save environment variables
curl -w '\n' -D - -XPOST \
  -H "Content-type: application/json" \
  -H "x-okapi-token:$authToken"  \
  -d '{"name":"DB_HOST","value":"db"}' \
  "$okapi/_/env"

curl -w '\n' -D - -XPOST \
  -H "Content-type: application/json" \
  -H "x-okapi-token:$authToken"  \
  -d '{"name":"DB_PORT","value":"5432"}' \
  "$okapi/_/env"

curl -w '\n' -D - -XPOST \
  -H "Content-type: application/json" \
  -H "x-okapi-token:$authToken"  \
  -d '{"name":"DB_USERNAME","value":"okapi_rw"}' \
  "$okapi/_/env"

curl -w '\n' -D - -XPOST \
  -H "Content-type: application/json" \
  -H "x-okapi-token:$authToken"  \
  -d '{"name":"DB_PASSWORD","value":"'"$password"'"}' \
  "$okapi/_/env"

curl -w '\n' -D - -XPOST \
  -H "Content-type: application/json" \
  -H "x-okapi-token:$authToken"  \
  -d '{"name":"DB_DATABASE","value":"okapi"}' \
  "$okapi/_/env"

echo "------------------- Add all applications ----------------------------"
curl -w '\n' -X POST "$mgrApplicationBaseUrl/applications" \
  -H "x-okapi-token:$authToken"  \
  -H 'Content-Type: application/json' \
  -d "@$descriptorsPath/eureka-setup-0.0.1-applications.json" | jq '.'
echo

echo "------------------- Add discovery information for all applications ----------------------------"
curl -w '\n' -X POST "$mgrApplicationBaseUrl/modules/discovery" \
  -H "x-okapi-token:$authToken"  \
  -H 'Content-Type: application/json' \
  -d "@$descriptorsPath/eureka-setup-0.0.1-applications-discovery.json" | jq '.'
echo

