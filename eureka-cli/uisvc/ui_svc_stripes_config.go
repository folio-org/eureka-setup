package uisvc

import (
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	apperrors "github.com/folio-org/eureka-setup/eureka-cli/errors"
	"github.com/folio-org/eureka-setup/eureka-cli/field"
	"github.com/go-git/go-git/v5/plumbing"
)

// stripesPlaceholderPattern matches a stripes.config.js placeholder such as ${kongUrl} or ${ kongUrl }.
var stripesPlaceholderPattern = regexp.MustCompile(`\$\{([^}]*)\}`)

// UIStripesConfigProcessor defines the interface for UI Stripes configuration operations
type UIStripesConfigProcessor interface {
	GetStripesURL() string
	GetStripesBranch() plumbing.ReferenceName
	GetStripesConfig() string
	PrepareStripesConfigJS(tenantName string, configPath string) error
	PrepareStripesModulesJS(outputDir string) error
}

// GetStripesURL returns the platform repository the UI is built from: application.stripes-url, or the upstream platform-lsp repository
func (us *UISvc) GetStripesURL() string {
	return action.GetStringOrDefault(field.ApplicationStripesURL, constant.PlatformLspRepositoryURL)
}

func (us *UISvc) GetStripesBranch() plumbing.ReferenceName {
	if action.IsSet(field.ApplicationStripesBranch) {
		branchStr := us.Action.ConfigApplicationStripesBranch
		slog.Info(us.Action.Name, "text", "Found stripes branch in config", "branch", branchStr)
		return plumbing.ReferenceName(branchStr)
	}
	slog.Info(us.Action.Name, "text", "Using default branch", "branch", constant.StripesBranch)

	return constant.StripesBranch
}

// GetStripesConfig returns the Stripes config the UI build substitutes and builds: application.stripes-config, or stripes.config.js
func (us *UISvc) GetStripesConfig() string {
	return action.GetStringOrDefault(field.ApplicationStripesConfig, constant.StripesConfigFile)
}

// PrepareStripesConfigJS substitutes the placeholders of the configured Stripes config (application.stripes-config, by default
// stripes.config.js) and writes the result to stripes.config.js, the file the repository's build reads.
func (us *UISvc) PrepareStripesConfigJS(tenantName string, configPath string) error {
	selected := us.GetStripesConfig()
	readFileBytes, err := us.readStripesConfig(configPath, selected)
	if err != nil {
		return err
	}
	if selected != constant.StripesConfigFile {
		slog.Info(us.Action.Name, "text", "Using stripes config", "file", selected)
	}
	stripesConfigJSFilePath := filepath.Join(configPath, constant.StripesConfigFile)

	clientIdSuffix := action.GetConfigEnv("KC_LOGIN_CLIENT_SUFFIX", us.Action.ConfigGlobalEnv)
	tenantOptions := fmt.Sprintf(`{%[1]s: {name: "%[1]s", displayName: "%[1]s", clientId: "%[1]s%s"}}`, tenantName, clientIdSuffix)
	replaceMap := map[string]string{
		"kongUrl":           constant.KongExternalHTTP,
		"tenantUrl":         us.Action.Param.PlatformLspURL,
		"keycloakUrl":       constant.KeycloakExternalHTTP,
		"hasAllPerms":       `false`,
		"isSingleTenant":    strconv.FormatBool(us.Action.Param.SingleTenant),
		"tenantOptions":     tenantOptions,
		"enableEcsRequests": strconv.FormatBool(us.Action.Param.EnableECSRequests),
		"aboutInstallDate":  fmt.Sprintf("'%s'", time.Now().Format("January 02, 2006")),
		"aboutInstallMsg":   "'Local build'",
	}

	replaced := make(map[string]bool)
	unresolved := make(map[string]bool)
	newReadFileStr := stripesPlaceholderPattern.ReplaceAllStringFunc(string(readFileBytes), func(placeholder string) string {
		key := strings.TrimSpace(stripesPlaceholderPattern.FindStringSubmatch(placeholder)[1])
		value, ok := replaceMap[key]
		if !ok {
			unresolved[placeholder] = true
			return placeholder
		}
		replaced[key] = true

		return value
	})
	var missing []string
	for _, key := range slices.Sorted(maps.Keys(replaceMap)) {
		if !replaced[key] {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		slog.Info(us.Action.Name, "text", "Keys not found in stripes.config.js", "keys", missing)
	}
	if len(replaced) == 0 {
		slog.Warn(us.Action.Name, "text", "stripes.config.js contains no substitution placeholders and is used as is, pass -u to restore a pristine checkout if this is unintended")
	}
	if len(unresolved) > 0 {
		return apperrors.StripesConfigPlaceholdersUnresolved(slices.Sorted(maps.Keys(unresolved)))
	}
	fmt.Println()
	fmt.Println("DUMPING stripes.config.js")
	fmt.Println(newReadFileStr)
	fmt.Println()

	err = os.WriteFile(stripesConfigJSFilePath, []byte(newReadFileStr), 0644)
	if err != nil {
		return err
	}

	return nil
}

// readStripesConfig reads the selected Stripes config, a path relative to the repository root
func (us *UISvc) readStripesConfig(configPath string, selected string) ([]byte, error) {
	if !filepath.IsLocal(selected) {
		return nil, apperrors.StripesConfigInvalid(field.ApplicationStripesConfig, selected)
	}
	content, err := os.ReadFile(filepath.Join(configPath, selected))
	if errors.Is(err, os.ErrNotExist) {
		return nil, apperrors.StripesConfigMissing(field.ApplicationStripesConfig, selected, us.GetStripesURL())
	}
	if err != nil {
		return nil, err
	}

	return content, nil
}

func (us *UISvc) PrepareStripesModulesJS(outputDir string) error {
	var modulesToRemove []string
	if us.Action.Param.SingleTenant {
		modulesToRemove = append(modulesToRemove, "@folio/consortia-settings")
	}
	if !us.Action.Param.LinkedData {
		modulesToRemove = append(modulesToRemove, "@folio/ld-folio-wrapper")
	}
	if len(modulesToRemove) == 0 {
		return nil
	}

	filePath := filepath.Join(outputDir, "stripes.modules.js")
	content, err := os.ReadFile(filePath)
	if errors.Is(err, os.ErrNotExist) {
		// A fork may keep its modules inline in its Stripes config; nothing to remove then
		slog.Info(us.Action.Name, "text", "No stripes.modules.js in the repository, optional modules not removed", "modules", modulesToRemove)
		return nil
	}
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		skip := false
		for _, mod := range modulesToRemove {
			if strings.HasPrefix(trimmed, fmt.Sprintf("'%s':", mod)) ||
				strings.HasPrefix(trimmed, fmt.Sprintf(`"%s":`, mod)) {
				skip = true
				slog.Info(us.Action.Name, "text", "Removed module from stripes.modules.js", "module", mod)
				break
			}
		}
		if !skip {
			result = append(result, line)
		}
	}
	finalContent := strings.Join(result, "\n")

	fmt.Println()
	fmt.Println("DUMPING stripes.modules.js")
	fmt.Println(finalContent)
	fmt.Println()

	return os.WriteFile(filePath, []byte(finalContent), 0644)
}
