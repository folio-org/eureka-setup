# Eureka Setup for local development

## Purpose

- Docker compose files and shell scripts to setup local development Eureka environment

## Commands

### Prerequisites

- Install necessary prerequisites (Windows users)
  - [Git](<https://git-scm.com/>)
  - [Choco](<https://chocolatey.org/install>)
  - [AWS CLI](<https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html>)
- Install **jq** via `choco install jq`
- Request AWS ECR access tokens from **Folio Kitfox Team** in Slack devops channel
- Add hostname to `/etc/hosts` using plain *notepad* opened via *Run as Administrator*:
  - `127.0.0.1 keycloak.eureka`
  
### Notes

- **Git** (i.e. **Git Bash**) already comes shipped with  `curl`
- **Choco** requires **admin privileges** to **install** and **use** at all times
- Nearly all ENV vars have been scrubbed for simplification
- Nearly all passwords have been replaced with `superAdmin`
- All commands have been adapted for Windows **Git Bash** usages
- All hostnames have been appended with optional `.eureka` suffix

### Primary commands

- Start system & module containers
- Links to UI URLs provided by the system containers:
  - [Keycloak UI (username:admin, password:admin)](<http://keycloak.eureka:8080>)
  - [Vault UI (use your extracted VAULT_TOKEN ENV var)](<http://localhost:8200>)
  - [Kafka UI (no auth)](<http://localhost:9080>)
  - [Kong UI (no auth)](<http://localhost:8002>)

```bash
# 1. Prepare AWS CLI for AWS ECR usage (to be done at least once)
# DO NOT SHARE ANY AWS TOKENS OR SECRETS WITH ANYONE 
# OR PUSH ANY OF THESE TOKENS OR SECRETS INTO ANY REPOSITORY
aws configure set aws_access_key_id [access_key] 
aws configure set aws_secret_access_key [secret_key] 
aws configure set default.region [region] 

# 2. Docker Login into AWS ECR
# Keys such as region, username and account_id will be provided by the Kitfox Team
aws ecr get-login-password --region [region] | docker login --username [username] --password-stdin [account_id].dkr.ecr.[region].amazonaws.com

# (Optional) List available version for a repository
aws ecr list-images --repository-name mgr-applications --no-paginate --output table
aws ecr list-images --repository-name mgr-tenants --no-paginate --output table
aws ecr list-images --repository-name mgr-tenant-entitlements --no-paginate --output table

# 3. Start all components
# WARNING: Before starting make sure to replace [account_id] and [region] in .env with your provided values
{
docker compose -p eureka -f docker-compose.core.yml up -d --build --always-recreate-deps --force-recreate && sleep 60
export VAULT_TOKEN=$(docker logs vault 2>&1 | grep 'init.sh: Root VAULT TOKEN is:' | sed 's/.*://' | xargs); echo "Using Vault Token: $VAULT_TOKEN"
docker compose -p eureka -f docker-compose.mgr.yml up -d --force-recreate && sleep 60
}

# 4. Monitor services
# All services with a health checks must be healthy 
docker compose -p eureka ps -a --format 'table {{.ID}}\t{{.Name}}\t{{.Status}}\t{{.Image}}'

# (Optional) Start, Restart or Kill a container(s)
export VAULT_TOKEN=$(docker logs vault 2>&1 | grep 'init.sh: Root VAULT TOKEN is:' | sed 's/.*://' | xargs); echo "Using Vault Token: $VAULT_TOKEN"
docker compose -p eureka -f docker-compose.mgr.yml restart mgr-tenants mgr-tenant-entitlements mgr-applications
docker compose -p eureka -f docker-compose.mgr.yml kill mgr-tenants mgr-tenant-entitlements mgr-applications

# (Optional) Monitor logs for an individual module
docker logs -f --tail 1000 vault
docker logs -f --tail 1000 keycloak
docker logs -f --tail 1000 mgr-tenants
docker logs -f --tail 1000 mgr-tenant-entitlements
docker logs -f --tail 1000 mgr-applications

# (Optional) Stop components
docker compose -p eureka -f docker-compose.mgr.yml down -v 
docker compose -p eureka -f docker-compose.core.yml down -v 

# (Optional) Stop all components
docker compose -p eureka down -v
```

### Secondary commands

- Diagnose system & module ports
  - All must result in something similar to: `keycloak.eureka [192.168.240.9] 8080 (?) open`

```bash
# Check system ports
docker exec -i nginx-netcat bash <<EOF
netcat -zv db.eureka 5432
netcat -zv kafka.eureka 9092
netcat -zv zookeeper.eureka 2181
netcat -zv vault.eureka 8200
netcat -zv keycloak.eureka 8080
netcat -zv keycloak-internal.eureka 8080
netcat -zv okapi.eureka 9130
netcat -zv api-gateway.eureka 8000
netcat -zv api-gateway.eureka 8001
netcat -zv api-gateway.eureka 8002
netcat -zv api-gateway.eureka 8443
netcat -zv api-gateway.eureka 8444

# Check management ports
netcat -zv mgr-tenants.eureka 8081
netcat -zv mgr-tenant-entitlements.eureka  8081
netcat -zv mgr-applications.eureka 8081

echo "DONE"
exit 0
EOF
```

### Tertiary commands

- Reinstall superAdmin client with an admin role in case the import has failed

```bash
docker exec -i keycloak-internal sh <<EOF
# Add superAdmin client
/opt/keycloak/bin/kcadm.sh create clients \
  --target-realm master \
  --set clientId="superAdmin" \
  --set serviceAccountsEnabled=true \
  --set publicClient=false \
  --set clientAuthenticatorType=client-secret \
  --set secret="superAdmin" \
  --set standardFlowEnabled=false

# Add superAdmin client service account admin role
/opt/keycloak/bin/kcadm.sh add-roles \
  --uusername service-account-superadmin \
  --rolename admin \
  --rolename create-realm

echo "DONE"
exit 0
EOF
```
