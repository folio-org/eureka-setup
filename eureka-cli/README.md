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
  - Add `127.0.0.1 kafka.eureka` entry to `/etc/hosts`
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

```shell
aws ecr list-images --repository-name [project_name] --no-paginate --output table
```

### Build a binary
  
```shell
mkdir -p ./bin
env GOOS=windows GOARCH=amd64 go build -o ./bin .
```

> See BUILD.md to build a platform-specific binary

### (Optional) Setup config in home folder

- This config will be used by default if `-c` or `--config` flag is not specified

```shell
./bin/eureka-cli.exe setup
```

### Deploy minimal platform application

- Use a specific config: `-c` or `--config`
- Enable debug: `-d` or `--debug`

```shell
./bin/eureka-cli.exe -c ./config.minimal.yaml deployApplication
```

- Undeploy using:

> ./bin/eureka-cli.exe -c ./config.minimal.yaml undeployApplication
