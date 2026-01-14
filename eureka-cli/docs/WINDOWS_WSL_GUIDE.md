# Windows WSL Guide

## Purpose

- Provide settings and auxiliary WSL2 and Docker CLI commands

## Commands

On Windows, the Eureka CLI relies on WSL2 to deploy containers, usually via either Rancher or Docker Desktop. To make the deployment more stable we need to create a `.wslconfig` file with necessary settings in the home directory. This file will limit the container daemon from consuming excessive amounts of host resources, making the deployment and subsequent work more comfortable.

- Below is the optimal and minimal settings for achieving a stable deployment

```txt
[wsl2]
processors=6
memory=24GB
swap=6GB

kernelCommandLine="cgroup_no_v1=all sysctl.vm.max_map_count=262144"
vmIdleTimeout=3600000
networkingMode=mirrored
guiApplications=false
```

> For more information, check the official reference: <https://learn.microsoft.com/en-us/windows/wsl/wsl-config>

- To apply these settings, the Rancher or Docker Desktop must be closed, and a WSL shutdown command must be issued

```bash
# 1. Close your running instances of Rancher Desktop

# 2. Issue a shutdown command to WSL
wsl --shutdown
```

- Check the statuses of your distros after the shutdown

```bash
wsl --list -v
```

> We expect `rancher-desktop` or `docker-desktop` to be in the status of *STOPPED*

After that you can launch Rancher or Docker Desktop as normal and see that the WSL2 settings are being applied.

- Check with `docker info`

```bash
docker info
```

- Example output

```txt
Server:
 ...
 Kernel Version: 6.6.87.2-microsoft-standard-WSL2
 Operating System: Rancher Desktop WSL Distribution
 OSType: linux
 Architecture: x86_64
 CPUs: 6
 Total Memory: 23.47GiB
 ...
```

- Or post-deployment with stats

```bash
docker stats --no-stream
```

If your machine is resource-constrained, consider deploying the environment with MS Teams, Google Chrome and IntelliJ closed.
