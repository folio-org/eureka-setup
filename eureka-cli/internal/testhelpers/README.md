# Internal Test Helpers Package

Reusable testing utilities for eureka-cli. Import path: `github.com/folio-org/eureka-setup/eureka-cli/internal/testhelpers`

## Files

| File                | Contents                                                              |
|:--------------------|:----------------------------------------------------------------------|
| `mocks.go`          | `NewMockAction()`, `MockHTTPClient`, `MockCommandExecutor`, `MockRegistrySvc`, `MockModuleEnv`, `MockDockerClient`, `MockTenantSvc` |
| `git_mocks.go`      | `MockGitClient` — mock for `gitclient.GitClientRunner` (`KongRepository`, `KeycloakRepository`, `PlatformCompleteRepository`, `Clone`, `ResetHardPullFromOrigin`) |
| `http_helpers.go`   | `MockHTTPServer`, `JSONResponse`, `ErrorResponse`, `EmptyResponse`, `SequentialResponses`, request assertion helpers |
| `file_helpers.go`   | `CreateTempJSONFile`, `CreateTempFile`, `CreateJSONFileInDir`, `CreateFileInDir`, `ReadFileContent` |
| `viper_helpers.go`  | `ViperTestConfig` — tracks original values and restores them in `Reset()`; `SetupViperForTest(map[string]any)` for bulk setup |
| `doc.go`            | Package doc comment                                                   |

## Key mocks

**`NewMockAction()`** — returns `*action.Action` with name `"test-action"` and empty `Param`. Set fields directly after construction (e.g. `action.ConfigProfileName = "platform-complete"`).

**`MockCommandExecutor`** — implements `execsvc.CommandRunner`: `Exec`, `ExecReturnOutput`, `ExecFromDir`. Also has `ExecIgnoreError` as an extra method not present in the interface — do not stub it unless the code under test calls it directly. When mocking `ExecReturnOutput`, always return `(bytes.Buffer, bytes.Buffer, error)`.

> **Note:** `cmd/cmd_test.go` defines its own `MockExecSvc` (three interface methods only) for use within `cmd/` tests. Do not replace it with `MockCommandExecutor` — they are separate types serving different test scopes.

**`MockHTTPClient`** — implements all `httpclient.HTTPClientRunner` methods. Use `.On("MethodName", mock.Anything, ...).Return(...)`.

**`MockRegistrySvc`** — `GetModules(verbose, forceRefresh bool)` — both args required in every `.On()` call. `forceRefresh=true` exercises network path; `false` exercises local file path.

**`MockDockerClient`** — implements `dockerclient.DockerClientRunner`: `Create`, `Close`, `PushImage`, `ForcePullImage`.

## Running tests

```bash
# Target a single package — machine is RAM-constrained
go test ./cmd/...
go test ./modulesvc/...

# Docker API tests use httptest, no real daemon needed:
# client.NewClientWithOpts(client.WithHost(ts.URL), client.WithVersion("1.41"))

# Lint once as finishing move
golangci-lint run ./...
```

See [TESTING_GUIDE.md](./TESTING_GUIDE.md) for AAA structure, HTTP patterns, table-driven tests, and service-specific notes.
