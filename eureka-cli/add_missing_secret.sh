#!/bin/bash

set -euo pipefail

# Script to set Vault secret and reset Keycloak user password
# Usage: ./add_missing_secret.sh [OPTIONS]

# Default values
KEYCLOAK_URL="http://keycloak.eureka:8080"
TENANT=""
USERNAME=""
VAULT_TOKEN=$(eureka-cli getVaultRootToken)
VAULT_URL="http://localhost:8200"

# Display usage
usage() {
  cat << EOF
Usage: $0 [OPTIONS]

Set Vault secret and reset Keycloak user password.

OPTIONS:
  -h, --help                       Display this help message
  -t, --tenant TENANT              Tenant name (e.g., diku) [required]
  -u, --username USERNAME          Username to reset [required]
  --keycloak-url URL               Keycloak URL (default: http://keycloak.eureka:8080)
  --vault-url URL                  Vault URL (default: http://localhost:8200)

EXAMPLE:
  $0 -t diku -u admin
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
      USERNAME="$2"
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

if [[ -z "$USERNAME" ]]; then
  echo "Error: Username is required"
  usage
fi

# Generate secure 32-character alphanumeric password
echo -e "\nGenerating secure password..."
if command -v openssl &> /dev/null; then
  NEW_PASSWORD=$(openssl rand -base64 32 | tr -dc 'A-Za-z0-9' | head -c 32)
else
  NEW_PASSWORD=$(cat /dev/urandom | tr -dc 'A-Za-z0-9' | head -c 32)
fi

echo -e "\nSetting password for user '$USERNAME' in tenant '$TENANT'"

# Set secret in Vault (upsert - creates or updates)
echo -e "\nSetting secret in Vault..."

# Read existing secrets from Vault
echo -e "\nReading existing secrets from Vault..."
VAULT_READ_RESPONSE=$(curl -s --header "X-Vault-Token: $VAULT_TOKEN" \
  "$VAULT_URL/v1/secret/data/folio/$TENANT")
echo -e "\nVault read response:\n\n$VAULT_READ_RESPONSE"

EXISTING_SECRETS=$(echo "$VAULT_READ_RESPONSE" | jq -r '.data.data // {}')
echo -e "\nExisting secrets:\n\n$EXISTING_SECRETS"

# Merge new secret with existing ones
MERGED_SECRETS=$(echo "$EXISTING_SECRETS" | jq --arg username "$USERNAME" --arg password "$NEW_PASSWORD" \
  '. + {($username): $password}')
echo -e "\nMerged secrets:\n\n$MERGED_SECRETS"

# Write merged secrets back to Vault
echo -e "\nWriting secrets back to Vault..."
VAULT_RESPONSE=$(curl -s -w "\n%{http_code}" --header "X-Vault-Token: $VAULT_TOKEN" \
  --request POST \
  --data "{\"data\": $MERGED_SECRETS}" \
  "$VAULT_URL/v1/secret/data/folio/$TENANT")

HTTP_CODE=$(echo "$VAULT_RESPONSE" | tail -n 1)
RESPONSE_BODY=$(echo "$VAULT_RESPONSE" | sed '$d')
echo -e "\nVault write response (HTTP $HTTP_CODE): $RESPONSE_BODY"

if [[ "$HTTP_CODE" == "403" ]]; then
  echo "Error: Vault authentication failed - invalid or insufficient token permissions"
  echo "Please provide a valid Vault token using: -v YOUR_VAULT_TOKEN"
  exit 1
elif [[ "$HTTP_CODE" != "200" && "$HTTP_CODE" != "204" ]]; then
  echo "Error: Failed to set Vault secret (HTTP $HTTP_CODE)"
  echo "$RESPONSE_BODY"
  exit 1
fi

echo -e "\n✓ Secret set in Vault at: secret/data/folio/$TENANT (key: $USERNAME)"

# Get Keycloak access token
echo "Getting Keycloak access token..."
ACCESS_TOKEN=$(eureka-cli getKeycloakAccessToken --tokenType master-admin-cli)

if [[ -z "$ACCESS_TOKEN" || "$ACCESS_TOKEN" == "null" ]]; then
  echo "Error: Failed to get Keycloak access token"
  exit 1
fi

# Get user ID from Keycloak
echo -e "\nLooking up user in Keycloak realm '$TENANT'..."
USER_LOOKUP_RESPONSE=$(curl -s -X GET "$KEYCLOAK_URL/admin/realms/$TENANT/users?username=$USERNAME&exact=true" \
  -H "Authorization: Bearer $ACCESS_TOKEN")
echo -e "\nUser lookup response:\n\n$USER_LOOKUP_RESPONSE"

USER_ID=$(echo "$USER_LOOKUP_RESPONSE" | jq -r '.[0].id')

if [[ "$USER_ID" == "null" || -z "$USER_ID" ]]; then
  echo "Error: User '$USERNAME' not found in realm '$TENANT'"
  exit 1
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
  echo "Error: Failed to reset password in Keycloak (HTTP $HTTP_CODE)"
  echo "$RESPONSE_BODY"
  exit 1
fi

echo -e "\n✓ Password reset in Keycloak"
echo ""
echo "=========================================="
echo "SUCCESS: Password updated successfully"
echo "=========================================="
echo "Tenant:   $TENANT"
echo "Username: $USERNAME"
echo "Password: $NEW_PASSWORD"
echo "=========================================="