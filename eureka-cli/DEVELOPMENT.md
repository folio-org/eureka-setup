# Development

## Purpose

- Auxiliary commands and settings for development

## Commands

### Enable live compilation

- Open a new shell terminal
- `cd` into `eureka-setup/eureka-cli`
- Run `air` to enable live compilation

> For every saved code changes a new binary will be made in `./bin` folder

- See `.air.toml` for more settings on live compilation

### Enable debugger in VisualStudio Code

- Go to *Run And Debug* in the VSCode
- Click on *create a launch.json file*
- Select *GO* and then *GO: Launch Package*
- Replace the generated `launch.json` with the one below and save

```json
{
  // Use IntelliSense to learn about possible attributes.
  // Hover to view descriptions of existing attributes.
  // For more information, visit: https://go.microsoft.com/fwlink/?linkid=830387
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Eureka CLI Debugger",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${cwd}/eureka-setup/eureka-cli",
      "output": "${cwd}/eureka-setup/eureka-cli/bin/eureka-cli-debug.exe",
      "env": {
        "GOOS": "windows", 
        "GOARCH": "amd64"
      },
      "args": ["deployApplication", "-d"],
      "showLog": true,
    }
  ]
}
```

- Add breakpoints and click on *RUN AND DEBUG Start Debugging*

> Must undeploy previously deployed application before starting

- The `args` can be modified to reflect which CLI command is to be debugged, e.g. `"args": ["createUsers", "-d"]` will run `createUsers` command in the debugged mode with verbose logs
