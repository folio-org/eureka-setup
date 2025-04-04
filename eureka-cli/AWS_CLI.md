# AWS CLI

## Purpose

- AWS CLI preparation commands

## Commands

### Install AWS CLI

> Download latest version for your OS from: <https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html>

### Configure AWS CLI

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
