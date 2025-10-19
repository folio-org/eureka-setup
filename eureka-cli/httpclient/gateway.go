package httpclient

import (
	"fmt"
	"runtime"

	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/spf13/viper"
)

func (hc *HTTPClient) GetGatewayURL() string {
	protoAndBaseURL := hc.GetGatewayProtoAndBaseURL()
	if protoAndBaseURL == "" {
		helpers.LogErrorPanic(hc.Action, fmt.Errorf("cannot construct getaway url for %s platform", runtime.GOOS))
		return ""
	}

	return protoAndBaseURL + ":%d%s"
}

func (hc *HTTPClient) GetGatewayProtoAndBaseURL() string {
	if viper.IsSet(field.ApplicationGatewayHostname) {
		return viper.GetString(field.ApplicationGatewayHostname)
	} else if helpers.HostnameExists(hc.Action, constant.DefaultDockerHostname) {
		return fmt.Sprintf("http://%s", constant.DefaultDockerHostname)
	} else if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		return fmt.Sprintf("http://%s", constant.DefaultDockerGatewayIP)
	}

	return ""
}
