#!/bin/sh

set -e

echo "$(date -u +%FT%T.%3NZ) [INFO] create-admin-user.sh: Creating admin user..."

# Enable userpass authentication method
vault auth enable userpass || echo "$(date -u +%FT%T.%3NZ) [WARN] create-admin-user.sh: userpass auth already enabled"

# Create admin policy with root privileges
vault policy write admin-policy - <<EOF
# Admin policy with full access
path "*" {
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
EOF

# Create admin user with password
# Using default credentials: username=admin, password=admin
vault write auth/userpass/users/admin \
  password=admin \
  policies=admin-policy

echo "$(date -u +%FT%T.%3NZ) [INFO] create-admin-user.sh: Admin user created successfully!"
echo "$(date -u +%FT%T.%3NZ) [INFO] create-admin-user.sh: Username: admin"
echo "$(date -u +%FT%T.%3NZ) [INFO] create-admin-user.sh: Password: admin"
echo "$(date -u +%FT%T.%3NZ) [INFO] create-admin-user.sh: You can now login via UI or CLI with these credentials"

