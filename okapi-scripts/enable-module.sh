#!/bin/bash

set -e

cd "$(dirname "$0")" || exit

okapi="http://localhost:9130"
tenant="supertenant"
username="superuser"

while [ $# -gt 0 ]; do
  case "$1" in
  --tenant* | -t*)
    if [[ "$1" != *=* ]]; then shift; fi
    tenant="${1#*=}"
    ;;
  --module-url*)
    if [[ "$1" != *=* ]]; then shift; fi
    moduleUrl="${1#*=}"
    ;;
  --module* | -m*)
    if [[ "$1" != *=* ]]; then shift; fi
    module="${1#*=}"
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
  --login)
    login='t'
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

echo "------------ Input Arguments ----------"
echo "username: $username"
echo "password: $password"
echo "okapi: $okapi"
echo "tenant: $tenant"
echo "module: $module"
echo "module url: $moduleUrl"
echo "module version: $moduleVersion"

export TMP_OKAPI_USERNAME="$username"
export TMP_OKAPI_PASSWORD="$password"
export TMP_OKAPI_URL="$okapi"
export TMP_OKAPI_TENANT_ID="$tenant"
export TMP_OKAPI_MODULE_ID="$module"
export TMP_OKAPI_MODULE_URL="$moduleUrl"
export TMP_OKAPI_MODULE_VERSION="$moduleVersion"

if [ -z "$moduleUrl" ]; then
  moduleUrl="http://$module:8081"
fi

if [[ $login == 't' && -n $username && -n $password ]]; then
  source ./okapi-login.sh
fi

source ./register-module.sh

if [ -n "$tenant" ]; then
  source ./enable-tenant.sh
fi
