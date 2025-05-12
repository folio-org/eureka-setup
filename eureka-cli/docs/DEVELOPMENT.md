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

- Go to *Run And Debug* in the VSCode
- Click on *create a launch.json file*
- Select *GO* and then *GO: Launch Package*
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
      "args": ["--config", "config.combined.yaml", "deployApplication", "-d"],
      "showLog": true,
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
      "args": ["--config", "config.combined.yaml", "interceptModule", 
        "-i", "mod-orders:13.1.0-SNAPSHOT.1021",
        "-m", "http://host.docker.internal:36001",
        "-s", "http://host.docker.internal:37001"
      ],
      "showLog": true,
    }
  ]
}
```

- Add breakpoints and click on *RUN AND DEBUG Start Debugging*

> Must undeploy previously deployed application before starting

- The `args` can be modified to reflect which CLI command is to be debugged, e.g. `"args": ["createUsers", "-d"]` will run `createUsers` command in the debugged mode with verbose logs

### Enable Module Interception in IntelliJ (an example for mod-orders and mod-finance)

- `cd` into `eureka-setup/eureka-cli`
- Deploy the Eureka environment using the *combined* profile: `./bin/eureka-cli -c config.combined.yaml deployApplication`
- Deploy the custom sidecars into the Eureka environment

> Verify that `host.docker.internal` is set in `/etc/hosts` or use default Docker Gateway IP `172.17.0.1` in Linux in the URLs

```bash
./bin/eureka-cli -c config.combined.yaml interceptModule -i mod-orders:13.1.0-SNAPSHOT.1021 -m http://host.docker.internal:36001 -s http://host.docker.internal:37001
./bin/eureka-cli -c config.combined.yaml interceptModule -i mod-finance:5.2.0-SNAPSHOT.289 -m http://host.docker.internal:36002 -s http://host.docker.internal:37002
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
; PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_DATABASE=folio
DB_USERNAME=folio_rw
DB_PASSWORD=supersecret
DB_CHARSET=UTF-8
DB_MAXPOOLSIZE=50
DB_QUERYTIMEOUT=60000

; Kafka
ENV=folio
KAFKA_HOST=localhost
KAFKA_PORT=9092

; Okapi (compatible with Kong)
OKAPI_HOST=localhost
OKAPI_PORT=37001
OKAPI_SERVICE_HOST=localhost
OKAPI_SERVICE_PORT=37001
OKAPI_SERVICE_URL=http://localhost:37001
OKAPI_URL=http://localhost:37001
```

</td>
<td>

```conf
; PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_DATABASE=folio
DB_USERNAME=folio_rw
DB_PASSWORD=supersecret
DB_CHARSET=UTF-8
DB_MAXPOOLSIZE=50
DB_QUERYTIMEOUT=60000

; Kafka
ENV=folio
KAFKA_HOST=localhost
KAFKA_PORT=9092

; Okapi (compatible with Kong)
OKAPI_HOST=localhost
OKAPI_PORT=37002
OKAPI_SERVICE_HOST=localhost
OKAPI_SERVICE_PORT=37002
OKAPI_SERVICE_URL=http://localhost:37002
OKAPI_URL=http://localhost:37002
```

</td>
</tr>
</tbody>
</table>

- Perform module healthchecks: `curl -sw "\n" --connect-timeout 3 http://localhost:36001/admin/health http://localhost:36002/admin/health`

> Expect: *"OK"*

- Perform sidecar healthchecks: `curl -sw "\n" --connect-timeout 3 http://localhost:37001/admin/health http://localhost:37002/admin/health`

> Expect: *{ "status": "UP" }*

- Finally test *mod-finance* interception by creating a *Fund Budget* in the *Finance App*

> Expect: Logs being created for *mod-finance* deployed in IntelliJ

- After that, create a *Purchase Order* with a *Purchase Order Line* and an attached *Fund Distribution*, using the *Fund* created in the *Finance App*, within the *Orders App*

> Expect: Logs being created for *mod-orders* and *mod-finance* deployed in IntelliJ

### Disable Module Interception

- Stop the module instances in IntelliJ
- Restore the default modules and sidecars in the Eureka environment

```bash
./bin/eureka-cli -c config.combined.yaml interceptModule -i mod-orders:13.1.0-SNAPSHOT.1021 -r
./bin/eureka-cli -c config.combined.yaml interceptModule -i mod-finance:5.2.0-SNAPSHOT.289 -r
```
