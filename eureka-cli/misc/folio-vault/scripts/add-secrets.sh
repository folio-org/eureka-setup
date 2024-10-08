# enable KV v2 Secrets Engine and populate secrets
# see https://developer.hashicorp.com/vault/api-docs/secret/kv/kv-v2 for API reference

set -e

secretsUri="$VAULT_ADDR/v1/secret"

# enable KV secret engine under "/secret" root
vault secrets enable -version=2 -path=secret -description='Eureka testing secrets' kv

# configure engine
curl -sv -X POST \
   -H "Content-Type: application/json" \
   -H "X-Vault-Token: $VAULT_TOKEN" \
   -d '{"max_versions":1,"cas_required":false,"delete_version_after":0}' \
   "$secretsUri/config"

# create global secrets (in folio/master)
secretsData='{
  "data": {
    "supersecret": "supersecret",
    "mgr-applications": "supersecret",
    "mgr-tenants": "supersecret",
    "mgr-tenant-entitlements": "supersecret"
  }
}'

curl -sv -X POST \
   -H "Content-Type: application/json" \
   -H "X-Vault-Token: $VAULT_TOKEN" \
   -d "$secretsData" \
   "$secretsUri/data/folio/master"

# list secrets created
echo "$(date -u +%FT%T.%3NZ) [INFO] Secrets in folio/master:"

curl -sv -X GET -H "X-Vault-Token: $VAULT_TOKEN" "$secretsUri"/data/folio/master

