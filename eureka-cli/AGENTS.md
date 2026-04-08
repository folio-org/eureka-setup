# Copilot Agent Instructions — eureka-cli

## Project overview

Go CLI tool (Cobra + Viper) that orchestrates local FOLIO/Eureka deployments via Docker.
Module: `github.com/folio-org/eureka-setup/eureka-cli`

## Key architecture

| Layer         | Package                 | Responsibility                                                               |
| ------------- | ----------------------- | -----------------------------------------------------------------------------|
| Entry point   | `main.go` + `cmd/`      | Cobra commands, flag parsing, wiring services                                |
| Shared state  | `action/`               | `Action` struct — holds all config values from Viper                         |
| Config fields | `field/`                | String constants for Viper keys (`lsp.url`, `far.url`, `registry.url`, …)    |
| Registry      | `registrysvc/`          | Fetches module lists from LSP/FAR, partitions folio/eureka                   |
| Models        | `models/`               | Shared structs — `ProxyModule`, `ApplicationModule`, `PlatformDescriptor`, … |
| HTTP          | `httpclient/`           | `HTTPClientRunner` interface + retry wrapper                                 |
| Test helpers  | `internal/testhelpers/` | `MockHTTPClient`, `MockRegistrySvc`, `NewMockAction()`                       |

## Module source of truth

Module versions come from **LSP + FAR**, not install JSON files.

- **LSP URL** (`lsp.url`): `https://raw.githubusercontent.com/folio-org/platform-lsp/refs/heads/snapshot/platform-descriptor.json`
- **FAR URL** (`far.url`): `https://far.ci.folio.org` → endpoint `/applications?query=id=={name}-{version}`
- **LSP JSON schema**: `{ "eureka-components": [{name, version}], "applications": { "required": [{name, version}], "optional": [{name, version}] } }`
- `registry.url` is still used by `listModuleVersions` and `GetModuleURL` — do not remove it.

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
ConfigLspURL      string  // viper: lsp.url
ConfigFarURL      string  // viper: far.url
ConfigRegistryURL string  // viper: registry.url  — keep for listModuleVersions
```

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
- `MockRegistrySvc` in `cmd/cmd_test.go` and `internal/testhelpers/mocks.go` must implement `RegistryProcessor` interface exactly

## Interface

```go
type RegistryProcessor interface {
    GetNamespace(version string) string
    GetModules(verbose bool) (*models.ProxyModulesByRegistry, error)
    ExtractModuleMetadata(modules *models.ProxyModulesByRegistry)
    GetAuthorizationToken() (string, error)
}
```

## Directories to skip

Do not read or index:

- `tmp/`
- `bin/`
- `docs/` (generated)
