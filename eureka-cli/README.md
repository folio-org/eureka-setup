# Eureka CLI

## Purpose

- A CLI to deploy local Eureka development environment

## Commands

### Prerequisites

- Install prerequisites:
  - [GO](<https://go.dev/doc/install>)
  - [Rancher Desktop](<https://rancherdesktop.io/>)
  - [AWS CLI](<https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html>)
- Configure hosts:
  - Add `127.0.0.1 keycloak.eureka` entry to `/etc/hosts`
- Monitor using below system components:
  - [Keycloak](<http://keycloak.eureka:8080>): admin:admin
  - [Vault](<http://localhost:8200>): Vault token
  - [Kafka](<http://localhost:9080>): No auth
  - [Kong](<http://localhost:8002>): No auth  

### Prepare AWS CLI

```shell
aws configure set aws_access_key_id [access_key]
aws configure set aws_secret_access_key [secret_key]
aws configure set default.region [region]
```

### Login into AWS ECR

```shell
aws ecr get-login-password --region [region] | docker login --username [username] --password-stdin [account_id].dkr.ecr.[region].amazonaws.com
```

### (Optional) List available project versions

> aws ecr list-images --repository-name [project_name] --no-paginate --output table

### Build binary
  
> go build -o ./bin/eureka-cli.exe .

### Setup config

> ./bin/eureka-cli.exe setup

### Deploy/undeploy entire system

- Use  config (optional) flag: `-c` or `--config`
- Enable debug (optional) flag: `-d` or `--debug`

> ./bin/eureka-cli.exe deploySystem
> ./bin/eureka-cli.exe undeploySystem
