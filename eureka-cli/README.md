# Eureka CLI

## Purpose

- Deploy Folio and Eureka modules into Eureka local development environment

## Commands

### Prerequisites

- Install necessary prerequisites (Windows users)
  - [AWS CLI](<https://docs.aws.amazon.com/cli/latest/userguide/>)
  - [GO](<https://go.dev/doc/install>)
- Request AWS ECR access tokens from **Folio Kitfox Team** in Slack devops channel

### Primary commands

- Enable Live Compilation in Terminal session #1
  - Will compile a Windows binary into `bin` folder

```bash
air
```

- Deploy/undeploy all modules and sidecars in another Terminal session #2

```bash
# 1. Prepare AWS CLI for AWS ECR usage (to be done at least once)
# DO NOT SHARE ANY AWS TOKENS OR SECRETS WITH ANYONE 
# OR PUSH ANY OF THESE TOKENS OR SECRETS INTO ANY REPOSITORY
aws configure set aws_access_key_id [access_key] 
aws configure set aws_secret_access_key [secret_key] 
aws configure set default.region [region] 

# 2. Setup CLI config
./bin/eureka-cli.exe setup

# 3. Deploy all modules and sidecars
# Enabled debug (optional) flag: -d or --debug
AWS_SDK_LOAD_CONFIG=true ./bin/eureka-cli.exe deployModules

# (Optional) Undeploy all modules and sidecars
# Enabled debug (optional) flag: -d or --debug
./bin/eureka-cli.exe undeployModules

# (Optional) Undeploy single module and its sidecar
# Module name (required) flag: -m or --moduleName 
# Enabled debug (optional) flag: -d or --debug
./bin/eureka-cli.exe undeployModule -m mod-login-keycloak
```
