# Eureka Setup for local development

## Purpose

- Supplementary Docker compose files and shell scripts to setup local development Eureka environment

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
aws ecr list-images --repository-name [project_name] --no-paginate --output table

# 3. Start all components
# WARNING: Before starting make sure to replace [account_id] and [region] in .env with your provided values
docker compose -p eureka -f docker-compose.yaml up -d --build --always-recreate-deps --force-recreate && sleep 60

# (Optional) Stop all components
docker compose -p eureka down -v
```

### Secondary commands

- Diagnose system

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
