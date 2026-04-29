package uisvc

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/field"
	"github.com/go-git/go-git/v5/plumbing"
)

// UIStripesConfigProcessor defines the interface for UI Stripes configuration operations
type UIStripesConfigProcessor interface {
	GetStripesBranch() plumbing.ReferenceName
	PrepareStripesConfigJS(tenantName string, configPath string) error
	PrepareStripesModulesJS(outputDir string) error
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

func (us *UISvc) PrepareStripesConfigJS(tenantName string, configPath string) error {
	stripesConfigJSFilePath := filepath.Join(configPath, "stripes.config.js")
	readFileBytes, err := os.ReadFile(stripesConfigJSFilePath)
	if err != nil {
		return err
	}

	clientIdSuffix := action.GetConfigEnv("KC_LOGIN_CLIENT_SUFFIX", us.Action.ConfigGlobalEnv)
	tenantOptions := fmt.Sprintf(`{%[1]s: {name: "%[1]s", displayName: "%[1]s", clientId: "%[1]s%s"}}`, tenantName, clientIdSuffix)
	replaceMap := map[string]string{
		"${kongUrl}":           constant.KongExternalHTTP,
		"${tenantUrl}":         us.Action.Param.PlatformLspURL,
		"${keycloakUrl}":       constant.KeycloakExternalHTTP,
		"${hasAllPerms}":       `false`,
		"${isSingleTenant}":    strconv.FormatBool(us.Action.Param.SingleTenant),
		"${tenantOptions}":     tenantOptions,
		"${enableEcsRequests}": strconv.FormatBool(us.Action.Param.EnableECSRequests),
		"${aboutInstallDate}":  fmt.Sprintf("'%s'", time.Now().Format("January 02, 2006")),
		"${aboutInstallMsg}":   fmt.Sprintf("'%s'", "Local build"),
	}

	var newReadFileStr = string(readFileBytes)
	for key, value := range replaceMap {
		if !strings.Contains(newReadFileStr, key) {
			slog.Info(us.Action.Name, "text", "Key not found in stripes.config.js", "key", key)
			continue
		}
		newReadFileStr = strings.ReplaceAll(newReadFileStr, key, value)
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
