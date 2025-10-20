package helpers

import (
	"fmt"
	"log/slog"
	"net"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/spf13/viper"
)

func GetGatewayURL(action *action.Action) string {
	protoAndBaseURL := GetGatewayProtoAndBaseURL(action)
	if protoAndBaseURL == "" {
		LogErrorPanic(action, fmt.Errorf("cannot construct getaway url for %s platform", runtime.GOOS))
		return ""
	}

	return protoAndBaseURL + ":%s%s"
}

func GetGatewayProtoAndBaseURL(action *action.Action) string {
	if viper.IsSet(field.ApplicationGatewayHostname) {
		return viper.GetString(field.ApplicationGatewayHostname)
	} else if HostnameExists(action, constant.DockerHostname) {
		return fmt.Sprintf("http://%s", constant.DockerHostname)
	} else if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		return fmt.Sprintf("http://%s", constant.DockerGatewayIP)
	}

	return ""
}

func SetFreePortFromRange(action *action.Action) int {
	for port := action.StartPort; port <= action.EndPort; port++ {
		if !slices.Contains(action.ReservedPorts, port) && IsPortFree(action, action.StartPort, action.EndPort, port) {
			action.ReservedPorts = append(action.ReservedPorts, port)
			return port
		}
	}
	LogErrorPanic(action, fmt.Errorf("cannot find free TCP ports in range %d-%d", action.StartPort, action.EndPort))
	return 0
}

func IsPortFree(action *action.Action, portStart, portEnd int, port int) bool {
	tcpListen, err := net.Listen("tcp", fmt.Sprintf(":%s", strconv.Itoa(port)))
	if err != nil {
		slog.Debug(action.Name, "text", fmt.Sprintf("TCP %d port is reserved or already bound in range %d-%d", port, portStart, portEnd))
		return false
	}
	defer func() {
		_ = tcpListen.Close()
	}()
	return true
}

func HostnameExists(action *action.Action, hostname string) bool {
	_, err := net.LookupHost(hostname)
	if err != nil {
		slog.Debug(action.Name, "text", fmt.Sprintf("host %s is unreachable: %s", hostname, err.Error()))
	}
	return err == nil
}

func ConstructURL(url string, schemaAndBaseURL string) string {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	return fmt.Sprintf("%s:%s", schemaAndBaseURL, url)
}

func ExtractPortFromURL(action *action.Action, url string) int {
	sidecarServer, err := GetPortFromURL(url)
	if err != nil {
		slog.Error(action.Name, "error", err)
		panic(err)
	}
	return sidecarServer
}
