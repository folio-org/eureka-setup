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
# 1. Setup CLI config
./bin/eureka-cli.exe setup

# 2. Deploy management modules
# Enable debug (optional) flag: -d or --debug
./bin/eureka-cli.exe deployManagement

# 3. Deploy other modules and sidecars
# Enable debug (optional) flag: -d or --debug
./bin/eureka-cli.exe deployModules

# (Optional) Undeploy other modules and sidecars
# Enable debug (optional) flag: -d or --debug
./bin/eureka-cli.exe undeployModules

# (Optional) Undeploy management modules
# Enable debug (optional) flag: -d or --debug
./bin/eureka-cli.exe undeployManagement

# (Optional) Undeploy single module and its sidecar
# Module name (required) flag: -m or --moduleName 
# Enable debug (optional) flag: -d or --debug
./bin/eureka-cli.exe undeployModule -m mod-login-keycloak
```
