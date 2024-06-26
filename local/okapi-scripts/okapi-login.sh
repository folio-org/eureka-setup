#!/bin/bash

set -e

okapi="${TMP_OKAPI_URL:-http://localhost:9130}"
tenant="${TMP_OKAPI_TENANT_ID:-supertenant}"
username="${TMP_OKAPI_USERNAME:-superuser}"
password="${TMP_OKAPI_PASSWORD}"

while [ $# -gt 0 ]; do
  case "$1" in
  --tenant* | -t*)
    if [[ "$1" != *=* ]]; then shift; fi
    tenant="${1#*=}"
    ;;
  --okapi* | -o*)
    if [[ "$1" != *=* ]]; then shift; fi
    okapi="${1#*=}"
    ;;
  --username* | -u*)
    if [[ "$1" != *=* ]]; then shift; fi
    username="${1#*=}"
    ;;
  --password* | -p*)
    if [[ "$1" != *=* ]]; then shift; fi
    password="${1#*=}"
    ;;
  *)
    printf >&2 "Error: Invalid argument: %s\n" "$1"
    exit 3
    ;;
  esac
  shift
done

echo "* username: $username"
echo "* password: $password"
echo "* okapi: $okapi"
echo "* tenant: $tenant"

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
