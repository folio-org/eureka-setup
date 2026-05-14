# ERM Profile Setup Guide

## Purpose

- Support commands to deploy a local dev environment with combined and ERM applications

## Commands

### Deploy applications with Eureka CLI

- Deploy the parent application with ACQ modules

```bash
eureka-cli deployApplication -p combined-native -oq
```

> Check README.md in eureka-cli/ folder for how to build the local native sidecar Docker image

- Deploy the child application with ERM modules

```bash
eureka-cli deployApplication -p erm -oq
```

### Clone, prepare and intercept ERM modules with Eureka CLI

- Git clone each module into a folder, e.g. `~/Folio/erm-modules`

```bash
cd ~/Folio/erm-modules/
git clone git@github.com:folio-org/mod-agreements.git
git clone git@github.com:folio-org/mod-licenses.git
git clone git@github.com:folio-org/mod-oa.git
git clone git@github.com:folio-org/mod-serials-management.git
```

- Change branch to `thunderjet-erm`

```bash
cd ~/Folio/erm-modules/mod-agreements && git checkout thunderjet-erm
cd ~/Folio/erm-modules/mod-licenses && git checkout thunderjet-erm
cd ~/Folio/erm-modules/mod-oa && git checkout thunderjet-erm
cd ~/Folio/erm-modules/mod-serials-management && git checkout thunderjet-erm
```

- Open erm-modules as a folder with IntelliJ, and import each module with Gradle

> Set JDK either to Java 17 or 19

- Run each module in IntelliJ with the supplied run configurations found in `run-configs/`

> The first few minutes are necessary for the module to boot up

- Run module interception with the CLI

```bash
eureka-cli -p erm interceptModule -n mod-agreements -g -m 36105 -s 37105
eureka-cli -p erm interceptModule -n mod-licenses -g -m 36106 -s 37106
eureka-cli -p erm interceptModule -n mod-oa -g -m 36107 -s 37107
eureka-cli -p erm interceptModule -n mod-serials-management -g -m 36108 -s 37108
```

- When the interception is no longer required, disable it for each module to prevent deadlocking on application undeployment

```bash
eureka-cli -p erm interceptModule -n mod-agreements -r
eureka-cli -p erm interceptModule -n mod-licenses -r
eureka-cli -p erm interceptModule -n mod-oa -r
eureka-cli -p erm interceptModule -n mod-serials-management -r
```

> ERM application cannot be undeployed unless the interception is disabled

### Undeploy all application

```bash
eureka-cli undeployApplication -p erm
eureka-cli undeployApplication -p combined-native
```
