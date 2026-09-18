package dockerclient

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/folio-org/eureka-setup/eureka-cli/internal/testhelpers"
	"github.com/moby/moby/api/types/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	kintImage = "docker.libsdev.k-int.com/knowledgeintegration/mod-ill:1.0.0"
	kintHost  = "docker.libsdev.k-int.com"
	hubImage  = "folioci/mod-users:19.4.0-SNAPSHOT.2086"
	notFound  = credentialsNotFound + "\n"
)

func decodeRegistryAuth(t *testing.T, encoded string) registry.AuthConfig {
	t.Helper()
	payload, err := base64.URLEncoding.DecodeString(encoded)
	require.NoError(t, err)
	var authConfig registry.AuthConfig
	require.NoError(t, json.Unmarshal(payload, &authConfig))
	return authConfig
}

func helperCommand(helper, expectedStdin string) any {
	return mock.MatchedBy(func(cmd *exec.Cmd) bool {
		if filepath.Base(cmd.Path) != credentialHelperPrefix+helper || len(cmd.Args) != 2 || cmd.Args[1] != "get" {
			return false
		}
		stdin, err := io.ReadAll(cmd.Stdin)
		return err == nil && string(stdin) == expectedStdin
	})
}

func basicAuth(username, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}

func TestRegistryHost_TableDrivenTests(t *testing.T) {
	tests := []struct {
		name      string
		imageName string
		expected  string
		wantErr   bool
	}{
		{name: "docker hub namespace image", imageName: hubImage, expected: dockerHubDomain},
		{name: "docker hub official image", imageName: "postgres:16", expected: dockerHubDomain},
		{name: "custom registry", imageName: kintImage, expected: kintHost},
		{name: "registry with port", imageName: "localhost:5000/mod-users:1.0.0", expected: "localhost:5000"},
		{name: "ecr", imageName: "123456789012.dkr.ecr.us-east-1.amazonaws.com/folio/mod-users:1.0.0", expected: "123456789012.dkr.ecr.us-east-1.amazonaws.com"},
		{name: "invalid reference", imageName: "MOD-USERS:1.0.0", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, err := registryHost(tt.imageName)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, host)
		})
	}
}

func TestAuthsEntry_TableDrivenTests(t *testing.T) {
	auths := map[string]dockerAuthEntry{
		"docker.libsdev.k-int.com":                 {Username: "exact"},
		"https://docker.libsdev.k-int.com/v2/":     {Username: "legacy"},
		"https://ghcr.io":                          {Username: "ghcr-legacy"},
		"https://index.docker.io/v1/":              {Username: "hub"},
		"https://index.docker.io/v1/access-token":  {Username: "hub-access-token"},
		"https://index.docker.io/v1/refresh-token": {Username: "hub-refresh-token"},
		"localhost:5000":                           {Username: "local"},
	}
	tests := []struct {
		name          string
		serverAddress string
		expected      string
		found         bool
	}{
		{name: "exact key wins over legacy key", serverAddress: kintHost, expected: "exact", found: true},
		{name: "legacy key with scheme", serverAddress: "ghcr.io", expected: "ghcr-legacy", found: true},
		{name: "docker hub index server", serverAddress: dockerHubIndexServer, expected: "hub", found: true},
		{name: "registry with port", serverAddress: "localhost:5000", expected: "local", found: true},
		{name: "unknown host", serverAddress: "quay.io", found: false},
		{name: "docker hub oauth entries do not match the domain", serverAddress: dockerHubDomain, found: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, found := authsEntry(auths, tt.serverAddress)
			assert.Equal(t, tt.found, found)
			assert.Equal(t, tt.expected, entry.Username)
		})
	}
}

func TestAuthsEntry_LegacyKeysAreVisitedDeterministically(t *testing.T) {
	auths := map[string]dockerAuthEntry{
		"https://ghcr.io/v2/": {Username: "second"},
		"https://ghcr.io":     {Username: "first"},
	}

	entry, found := authsEntry(auths, "ghcr.io")
	assert.True(t, found)
	assert.Equal(t, "first", entry.Username)
}

func TestGetRegistryAuth_TableDrivenTests(t *testing.T) {
	kintAuth := basicAuth("kint-user", "s3cret")
	tests := []struct {
		name      string
		config    string // content of config.json; empty means no file
		imageName string
		helper    string // expected credential helper; empty means the helper must not run
		stdin     string
		stdout    string
		stderr    string
		execErr   error
		expectErr bool
		expected  registry.AuthConfig // zero value means an anonymous pull (empty auth)
	}{
		{name: "no config file", imageName: kintImage},
		{name: "empty config file is treated like docker/cli does, as no credentials", config: " ", imageName: kintImage},
		{name: "malformed config file", config: `{"auths": `, imageName: kintImage, expectErr: true},
		{name: "invalid image reference", config: `{}`, imageName: "Docker.Libsdev/MOD-ILL:1.0.0", expectErr: true},
		{name: "no auths entry for host", config: `{"auths": {"ghcr.io": {"auth": "dXNlcjpwYXNz"}}}`, imageName: kintImage},
		{name: "static auths entry with scheme", config: `{"auths": {"https://docker.libsdev.k-int.com": {"auth": "` + kintAuth + `"}}}`, imageName: kintImage,
			expected: registry.AuthConfig{Username: "kint-user", Password: "s3cret", ServerAddress: kintHost}},
		{name: "docker hub web login keeps the registry entry and ignores the oauth entries", imageName: hubImage,
			config:   `{"auths": {"https://index.docker.io/v1/access-token": {"auth": "` + basicAuth("<token>", "access") + `"}, "https://index.docker.io/v1/": {"auth": "` + basicAuth("hub-user", "hub-pass") + `"}, "https://index.docker.io/v1/refresh-token": {"auth": "` + basicAuth("<token>", "refresh") + `"}}}`,
			expected: registry.AuthConfig{Username: "hub-user", Password: "hub-pass", ServerAddress: dockerHubIndexServer}},
		{name: "docker hub oauth entries alone give no credentials", imageName: hubImage,
			config: `{"auths": {"https://index.docker.io/v1/access-token": {"auth": "` + basicAuth("<token>", "access") + `"}}}`},
		{name: "static auths entry without credentials", config: `{"auths": {"docker.libsdev.k-int.com": {}}}`, imageName: kintImage},
		{name: "static auths entry with invalid base64", config: `{"auths": {"docker.libsdev.k-int.com": {"auth": "%%%"}}}`, imageName: kintImage, expectErr: true},
		{name: "credsStore", config: `{"credsStore": "secretservice"}`, imageName: kintImage,
			helper: "secretservice", stdin: kintHost, stdout: `{"ServerURL":"docker.libsdev.k-int.com","Username":"kint-user","Secret":"s3cret"}` + "\n",
			expected: registry.AuthConfig{Username: "kint-user", Password: "s3cret", ServerAddress: kintHost}},
		{name: "credsStore asks for the docker hub index server", config: `{"credsStore": "desktop"}`, imageName: hubImage,
			helper: "desktop", stdin: dockerHubIndexServer, stdout: `{"ServerURL":"https://index.docker.io/v1/","Username":"hub-user","Secret":"hub-pass"}`,
			expected: registry.AuthConfig{Username: "hub-user", Password: "hub-pass", ServerAddress: dockerHubIndexServer}},
		{name: "credHelpers takes precedence over credsStore", config: `{"credsStore": "secretservice", "credHelpers": {"docker.libsdev.k-int.com": "pass"}}`, imageName: kintImage,
			helper: "pass", stdin: kintHost, stdout: `{"Username":"kint-user","Secret":"s3cret"}`,
			expected: registry.AuthConfig{Username: "kint-user", Password: "s3cret", ServerAddress: kintHost}},
		{name: "credHelpers keyed by the docker hub index server", config: `{"credHelpers": {"https://index.docker.io/v1/": "desktop"}}`, imageName: hubImage,
			helper: "desktop", stdin: dockerHubIndexServer, stdout: `{"ServerURL":"https://index.docker.io/v1/","Username":"hub-user","Secret":"hub-pass"}`,
			expected: registry.AuthConfig{Username: "hub-user", Password: "hub-pass", ServerAddress: dockerHubIndexServer}},
		{name: "helper identity token", config: `{"credsStore": "desktop"}`, imageName: kintImage,
			helper: "desktop", stdin: kintHost, stdout: `{"Username":"<token>","Secret":"identity-token"}`,
			expected: registry.AuthConfig{IdentityToken: "identity-token", ServerAddress: kintHost}},
		{name: "helper miss is final, the auths entry is not a fallback", config: `{"auths": {"docker.libsdev.k-int.com": {"auth": "` + basicAuth("file-user", "file-pass") + `"}}, "credsStore": "secretservice"}`, imageName: kintImage,
			helper: "secretservice", stdin: kintHost, stdout: notFound, execErr: errors.New("exit status 1")},
		{name: "helper failure mentioning not found is an error, not a miss", config: `{"credsStore": "pass"}`, imageName: kintImage,
			helper: "pass", stdin: kintHost, stdout: `pass not initialized: exec: "pass": executable file not found in $PATH` + "\n", execErr: errors.New("exit status 1"), expectErr: true},
		{name: "helper execution error", config: `{"credsStore": "missing"}`, imageName: kintImage,
			helper: "missing", stdin: kintHost, execErr: errors.New(`exec: "docker-credential-missing": executable file not found in $PATH`), expectErr: true},
		{name: "helper malformed output", config: `{"credsStore": "secretservice"}`, imageName: kintImage,
			helper: "secretservice", stdin: kintHost, stdout: "not json", expectErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			configDir := t.TempDir()
			if tt.config != "" {
				testhelpers.CreateFileInDir(t, configDir, dockerConfigFile, tt.config)
			}
			t.Setenv(dockerConfigEnv, configDir)
			mockExec := new(testhelpers.MockCommandExecutor)
			if tt.helper != "" {
				mockExec.On("ExecReturnOutput", helperCommand(tt.helper, tt.stdin)).
					Return(*bytes.NewBufferString(tt.stdout), *bytes.NewBufferString(tt.stderr), tt.execErr)
			}
			client := New(testhelpers.NewMockAction(), mockExec)

			// Act
			auth, err := client.GetRegistryAuth(tt.imageName)

			// Assert
			if tt.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.expected == (registry.AuthConfig{}) {
				assert.Empty(t, auth)
			} else {
				assert.Equal(t, tt.expected, decodeRegistryAuth(t, auth))
			}
			mockExec.AssertExpectations(t)
		})
	}
}
