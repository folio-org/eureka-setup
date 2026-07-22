# CLI Development Setup Guide

## Purpose

- Auxiliary CLI development setup commands to aid developers with enabling live Compilation, and debugging source code in VSCode

## Enable Live Compilation (air binary)

- Open a new shell terminal
- `cd` into `eureka-setup/eureka-cli`
- Install `air` binary: `go install github.com/air-verse/air@latest`
- Run `air` to enable live compilation

> Will poll for code changes to recreate a binary in `./bin` folder

- See `.air.toml` for more settings on live compilation

## Enable Debugger in VSCode (delve binary)

- Open a new shell terminal
- Install `delve` binary: `go install github.com/go-delve/delve/cmd/dlv@latest`
- Go to _Run And Debug_ in the VSCode
- Click on _create a launch.json file_
- Select _GO_ and then _GO: Launch Package_
- Replace the generated `launch.json` with the one below and save

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Eureka CLI deployApplication",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${cwd}/eureka-cli",
      "output": "${cwd}/bin/eureka-cli-debug.exe",
      "env": {
        "GOOS": "windows",
        "GOARCH": "amd64"
      },
      "args": ["--profile", "combined", "deployApplication", "-d"],
      "showLog": true
    },

    {
      "name": "Eureka CLI undeployApplication",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${cwd}/eureka-cli",
      "output": "${cwd}/bin/eureka-cli-debug.exe",
      "env": {
        "GOOS": "windows",
        "GOARCH": "amd64"
      },
      "args": ["--profile", "combined", "undeployApplication"],
      "showLog": true
    },

    {
      "name": "Eureka CLI interceptModule",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${cwd}/eureka-cli",
      "output": "${cwd}/bin/eureka-cli-debug.exe",
      "env": {
        "GOOS": "windows",
        "GOARCH": "amd64"
      },
      "args": ["--profile", "combined", "interceptModule", "-n", "mod-orders", "-g", "-m", "36001", "-s", "37001"],
      "showLog": true
    },

    {
      "name": "Eureka CLI upgradeModule (upgrade)",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${cwd}/eureka-cli",
      "output": "${cwd}/bin/eureka-cli-debug.exe",
      "env": {
        "GOOS": "windows",
        "GOARCH": "amd64"
      },
      "args": ["--profile", "combined", "upgradeModule", "-n", "mod-orders", "--modulePath", "~/Folio/folio-modules/mod-orders"],
      "showLog": true
    },

    {
      "name": "Eureka CLI upgradeModule (downgrade)",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${cwd}/eureka-cli",
      "output": "${cwd}/bin/eureka-cli-debug.exe",
      "env": {
        "GOOS": "windows",
        "GOARCH": "amd64"
      },
      "args": ["--profile", "combined", "upgradeModule", "-n", "mod-orders", "--moduleVersion", "13.1.0-SNAPSHOT.1093", "--namespace", "folioci", "--modulePath", "~/Folio/folio-modules/mod-orders"],
      "showLog": true
    }
  ]
}
```

- Add breakpoints and click on _RUN AND DEBUG Start Debugging_

> Must undeploy previously deployed application before starting

- The `args` can be modified to reflect which CLI command is to be debugged, e.g. `"args": ["createUsers", "-d"]` will run `createUsers` command in the debugged mode with verbose logs

## Add new CLI command

- To add a new CLI command install the latest versions of Cobra and Viper auxiliary CLIs

```bash
go get -u github.com/spf13/cobra@latest
go get -u github.com/spf13/viper@latest
```

- After that you can create a new Eureka CLI command (also know as an **action**) with the Cobra CLI

```bash
cd eureka-cli
cobra-cli add [my_new_command_name] --viper --author "Open Library Foundation" --license "Apache"
```

> The new command file will be created in the cmd/ package folder

- Once created, register the new command action name in `action/action_name.go`
- If the command receives arguments, register each new param in `action/action_param.go`

## Update system components

_Keycloak_ and _Kong_ are pulled as upstream FOLIO images, selected via `FOLIO_KEYCLOAK_NAMESPACE` / `FOLIO_KEYCLOAK_VERSION` and `FOLIO_KONG_NAMESPACE` / `FOLIO_KONG_VERSION` in `misc/.env`. The defaults track the `folioci` snapshot namespace at `:latest` and are not pinned; for a reproducible setup, switch to `folioorg` with a release version, or keep `folioci` and pin a specific `*-SNAPSHOT.*` version. Note that Docker only pulls an image that is missing locally, so a floating tag like `:latest` is reused as first pulled on every subsequent deployment - refresh it deliberately by pinning a newer version, or manually from `~/.eureka/misc`:

```bash
docker compose --project-name eureka pull kong keycloak-internal
eureka-cli deploySystem
```

To test locally patched _Keycloak_ or _Kong_ images, build the image from your local checkout, tag it under any namespace/version, and point the `.env` variables at that tag - `docker compose` uses the local image as long as it exists and never pulls it:

```bash
docker build -t dev/folio-keycloak:local ~/git/folio-keycloak
# In misc/.env: FOLIO_KEYCLOAK_NAMESPACE=dev, FOLIO_KEYCLOAK_VERSION=local
```

_Netcat_ and _Vault_ are always built locally (from `misc/folio-netcat` and `misc/folio-vault`) and are never pulled from a registry. `docker compose up` builds each once and then reuses the cached image, so after a CLI upgrade that changes their `Dockerfile` you must refresh the local image explicitly - otherwise the stale one is reused:

```bash
# Re-extract the embedded build context into ~/.eureka, then force a rebuild
eureka-cli buildSystem -o          # -o overwrites files under ~/.eureka
# or, in one step during a deploy:
eureka-cli deployApplication -bo
```

> Vault is stateful: its bootstrap scripts are baked into the image and run against the persisted `vault-*` volumes on startup. A stale Vault image shows up as a missing Userpass sign-in method or a failed `admin/admin` login (see Troubleshooting). If a rebuilt image changes the bootstrap in a way that is incompatible with existing state, wipe the volumes and re-bootstrap with `eureka-cli undeploySystem` (`compose down --volumes`) followed by a fresh deploy.

_Kafka tools_ is the official `apache/kafka` image (JVM variant) idling on `sleep infinity`, pinned to `KAFKA_VERSION` so it always matches the broker. It needs no custom image or manual maintenance - bump `KAFKA_VERSION` in `misc/.env` to move both the broker and the tools together.
