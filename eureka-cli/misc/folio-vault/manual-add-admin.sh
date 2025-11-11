#!/bin/sh

# This script adds an admin user to an existing vault instance
# Run this inside the vault container or with VAULT_ADDR and VAULT_TOKEN set

set -e

echo "Adding admin user to existing vault..."

# Check if VAULT_TOKEN is set
if [ -z "$VAULT_TOKEN" ]; then
  echo "ERROR: VAULT_TOKEN environment variable is not set"
  echo "Please set it with: export VAULT_TOKEN=<your-root-token>"
  exit 1
fi

# Enable userpass authentication method
vault auth enable userpass 2>/dev/null || echo "userpass auth already enabled"

# Create admin policy with root privileges
vault policy write admin-policy - <<EOF
# Admin policy with full access
path "*" {
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
EOF

# Create admin user with password
vault write auth/userpass/users/admin \
  password=admin \
  policies=admin-policy

echo "Admin user created successfully!"
echo "Username: admin"
echo "Password: admin"
echo ""
echo "You can now login to Vault UI at http://localhost:8200"
echo "Select 'Username' as the authentication method and use the credentials above"

