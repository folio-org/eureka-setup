# Copilot Agent Instructions — eureka-cli

## Project overview

Go CLI tool (Cobra + Viper) that orchestrates local FOLIO/Eureka deployments via Docker.
Module: `github.com/folio-org/eureka-setup/eureka-cli`

## Key architecture

| Layer              | Package                  | Responsibility                                                                           |
|:-------------------|:-------------------------|:-----------------------------------------------------------------------------------------|
| Entry point        | `main.go` + `cmd/`       | Cobra commands, flag parsing, wiring services via `runconfig.New`                        |
| Shared state       | `action/`                | `Action` struct — holds all config values from Viper                                     |
| Config fields      | `field/`                 | String constants for all Viper keys — never use raw strings in `viper.Get*`              |
| Run config         | `runconfig/`             | `RunConfig{*Infrastructure, *Services}` wired by DI at startup                           |
| Registry           | `registrysvc/`           | Fetches module lists from LSP/FAR, partitions folio/eureka, persists to `~/modules.json` |
| Module deploy      | `modulesvc/`             | Pull, deploy, undeploy, readiness; split across `module_svc*.go` files                   |
| Management API     | `managementsvc/`         | mgr-* API calls: tenants, entitlements, applications, discovery                          |
| Keycloak           | `keycloaksvc/`           | Keycloak admin: users, roles, capability sets, tokens                                    |
| Kong               | `kongsvc/`               | Kong route reads + readiness check                                                       |
| UI                 | `uisvc/`                 | UI Docker container deploy + stripes config/package.json generation                      |
| Intercept module   | `interceptmodulesvc/`    | Intercept and redirect module traffic (telepresence-style)                               |
| Upgrade module     | `upgrademodulesvc/`      | Deploy upgraded/downgraded module version + sidecar                                      |
| Consortium         | `consortiumsvc/`         | Consortium/ECS management: entitlements, central ordering                                |
| Tenant             | `tenantsvc/`             | Tenant configuration and parameter management                                            |
| User               | `usersvc/`               | User CRUD operations                                                                     |
| Search             | `searchsvc/`             | Elasticsearch reindex operations                                                         |
| Kafka              | `kafkasvc/`              | Kafka health checks and consumer lag monitoring                                          |
| AWS                | `awssvc/`                | AWS ECR registry authentication                                                          |
| HTTP               | `httpclient/`            | `HTTPClientRunner` interface + retry wrapper                                             |
| Exec               | `execsvc/`               | Shell exec wrapper: `Exec`, `ExecReturnOutput`, `ExecFromDir`                            |
| Docker client      | `dockerclient/`          | Docker client factory + image push/pull helpers                                          |
| Vault client       | `vaultclient/`           | HashiCorp Vault API client                                                               |
| Git client         | `gitclient/`             | Git clone/pull operations                                                                |
| Git repository     | `gitrepository/`         | `GitRepository` data model for git repo metadata                                         |
| Module props       | `moduleprops/`           | Reads `config.*.yaml` into `BackendModule` / `FrontendModule` maps                       |
| Module env         | `moduleenv/`             | Per-module environment variable resolution                                               |
| Models             | `models/`                | Shared data types: `BackendModule`, `Container`, `ProxyModule`, etc.                     |
| Helpers            | `helpers/`               | Pure utility functions (memory conversion, restart policy, etc.)                         |
| Errors             | `errors/`                | Typed error factory functions — all call sites import this, never `fmt.Errorf`           |
| Constants          | `constant/`              | Package-level constants (timeouts, patterns, network IDs)                                |
| Test helpers       | `internal/testhelpers/`  | Shared mocks and HTTP/file helpers — test use only                                       |

## Module source of truth

Module versions come from **LSP + FAR**, not install JSON files.

- **LSP URL** (`lsp.url`): `https://raw.githubusercontent.com/folio-org/platform-lsp/refs/heads/snapshot/platform-descriptor.json`
- **FAR URL** (`far.url`): `https://far.ci.folio.org` → endpoint `/applications?query=id=={name}-{version}`
- **LSP JSON schema**: `{ "eureka-components": [{name, version}], "applications": { "required": [{name, version}], "optional": [{name, version}] } }`
- `registry.url` is still used by `listModuleVersions` and `GetModuleURL` — do not remove it.

## Module list persistence

The flattened module list is persisted as `~/modules.json` (`constant.ModulesFile`).

| Caller            | Call                         | Behaviour                                          |
| ----------------- | ---------------------------- | -------------------------------------------------- |
| `deployModules`   | `GetModules(true, true)`     | Always fetches from LSP+FAR, overwrites file       |
| `deployManagement`| `GetModules(true, true)`     | Always fetches from LSP+FAR, overwrites file       |
| `interceptModule` | `GetModules(false, false)`   | Uses file if present; fetches+creates if missing   |
| `upgradeModule`   | `GetModules(false, false)`   | Uses file if present; fetches+creates if missing   |

`--skipRegistry` flag forces read from file only — hard error if absent. Available on `deployModules`, `deployManagement`, `interceptModule`, `upgradeModule`.

## isEurekaModule predicate

```go
strings.HasSuffix(name, "-keycloak") ||
strings.HasPrefix(name, "mgr-") ||
name == "folio-kong" ||
name == "folio-module-sidecar" ||
name == "mod-scheduler"
```

Everything else is a FOLIO module.

## `Action` struct fields (relevant subset)

```go
ConfigProfileName  string  // viper: profile.name  — container name prefix for non-mgr modules
ConfigLspURL       string  // viper: lsp.url
ConfigFarURL       string  // viper: far.url
ConfigRegistryURL  string  // viper: registry.url  — keep for listModuleVersions
```

## `field/` package

All Viper keys are constants in `field/`. Never use raw string literals in `viper.GetString()` or `viper.Get()`.

```go
viper.GetString(field.ProfileName)     // "profile.name"
viper.GetString(field.LspURL)          // "lsp.url"
viper.GetString(field.FarURL)          // "far.url"
viper.GetString(field.RegistryURL)     // "registry.url"
viper.Get(field.BackendModules)        // "backend-modules"
viper.Get(field.FrontendModules)       // "frontend-modules"
```

Key groups: `Profile`, `Application`, `Lsp`, `Far`, `Registry`, `Tenants`, `Users`, `Roles`, `Consortiums`, `BackendModules`, `FrontendModules`, `SidecarModule`, `ExtraVolumes`, `TemplateEnv`.

## FAR response shape

`ApplicationsResponse.ApplicationDescriptors` is `[]map[string]any`.
Each descriptor has a `"modules"` key whose value is `[]any` (after JSON unmarshal).
`helpers.GetAnySlice(descriptor, "modules")` returns `[]any` — type-assert each element to `map[string]any`.

## Config YAML files

All 10 `config.*.yaml` files must have `lsp.url` and `far.url` keys.
`install.folio` and `install.eureka` keys have been removed and must not be re-added.

## Concurrency pattern (getFlattenedModuleVersions)

```go
results := make([]result, len(applications))
var wg sync.WaitGroup
wg.Add(len(applications))
for idx, app := range applications {
    go func(i int, a PlatformApplication) {
        defer wg.Done()
        // fetch FAR, write to results[i]
    }(idx, app)
}
wg.Wait()
// iterate results, return on first error (fail-fast)
```

Pre-allocated slice with per-goroutine index — no mutex needed.

## Testing conventions

- Mock framework: `testify/mock`
- `testhelpers.NewMockAction()` returns `*action.Action` (real struct, not mock)
- Set `act.ConfigLspURL` and `act.ConfigFarURL` directly before passing to `registrysvc.New()`
- `stubLSP` / `stubFAR` helpers in `registrysvc/registry_svc_test.go` for LSP/FAR stubs
- FAR module slices must be `[]any{map[string]any{...}, ...}` — not `[]map[string]any` — because `GetAnySlice` does `rawValue.([]any)`
- `MockRegistrySvc` in `cmd/cmd_test.go` and `internal/testhelpers/mocks.go` must implement `RegistryProcessor` interface exactly: `GetModules(verbose bool, forceRefresh bool)`
- Persistence tests write/read `~/modules.json` directly; use `t.Cleanup(func() { _ = os.Remove(filePath) })` — note the `_ =` to satisfy errcheck
- Do NOT run `go test ./...` (RAM constrained) — target only affected packages
- Run `golangci-lint run ./...` once as a finishing move, not after every edit

### Two exec mock types — do not confuse

`execsvc.CommandRunner` has three methods: `Exec`, `ExecReturnOutput`, `ExecFromDir`.

| Type                              | Location                            | Use                                                      |
|:----------------------------------|:------------------------------------|:---------------------------------------------------------|
| `testhelpers.MockCommandExecutor` | `internal/testhelpers/mocks.go`     | Service package tests (`uisvc/`, `dockerclient/`, etc.)  |
| `MockExecSvc`                     | `cmd/cmd_test.go`                   | `cmd/` tests only                                        |

`MockCommandExecutor` also has `ExecIgnoreError` as an extra method not present in `CommandRunner` — ignore it when mocking the interface.

## Interface

```go
type RegistryProcessor interface {
    GetNamespace(version string) string
    GetModules(verbose bool, forceRefresh bool) (*models.ProxyModulesByRegistry, error)
    ExtractModuleMetadata(modules *models.ProxyModulesByRegistry)
    GetAuthorizationToken() (string, error)
}
```

## Directories to skip

Do not read or index:

- `tmp/`
- `bin/`
- `docs/` (generated)
