package uistep

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/runparams"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/viper"
)

func (us *UIStep) GetStripesBranch() plumbing.ReferenceName {
	if viper.IsSet(field.ApplicationStripesBranch) {
		branchStr := viper.GetString(field.ApplicationStripesBranch)
		stripesBranch := plumbing.ReferenceName(branchStr)

		slog.Info(us.Action.Name, "text", fmt.Sprintf("Found stripes branch in config: %s", stripesBranch))

		return stripesBranch
	}
	slog.Info(us.Action.Name, "text", fmt.Sprintf("No stripes branch is defined in config, using default branch: %s", constant.StripesBranch))

	return constant.StripesBranch
}

func (us *UIStep) PrepareStripesConfigJS(rp *runparams.RunParams, configPath string, tenant string) error {
	stripesConfigJSFilePath := fmt.Sprintf("%s/stripes.config.js", configPath)

	readFileBytes, err := os.ReadFile(stripesConfigJSFilePath)
	if err != nil {
		return err
	}

	replaceMap := map[string]string{
		"${kongUrl}":           constant.KongExternalHTTP,
		"${tenantUrl}":         rp.PlatformCompleteURL,
		"${keycloakUrl}":       constant.KeycloakExternalHTTP,
		"${hasAllPerms}":       `false`,
		"${isSingleTenant}":    strconv.FormatBool(rp.SingleTenant),
		"${tenantOptions}":     fmt.Sprintf(`{%[1]s: {name: "%[1]s", clientId: "%[1]s%s"}}`, tenant, helpers.GetConfigEnv("KC_LOGIN_CLIENT_SUFFIX")),
		"${enableEcsRequests}": strconv.FormatBool(rp.EnableECSRequests),
	}

	var newReadFileStr = string(readFileBytes)
	for key, value := range replaceMap {
		if !strings.Contains(newReadFileStr, key) {
			slog.Info(us.Action.Name, "text", fmt.Sprintf("Key not found in stripes.config.js: %s", key))
			continue
		}

		newReadFileStr = strings.ReplaceAll(newReadFileStr, key, value)
	}

	newReadFileStr = strings.ReplaceAll(newReadFileStr, "'@folio/users' : {}", "'@folio/users' : {},\n    '@folio/consortia-settings' : {}")

	fmt.Println()
	fmt.Println("### Dumping stripes.config.js ###")
	fmt.Println(newReadFileStr)
	fmt.Println()

	err = os.WriteFile(stripesConfigJSFilePath, []byte(newReadFileStr), 0)
	if err != nil {
		return err
	}

	return nil
}
