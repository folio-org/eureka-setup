#!/bin/bash

set -e

okapi="${TMP_OKAPI_URL:-http://localhost:9130}"
module="${TMP_OKAPI_MODULE_ID}"
moduleUrl="${TMP_OKAPI_MODULE_URL}"
moduleVersion="${TMP_OKAPI_MODULE_VERSION:-latest}"
authToken="${TMP_OKAPI_AUTH_TOKEN}"

while [ $# -gt 0 ]; do
  case "$1" in
  --module-url*)
    if [[ "$1" != *=* ]]; then shift; fi
    moduleUrl="${1#*=}"
    ;;
  --module* | -m*)
    if [[ "$1" != *=* ]]; then shift; fi
    module="${1#*=}"
    ;;
  --auth-token*)
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

echo "------------------ [$module] Registering in okapi ------------------"
if [[ -z "$moduleUrl" ]]; then
  echo "Setting default module url = 'http://$module:8081', because it was empty."
  moduleUrl="http://$module:8081"
fi

echo "* okapi: $okapi"
echo "* auth token: $authToken"
echo "* module url: $moduleUrl"
echo "* module version: $moduleVersion"
echo

indexDataModules='http://folio-registry.aws.indexdata.com/_/proxy/modules'
tmpDirectory=".temp-descriptors"
if [ ! -d "$tmpDirectory" ]; then
  mkdir "$tmpDirectory"
fi

if [[ "$moduleVersion" =~ ^[0-9] ]]; then
  # perform logic when we know the version of application
  fullModuleName="$module-$moduleVersion"
  moduleDescriptorFile="./$tmpDirectory/$fullModuleName.json"
  if [[ -f "$moduleDescriptorFile" ]]; then
    echo "Using existing file descriptor: $moduleDescriptorFile"
    echo
  else
    echo "------------------ [$module-$moduleVersion] Receiving Module Descriptor ------------------"
    echo
    curl -sS -w '\n' "$indexDataModules?filter=$fullModuleName&latest=1&full=true" | jq '.[0]' >"$moduleDescriptorFile"
  fi
else
  # it's required to check latest version of application in okapi
  echo "------------------ [$module-latest] Receiving Module Descriptor ------------------"
  echo
  moduleDescriptorFile="./$tmpDirectory/${fullModuleName}.json"
  if [ ! -f "$moduleDescriptorFile" ]; then
    curl -sS -w '\n' "$indexDataModules?filter=$fullModuleName&latest=1&full=true" | jq '.[0]' >"$moduleDescriptorFile"
  fi
  moduleVersion=$(jq -r '.id' "$moduleDescriptorFile")
  fullModuleName="$module-$moduleVersion"
fi

echo "------------------ [$fullModuleName] Pushing Module Descriptor ------------------"
curl -sL -w '\n' -XPOST \
  -H "Content-type: application/json" \
  -H "x-okapi-token:$authToken" \
  -d "@$moduleDescriptorFile" \
  -o /dev/null \
  "$okapi/_/proxy/modules?check=false"
echo

echo "------------------ [$fullModuleName] Pushing Deployment Descriptor ------------------"
deploymentJson='{"srvcId":"'$fullModuleName'","instId":"'$fullModuleName'","url":"'$moduleUrl'"}'
curl -sS -w '\n' -XPOST \
  -H "Content-type: application/json" \
  -H "x-okapi-token:$authToken" \
  -d "$deploymentJson" \
  "$okapi/_/discovery/modules"
