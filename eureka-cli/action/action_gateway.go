package action

import (
	"fmt"
	"log/slog"
	"runtime"
	"strings"

	"github.com/j011195/eureka-setup/eureka-cli/constant"
	"github.com/j011195/eureka-setup/eureka-cli/errors"
	"github.com/j011195/eureka-setup/eureka-cli/field"
	"github.com/j011195/eureka-setup/eureka-cli/helpers"
	"github.com/spf13/viper"
)

func GetGatewayURLTemplate(actionName string) (string, error) {
	gatewayURL, err := GetGatewayURL(actionName)
	if err != nil {
		return "", err
	}

	return gatewayURL + ":%s", nil
}

func GetGatewayURL(actionName string) (string, error) {
	slog.Debug(actionName, "text", "RETRIEVING GATEWAY URL")
	gatewayURL, err := getConfigGatewayURL(actionName)
	if gatewayURL == "" {
		gatewayURL, err = getDefaultGatewayURL(actionName)
	}
	if gatewayURL == "" {
		gatewayURL, err = getOtherGatewayURL(actionName)
	}
	if gatewayURL == "" || err != nil {
		return "", errors.GatewayURLConstructFailed(runtime.GOOS, err)
	}
	slog.Debug(actionName, "text", "Retrieved gateway URL", "url", gatewayURL)

	return gatewayURL, nil
}

func getConfigGatewayURL(actionName string) (gatewayURL string, err error) {
	if !viper.IsSet(field.ApplicationGatewayHostname) {
		return "", nil
	}

	hostname := viper.GetString(field.ApplicationGatewayHostname)
	if err = helpers.IsHostnameReachable(actionName, hostname); err != nil {
		slog.Warn(actionName, "text", "Retrieving config gateway hostname was unsuccessful", "error", err)
		return "", err
	}

	if !strings.HasPrefix(hostname, "http://") {
		gatewayURL = fmt.Sprintf("http://%s", hostname)
	} else {
		gatewayURL = hostname
	}

	return gatewayURL, nil
}

func getDefaultGatewayURL(actionName string) (gatewayURL string, err error) {
	if err = helpers.IsHostnameReachable(actionName, constant.DockerHostname); err != nil {
		slog.Warn(actionName, "text", "Retrieving default gateway URL was unsuccessful", "error", err)
		return "", nil
	}

	return fmt.Sprintf("http://%s", constant.DockerHostname), nil
}

func getOtherGatewayURL(actionName string) (gatewayURL string, err error) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		err = errors.UnsupportedPlatform(runtime.GOOS, constant.DockerGatewayIP)
		slog.Warn(actionName, "text", "Retrieving other gateway URL was unsuccessful", "error", err)
		return "", err
	}

	return fmt.Sprintf("http://%s", constant.DockerGatewayIP), nil
}
