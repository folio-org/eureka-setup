#!/bin/bash

set -e

cd "$(dirname "$0")" || exit

okapi="http://localhost:9130"
adminPassword=""
dbPassword=""

modAuthTokenVersion="2.15.1"
modLoginVersion="7.11.0"
modUserVersion="19.3.1"
modPermissionsVersion="6.5.0"

while [ $# -gt 0 ]; do
  case "$1" in
  --okapi* | -o*)
    if [[ "$1" != *=* ]]; then shift; fi
    okapi="${1#*=}"
    ;;
  --admin-password*)
    if [[ "$1" != *=* ]]; then shift; fi
    adminPassword="${1#*=}"
    ;;
  --db-password*)
    if [[ "$1" != *=* ]]; then shift; fi
    dbPassword="${1#*=}"
    ;;
  *)
    printf >&2 "Error: Invalid argument: %s\n" "$1"
    exit 3
    ;;
  esac
  shift
done

tenant="supertenant"

if [[ -z "$adminPassword" ]]; then
  echo "Okapi superuser password must be set using '--admin-password' option"
  exit 1
fi

if [[ -z "$dbPassword" ]]; then
  echo "Okapi database password must be set using '--db-password' option"
  exit 1
fi

echo "------------------ Input values ------------------"
echo "* okapi: $okapi"
echo "* tenant: $tenant"
echo "* admin username: superuser"
echo "* admin password: $adminPassword"
echo

export TMP_TENANT_ID="$tenant"
export TMP_OKAPI_URL="$okapi"

# register required modules
echo "------------------ Registering service in okapi ------------------"
echo
source ./register-module.sh --module "mod-permissions" --version "$modPermissionsVersion"
source ./register-module.sh --module "mod-users" --version "$modUserVersion"
source ./register-module.sh --module "mod-login" --version "$modLoginVersion"
source ./register-module.sh --module "mod-authtoken" --version "$modAuthTokenVersion"

# verify enabled modules
curl -f -w '\n' -D - "$okapi/_/proxy/modules"

# save environment variables
curl -f -w '\n' -D - -XPOST \
  -H "Content-type: application/json" \
  -d '{"name":"DB_HOST","value":"db"}' \
  "$okapi/_/env"

curl -f -w '\n' -D - -XPOST \
  -H "Content-type: application/json" \
  -d '{"name":"DB_PORT","value":"5432"}' \
  "$okapi/_/env"

curl -f -w '\n' -D - -XPOST \
  -H "Content-type: application/json" \
  -d '{"name":"DB_USERNAME","value":"okapi_rw"}' \
  "$okapi/_/env"

curl -f -w '\n' -D - -XPOST \
  -H "Content-type: application/json" \
  -d '{"name":"DB_PASSWORD","value":"'"$dbPassword"'"}' \
  "$okapi/_/env"

curl -f -w '\n' -D - -XPOST \
  -H "Content-type: application/json" \
  -d '{"name":"DB_DATABASE","value":"okapi"}' \
  "$okapi/_/env"

echo "------------------ Securing okapi with superuser ------------------"
echo

echo "------------------ [mod-permissions] Install user ------------------"
echo

source ./enable-tenant.sh --module "mod-permissions" --version "$modPermissionsVersion"

curl -f -w '\n' -D - -XPOST -H "Content-type: application/json" \
  -H "X-Okapi-Tenant:$tenant" \
  -d '{
    "userId":"99999999-9999-4999-9999-999999999999",
    "permissions":[
        "okapi.all",
        "perms.all"
      ]
    }' \
  "$okapi/perms/users"

echo "------------------ [mod-users] Install user ------------------"
echo
source ./enable-tenant.sh --module "mod-users" --version "$modUserVersion"

curl -f -w '\n' -D - -XPOST \
  -H "Content-type: application/json" \
  -H "X-Okapi-Tenant:$tenant" \
  -d '{
    "id": "99999999-9999-4999-9999-999999999999",
    "username": "superuser",
    "active": "true",
    "personal": {
      "lastName": "Superuser",
      "firstName": "Super"
      }
    }' \
  "$okapi/users"

echo "------------------ [mod-login] Install user ------------------"
echo
source ./enable-tenant.sh --module "mod-login" --version "$modLoginVersion"

enableLoginRequest='{ "username": "superuser", "password": "'$adminPassword'" }'
curl -f -w '\n' -D - -X POST \
  -H "Content-type: application/json" \
  -H "X-Okapi-Tenant:$tenant" \
  -d "$enableLoginRequest" \
  "$okapi/authn/credentials"

echo "------------------ [mod-authtoken] Install module ------------------"
echo

source ./enable-tenant.sh --module "mod-authtoken" --version "$modAuthTokenVersion"

enableLoginRequest='{ "username":"superuser", "password":"'$adminPassword'" }'
curl -f -w '\n' -D /tmp/loginheaders -X POST \
  -H "Content-type: application/json" \
  -H "X-Okapi-Tenant:$tenant" \
  -d "$enableLoginRequest" \
  "$okapi/authn/login"
