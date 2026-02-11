#!/bin/bash

set -euo pipefail

ACTION="create"
CENTRAL_TENANT="ecs"
CONSORTIUM_ID=""
GATEWAY="http://localhost:8000"

while [[ $# -gt 0 ]]; do
  case $1 in
    -a|--action)
      if [ -z "${2:-}" ]; then
        echo "Error: -a|--action requires a value"
        exit 1
      fi
      ACTION="$2"
      shift 2
      ;;
    -c|--central-tenant)
      if [ -z "${2:-}" ]; then
        echo "Error: -c|--central-tenant requires a value"
        exit 1
      fi
      CENTRAL_TENANT="$2"
      shift 2
      ;;
    --consortium-id)
      if [ -z "${2:-}" ]; then
        echo "Error: --consortium-id requires a value"
        exit 1
      fi
      CONSORTIUM_ID="$2"
      shift 2
      ;;
    -h|--help)
      echo "Usage: $0 [OPTIONS]"
      echo "Options:"
      echo "  -a, --action VALUE          Action to perform: create or delete (default: create)"
      echo "  -c, --central-tenant VALUE  Central tenant name (default: ecs)"
      echo "  --consortium-id VALUE       Consortium ID (required)"
      echo "  -h, --help                  Show this help message"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      echo "Use --help for usage information"
      exit 1
      ;;
  esac
done

if [[ "$ACTION" != "create" && "$ACTION" != "delete" ]]; then
  echo "Error: Action must be 'create' or 'delete'"
  exit 1
fi

if [ -z "$CONSORTIUM_ID" ]; then
  echo "Error: --consortium-id is required"
  exit 1
fi

mkdir -p tmp

echo "Retrieving token from $CENTRAL_TENANT central tenant"
TOKEN=$(eureka-cli -p ecs-migration getKeycloakAccessToken -t $CENTRAL_TENANT)

run_task() {
  local TENANT=$1
  local OUTPUT=$2
  
  if [ "$ACTION" = "create" ]; then
    echo -e "\nCreating users for $TENANT member tenant"
    curl -X POST -sf "$GATEWAY/consortia/$CONSORTIUM_ID/tenants/$TENANT/identity-provider" -o "$OUTPUT" \
      -H "x-okapi-tenant: $CENTRAL_TENANT" \
      -H "x-okapi-token: $TOKEN" \
      -H "Content-Type: application/json" \
      -d '{"createProvider": "true","migrateUsers": "true"}' || { echo "Task for $TENANT failed"; kill 0; }
  else
    echo -e "\nDeleting users for $TENANT member tenant"
    curl -X DELETE -sf "$GATEWAY/consortia/$CONSORTIUM_ID/tenants/$TENANT/identity-provider" -o "$OUTPUT" \
      -H "x-okapi-tenant: $CENTRAL_TENANT" \
      -H "x-okapi-token: $TOKEN" \
      -H "Content-Type: application/json" || { echo "Task for $TENANT failed"; kill 0; }
  fi
}

echo -e "\nPreparing to run tasks"
for i in {1..8}; do
  run_task "university${i}" "tmp/university${i}_${ACTION}_users.json" &
done

wait
echo -e "\nAll tasks finished successfully"
