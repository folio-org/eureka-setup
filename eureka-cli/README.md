# Eureka CLI

## Purpose

- Deploy Folio and Eureka modules into Eureka local development environment

## Commands

### Prerequisites

- Install necessary prerequisites (Windows users)
  - [GO](<https://go.dev/doc/install>)

### Primary commands

- Enable Live Compilation in Terminal session #1
  - Will compile a Windows binary into `bin` folder

```bash
air
```

- Deploy/undeploy all modules and sidecars in another Terminal session #2

```bash
# Setup CLI config
./bin/eureka-cli.exe setup

# Deploy/undeploy system
# Enable debug (optional) flag: -d or --debug
./bin/eureka-cli.exe deploySystem
./bin/eureka-cli.exe undeploySystem

# Deploy/undeploy management modules
# Enable debug (optional) flag: -d or --debug
./bin/eureka-cli.exe deployManagement
./bin/eureka-cli.exe undeployManagement

# Deploy/undeploy other modules and sidecars
# Enable debug (optional) flag: -d or --debug
./bin/eureka-cli.exe deployModules
./bin/eureka-cli.exe undeployModules

# (Optional) Undeploy single module and its sidecar
# Module name (required) flag: -m or --moduleName 
# Enable debug (optional) flag: -d or --debug
./bin/eureka-cli.exe undeployModule -m mod-login-keycloak
```
