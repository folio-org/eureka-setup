# Development guide

## Purpose

- Development commands to aid with Live Compilation, Debugger in VSCode and Module Interception in IntelliJ

## Commands

### Enable Live Compilation

- Open a new shell terminal
- `cd` into `eureka-setup/eureka-cli`
- Install `air` binary: `go install github.com/air-verse/air@latest`
- Run `air` to enable live compilation

> Will poll for code changes to recreate a binary in `./bin` folder

- See `.air.toml` for more settings on live compilation

### Enable Debugger in VSCode to analyze Eureka CLI commands

- Go to _Run And Debug_ in the VSCode
- Click on _create a launch.json file_
- Select _GO_ and then _GO: Launch Package_
- Replace the generated `launch.json` with the one below and save

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Eureka CLI deployApplication command",
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
      "name": "Eureka CLI interceptModule command",
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
    }
  ]
}
```

- Add breakpoints and click on _RUN AND DEBUG Start Debugging_

> Must undeploy previously deployed application before starting

- The `args` can be modified to reflect which CLI command is to be debugged, e.g. `"args": ["createUsers", "-d"]` will run `createUsers` command in the debugged mode with verbose logs

### Enable Module Interception in IntelliJ (an example for mod-orders and mod-finance)

- `cd` into `eureka-setup/eureka-cli`
- Deploy the Eureka environment using the _combined_ profile: `eureka-cli deployApplication`

> If `--profile` flag is not explicitly specified, __combined__ is assumed as the default profile

- Deploy the custom sidecars into the Eureka environment

> Verify that `host.docker.internal` is set in `/etc/hosts` or use default Docker Gateway IP `172.17.0.1` in Linux in the URLs

```bash
# Find the module id that you want to intercept with listModules command
eureka-cli listModules -n mod-finance
eureka-cli listModules -n mod-orders

eureka-cli interceptModule -n mod-orders -m http://host.docker.internal:36001 -s http://host.docker.internal:37001
eureka-cli interceptModule -n mod-finance -m http://host.docker.internal:36002 -s http://host.docker.internal:37002

# Alternatively, you can use the default Kong gateway with --defaultGateway/-g flag, and by passing module and sidecar ports directly
eureka-cli interceptModule -n mod-orders -gm 36001 -s 37001
eureka-cli interceptModule -n mod-finance -gm 36002 -s 37002
```

> Module and sidecar exposed ports in the example are not fixed and can be changed to suit your needs, e.g. if your mod-orders in IntelliJ is usually started on port 9800, `--moduleUrl` / `-m` will be <http://host.docker.internal:9800>

- Start the module instances in IntelliJ

<table>
<caption>IntelliJ Run Configurations and Env Files</caption>
<thead>
<tr>
<th>mod-orders</th>
<th>mod-finance</th>
</tr>
</thead>
<tbody>
<tr>
<td><img src="../images/mod_orders_run_config.png" alt="mod_orders_run_config" /></td>
<td><img src="../images/mod_finance_run_config.png" alt="mod_finance_run_config" /></td>
</tr>
<tr>
<td>

```conf
DB_HOST=localhost
DB_PORT=5432
DB_DATABASE=folio
DB_USERNAME=folio_rw
DB_PASSWORD=supersecret
DB_CHARSET=UTF-8
DB_MAXPOOLSIZE=50
DB_QUERYTIMEOUT=60000

ENV=folio
KAFKA_HOST=localhost
KAFKA_PORT=9092
```

</td>
<td>

```conf
DB_HOST=localhost
DB_PORT=5432
DB_DATABASE=folio
DB_USERNAME=folio_rw
DB_PASSWORD=supersecret
DB_CHARSET=UTF-8
DB_MAXPOOLSIZE=50
DB_QUERYTIMEOUT=60000

ENV=folio
KAFKA_HOST=localhost
KAFKA_PORT=9092
```

</td>
</tr>
</tbody>
</table>

- Test _mod-finance_ interception by creating a _Fund Budget_ in the _Finance App_

> Expect: Logs being created for _mod-finance_ deployed in IntelliJ

- After that, create a _Purchase Order_ with a _Purchase Order Line_ and an attached _Fund Distribution_, using the _Fund_ created in the _Finance App_, within the _Orders App_

> Expect: Logs being created for _mod-orders_ and _mod-finance_ deployed in IntelliJ

### Disable Module Interception

- Stop the module instances in IntelliJ
- Restore the default modules and sidecars in the Eureka environment

```bash
eureka-cli interceptModule -n mod-orders -r
eureka-cli interceptModule -n mod-finance -r
```
