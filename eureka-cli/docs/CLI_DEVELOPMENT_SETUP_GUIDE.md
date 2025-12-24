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
