package uisvc

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/j011195/eureka-setup/eureka-cli/constant"
	"github.com/j011195/eureka-setup/eureka-cli/field"
)

// UIStripesConfigProcessor defines the interface for UI Stripes configuration operations
type UIStripesConfigProcessor interface {
	GetStripesBranch() plumbing.ReferenceName
	PrepareStripesConfigJS(tenantName string, configPath string) error
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
		"${tenantUrl}":         us.Action.Param.PlatformCompleteURL,
		"${keycloakUrl}":       constant.KeycloakExternalHTTP,
		"${hasAllPerms}":       `false`,
		"${isSingleTenant}":    strconv.FormatBool(us.Action.Param.SingleTenant),
		"${tenantOptions}":     tenantOptions,
		"${enableEcsRequests}": strconv.FormatBool(us.Action.Param.EnableECSRequests),
	}

	var newReadFileStr = string(readFileBytes)
	for key, value := range replaceMap {
		if !strings.Contains(newReadFileStr, key) {
			slog.Info(us.Action.Name, "text", "Key not found in stripes.config.js", "key", key)
			continue
		}
		newReadFileStr = strings.ReplaceAll(newReadFileStr, key, value)
	}
	if !us.Action.Param.SingleTenant {
		newReadFileStr = strings.ReplaceAll(newReadFileStr, "'@folio/users' : {}", "'@folio/users' : {},\n    '@folio/consortia-settings' : {}")
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
