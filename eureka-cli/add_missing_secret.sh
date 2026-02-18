#!/bin/bash

set -euo pipefail

# Script to set Vault secret and reset Keycloak user password
# Usage: ./add_missing_secret.sh [OPTIONS]

# Default values
KEYCLOAK_URL="http://keycloak.eureka:8080"
TENANT=""
USERNAMES=()
VAULT_URL="http://localhost:8200"
VAULT_TOKEN=""

# Display usage
usage() {
  cat << EOF
Usage: $0 [OPTIONS]

Set Vault secret and reset Keycloak user password.

OPTIONS:
  -h, --help                       Display this help message
  -t, --tenant TENANT              Tenant name (e.g., diku) [required]
  -u, --username USERNAME(S)       Username(s) to reset (space-delimited) [required]
  --keycloak-url URL               Keycloak URL (default: http://keycloak.eureka:8080)
  --vault-url URL                  Vault URL (default: http://localhost:8200)

EXAMPLES:
  $0 -t diku -u admin
  $0 -t diku -u "mod-users mod-roles-keycloak mod-scheduler"
EOF
  exit 1
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    -h|--help)
      usage
      ;;
    -t|--tenant)
      TENANT="$2"
      shift 2
      ;;
    -u|--username)
      # Split space-delimited usernames into array
      read -ra USERNAMES <<< "$2"
      shift 2
      ;;
    --keycloak-url)
      KEYCLOAK_URL="$2"
      shift 2
      ;;
    --vault-url)
      VAULT_URL="$2"
      shift 2
      ;;
    *)
      echo "Error: Unknown option: $1"
      usage
      ;;
  esac
done

# Validate required arguments
if [[ -z "$TENANT" ]]; then
  echo "Error: Tenant is required"
  usage
fi

if [[ ${#USERNAMES[@]} -eq 0 ]]; then
  echo "Error: At least one username is required"
  usage
fi

# Check for required commands
for cmd in jq curl; do
  if ! command -v $cmd &> /dev/null; then
    echo "Error: Required command '$cmd' is not installed"
    exit 1
  fi
done

if ! command -v openssl &> /dev/null && [[ ! -r /dev/urandom ]]; then
  echo "Error: Neither 'openssl' nor '/dev/urandom' is available for password generation"
  exit 1
fi

# Get Vault token
echo "Getting Vault root token..."
VAULT_TOKEN=$(eureka-cli getVaultRootToken)

if [[ -z "$VAULT_TOKEN" || "$VAULT_TOKEN" == "null" ]]; then
  echo "Error: Failed to get Vault root token"
  exit 1
fi

echo -e "\nProcessing ${#USERNAMES[@]} user(s) for tenant '$TENANT'..."

# Read existing secrets from Vault once
echo -e "\nReading existing secrets from Vault..."
VAULT_READ_RESPONSE=$(curl -s --header "X-Vault-Token: $VAULT_TOKEN" \
  "$VAULT_URL/v1/secret/data/folio/$TENANT")
echo -e "\nVault read response:\n\n$VAULT_READ_RESPONSE"

EXISTING_SECRETS=$(echo "$VAULT_READ_RESPONSE" | jq -r '.data.data // {}')

if [[ -z "$EXISTING_SECRETS" ]]; then
  echo "Error: Failed to parse existing secrets from Vault"
  exit 1
fi

echo -e "\nExisting secrets:\n\n$EXISTING_SECRETS"

# Get Keycloak access token once
echo -e "\nGetting Keycloak access token..."
ACCESS_TOKEN=$(eureka-cli getKeycloakAccessToken --tokenType master-admin-cli)

if [[ -z "$ACCESS_TOKEN" || "$ACCESS_TOKEN" == "null" ]]; then
  echo "Error: Failed to get Keycloak access token"
  exit 1
fi

# Store results for summary
declare -A USER_PASSWORDS
FAILED_USERS=()
MERGED_SECRETS="$EXISTING_SECRETS"

# Process each username
for USERNAME in "${USERNAMES[@]}"; do
  echo ""
  echo "=========================================="
  echo "Processing user: $USERNAME"
  echo "=========================================="
  
  # Generate secure 32-character alphanumeric password
  echo -e "\nGenerating secure password..."
  if command -v openssl &> /dev/null; then
    NEW_PASSWORD=$(openssl rand -base64 32 | tr -dc 'A-Za-z0-9' | head -c 32)
  else
    NEW_PASSWORD=$(cat /dev/urandom | tr -dc 'A-Za-z0-9' | head -c 32)
  fi
  
  # Validate password was generated
  if [[ -z "$NEW_PASSWORD" || ${#NEW_PASSWORD} -lt 32 ]]; then
    echo "Error: Failed to generate secure password for user '$USERNAME'"
    exit 1
  fi
  
  # Merge new secret with existing ones
  MERGED_SECRETS=$(echo "$MERGED_SECRETS" | jq --arg username "$USERNAME" --arg password "$NEW_PASSWORD" \
    '. + {($username): $password}')
  
  if [[ -z "$MERGED_SECRETS" || "$MERGED_SECRETS" == "null" ]]; then
    echo "Error: Failed to merge secrets for user '$USERNAME'"
    exit 1
  fi
  
  # Get user ID from Keycloak
  echo -e "\nLooking up user in Keycloak realm '$TENANT'..."
  USER_LOOKUP_RESPONSE=$(curl -s -X GET "$KEYCLOAK_URL/admin/realms/$TENANT/users?username=$USERNAME&exact=true" \
    -H "Authorization: Bearer $ACCESS_TOKEN")
  echo -e "\nUser lookup response:\n\n$USER_LOOKUP_RESPONSE"
  
  USER_ID=$(echo "$USER_LOOKUP_RESPONSE" | jq -r '.[0].id')
  
  if [[ "$USER_ID" == "null" || -z "$USER_ID" ]]; then
    echo "⚠ Warning: User '$USERNAME' not found in realm '$TENANT' - skipping Keycloak password reset"
    FAILED_USERS+=("$USERNAME (not found in Keycloak)")
    USER_PASSWORDS[$USERNAME]=$NEW_PASSWORD
    continue
  fi
  
  echo -e "\n✓ Found user with ID: $USER_ID"
  
  # Reset user password in Keycloak
  echo -e "\nResetting password in Keycloak..."
  PASSWORD_RESET_RESPONSE=$(curl -s -w "\n%{http_code}" -X PUT \
    "$KEYCLOAK_URL/admin/realms/$TENANT/users/$USER_ID/reset-password" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"type\":\"password\",\"value\":\"$NEW_PASSWORD\",\"temporary\":false}")
  
  HTTP_CODE=$(echo "$PASSWORD_RESET_RESPONSE" | tail -n 1)
  
  if [[ "$HTTP_CODE" != "204" && "$HTTP_CODE" != "200" ]]; then
    echo "⚠ Warning: Failed to reset password in Keycloak (HTTP $HTTP_CODE)"
    FAILED_USERS+=("$USERNAME (Keycloak password reset failed)")
    USER_PASSWORDS[$USERNAME]=$NEW_PASSWORD
    continue
  fi
  
  echo -e "\n✓ Password reset in Keycloak for user: $USERNAME"
  USER_PASSWORDS[$USERNAME]=$NEW_PASSWORD
done

# Write all merged secrets back to Vault in one call
echo -e "\n=========================================="
echo "Writing all secrets back to Vault..."
echo "=========================================="
echo -e "\nMerged secrets:\n\n$MERGED_SECRETS"

VAULT_RESPONSE=$(curl -s -w "\n%{http_code}" --header "X-Vault-Token: $VAULT_TOKEN" \
  --request POST \
  --data "{\"data\": $MERGED_SECRETS}" \
  "$VAULT_URL/v1/secret/data/folio/$TENANT")

HTTP_CODE=$(echo "$VAULT_RESPONSE" | tail -n 1)
RESPONSE_BODY=$(echo "$VAULT_RESPONSE" | sed '$d')
echo -e "\nVault write response (HTTP $HTTP_CODE): $RESPONSE_BODY"

if [[ "$HTTP_CODE" == "403" ]]; then
  echo "Error: Vault authentication failed - invalid or insufficient token permissions"
  exit 1
elif [[ "$HTTP_CODE" != "200" && "$HTTP_CODE" != "204" ]]; then
  echo "Error: Failed to set Vault secrets (HTTP $HTTP_CODE)"
  echo "$RESPONSE_BODY"
  exit 1
fi

echo -e "\n✓ All secrets set in Vault at: secret/data/folio/$TENANT"

# Print summary
echo ""
echo "=========================================="
echo "SUMMARY: Password Update Results"
echo "=========================================="
echo "Tenant: $TENANT"
echo ""
echo "Passwords generated and stored in Vault:"

# Sort usernames for consistent output
for USERNAME in $(echo "${!USER_PASSWORDS[@]}" | tr ' ' '\n' | sort); do
  echo "  - $USERNAME: ${USER_PASSWORDS[$USERNAME]}"
done

if [[ ${#FAILED_USERS[@]} -gt 0 ]]; then
  echo ""
  echo "⚠ Warnings (passwords stored in Vault but Keycloak updates failed):"
  for FAILED in "${FAILED_USERS[@]}"; do
    echo "  - $FAILED"
  done
  echo ""
  echo "Note: These users' passwords were saved to Vault but require manual Keycloak update."
fi

echo "=========================================="