package dockerclient

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/distribution/reference"
	"github.com/folio-org/eureka-setup/eureka-cli/errors"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/moby/moby/api/types/registry"
)

const (
	dockerConfigEnv        = "DOCKER_CONFIG"
	dockerConfigDir        = ".docker"
	dockerConfigFile       = "config.json"
	dockerHubDomain        = "docker.io"
	dockerHubIndexServer   = "https://index.docker.io/v1/"
	credentialHelperPrefix = "docker-credential-"
	tokenUsername          = "<token>"
	// credentialsNotFound is the message docker-credential-helpers prints for a missing entry
	credentialsNotFound = "credentials not found in native keychain"
)

// dockerConfig is the subset of ~/.docker/config.json relevant for registry credentials
type dockerConfig struct {
	Auths       map[string]dockerAuthEntry `json:"auths"`
	CredsStore  string                     `json:"credsStore"`
	CredHelpers map[string]string          `json:"credHelpers"`
}

type dockerAuthEntry struct {
	Auth          string `json:"auth"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	IdentityToken string `json:"identitytoken"`
}

// credentialHelperResponse is the part of the `docker-credential-<helper> get` output that is used
type credentialHelperResponse struct {
	Username string
	Secret   string
}

// GetRegistryAuth resolves the credentials stored by `docker login` for the registry that hosts imageName
// and returns them in the encoded form the Docker API expects as RegistryAuth. Source selection follows the
// Docker CLI: the credential helper bound to the registry (credHelpers, keyed by the host or, for Docker Hub,
// by https://index.docker.io/v1/), else the default credential store (credsStore), and only when neither is
// configured the static auths entry; a configured helper's answer is final, a miss is not a fallback to auths.
// An empty string means no credentials were found, i.e. an anonymous pull.
func (dc *DockerClient) GetRegistryAuth(imageName string) (string, error) {
	host, err := registryHost(imageName)
	if err != nil {
		return "", err
	}

	config, err := readDockerConfig()
	if err != nil {
		return "", err
	}

	authConfig, err := dc.resolveAuthConfig(config, authConfigKey(host))
	if err != nil {
		return "", err
	}
	if authConfig == nil {
		return "", nil
	}
	slog.Info(dc.Action.Name, "text", "Using Docker credentials for image registry", "host", host, "username", authConfig.Username)

	payload, err := json.Marshal(authConfig)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(payload), nil
}

// authConfigKey is the key under which `docker login` stores a registry: its host, or the index server for Docker Hub
func authConfigKey(host string) string {
	if host == dockerHubDomain {
		return dockerHubIndexServer
	}

	return host
}

func (dc *DockerClient) resolveAuthConfig(config dockerConfig, serverAddress string) (*registry.AuthConfig, error) {
	helper := config.CredHelpers[serverAddress]
	if helper == "" {
		helper = config.CredsStore
	}
	if helper != "" {
		return dc.authConfigFromHelper(helper, serverAddress)
	}

	entry, found := authsEntry(config.Auths, serverAddress)
	if !found {
		return nil, nil
	}

	return authConfigFromEntry(entry, serverAddress)
}

func (dc *DockerClient) authConfigFromHelper(helper, serverAddress string) (*registry.AuthConfig, error) {
	cmd := exec.Command(credentialHelperPrefix+helper, "get")
	cmd.Stdin = strings.NewReader(serverAddress)
	stdout, _, err := dc.ExecSvc.ExecReturnOutput(cmd)
	if err != nil {
		// Helpers built on docker-credential-helpers report a missing entry as a non-zero exit with exactly this
		// message on stdout, which is what the Docker CLI checks; anything else is a real failure
		if strings.TrimSpace(stdout.String()) == credentialsNotFound {
			return nil, nil
		}
		return nil, errors.RegistryAuthResolveFailed(serverAddress, helper, err)
	}

	var response credentialHelperResponse
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &response); err != nil {
		return nil, errors.RegistryAuthResolveFailed(serverAddress, helper, err)
	}
	if response.Username == "" || response.Secret == "" {
		return nil, nil
	}
	if response.Username == tokenUsername {
		return &registry.AuthConfig{IdentityToken: response.Secret, ServerAddress: serverAddress}, nil
	}

	return &registry.AuthConfig{Username: response.Username, Password: response.Secret, ServerAddress: serverAddress}, nil
}

func authConfigFromEntry(entry dockerAuthEntry, serverAddress string) (*registry.AuthConfig, error) {
	authConfig := &registry.AuthConfig{Username: entry.Username, Password: entry.Password, IdentityToken: entry.IdentityToken, ServerAddress: serverAddress}
	if entry.Auth != "" {
		decoded, err := base64.StdEncoding.DecodeString(entry.Auth)
		if err != nil {
			return nil, errors.DockerConfigInvalid("auths entry for "+serverAddress, err)
		}
		username, password, found := strings.Cut(string(decoded), ":")
		if found {
			authConfig.Username = username
			authConfig.Password = strings.TrimSuffix(password, "\x00")
		}
	}
	if authConfig.IdentityToken == "" && (authConfig.Username == "" || authConfig.Password == "") {
		return nil, nil
	}

	return authConfig, nil
}

// registryHost returns the registry domain of an image reference, e.g. docker.io for folioci/mod-users:1.0.0
// and docker.libsdev.k-int.com for docker.libsdev.k-int.com/knowledgeintegration/mod-ill:1.0.0
func registryHost(imageName string) (string, error) {
	named, err := reference.ParseNormalizedNamed(imageName)
	if err != nil {
		return "", errors.InvalidImageReference(imageName, err)
	}

	return reference.Domain(named), nil
}

// authsEntry looks up the auths entry for a server address the way the Docker CLI does: the exact key first,
// then legacy keys that carry a scheme or path (e.g. https://docker.libsdev.k-int.com/v2/) by their host.
// Docker Hub is stored under https://index.docker.io/v1/, so its OAuth entries (.../access-token,
// .../refresh-token) never match. Legacy keys are visited in sorted order to keep the choice deterministic.
func authsEntry(auths map[string]dockerAuthEntry, serverAddress string) (dockerAuthEntry, bool) {
	if entry, found := auths[serverAddress]; found {
		return entry, true
	}
	for _, key := range slices.Sorted(maps.Keys(auths)) {
		if normalizeHost(key) == serverAddress {
			return auths[key], true
		}
	}

	return dockerAuthEntry{}, false
}

func normalizeHost(key string) string {
	key = strings.TrimPrefix(key, "https://")
	key = strings.TrimPrefix(key, "http://")
	if idx := strings.Index(key, "/"); idx != -1 {
		key = key[:idx]
	}

	return key
}

// readDockerConfig reads $DOCKER_CONFIG/config.json or ~/.docker/config.json; a missing file yields an empty config
func readDockerConfig() (dockerConfig, error) {
	configDir := os.Getenv(dockerConfigEnv)
	if configDir == "" {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return dockerConfig{}, err
		}
		configDir = filepath.Join(userHome, dockerConfigDir)
	}
	configPath := filepath.Join(configDir, dockerConfigFile)

	var config dockerConfig
	err := helpers.ReadJSONFromFile(configPath, &config)
	if err != nil && !os.IsNotExist(err) {
		return dockerConfig{}, errors.DockerConfigInvalid(configPath, err)
	}

	return config, nil
}
