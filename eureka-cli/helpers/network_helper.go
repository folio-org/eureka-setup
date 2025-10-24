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

func GetGatewayURL(actionName string) (string, error) {
	protoAndBaseURL := GetProtoAndBaseURL(actionName)
	if protoAndBaseURL == "" {
		return "", fmt.Errorf("cannot construct getaway url for %s platform", runtime.GOOS)
	}

	return protoAndBaseURL + ":%s", nil
}

func GetProtoAndBaseURL(action string) string {
	if viper.IsSet(field.ApplicationGatewayHostname) {
		return viper.GetString(field.ApplicationGatewayHostname)
	} else if HostnameExists(action, constant.DockerHostname) {
		return fmt.Sprintf("http://%s", constant.DockerHostname)
	} else if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		return fmt.Sprintf("http://%s", constant.DockerGatewayIP)
	}

	return ""
}

func SetFreePortFromRange(action *action.Action) (int, error) {
	for port := action.StartPort; port <= action.EndPort; port++ {
		if !slices.Contains(action.ReservedPorts, port) && IsPortFree(action, action.StartPort, action.EndPort, port) {
			action.ReservedPorts = append(action.ReservedPorts, port)
			return port, nil
		}
	}

	return 0, fmt.Errorf("cannot find free TCP ports in range %d-%d", action.StartPort, action.EndPort)
}

func IsPortFree(action *action.Action, portStart, portEnd int, port int) bool {
	tcpListen, err := net.Listen("tcp", fmt.Sprintf(":%s", strconv.Itoa(port)))
	if err != nil {
		slog.Debug(action.Name, "text", "TCP port is reserved or already bound in range", "port", port, "portStart", portStart, "portEnd", portEnd)
		return false
	}
	defer CloseListener(tcpListen)

	return true
}

func HostnameExists(actionName string, hostname string) bool {
	_, err := net.LookupHost(hostname)
	if err != nil {
		slog.Debug(actionName, "text", "Host is unreachable with error", "hostname", hostname, "error", err.Error())
	}

	return err == nil
}

func ConstructURL(url string, schemaAndBaseURL string) string {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}

	return fmt.Sprintf("%s:%s", schemaAndBaseURL, url)
}

func ExtractPortFromURL(url string) (int, error) {
	sidecarServer, err := GetPortFromURL(url)
	if err != nil {
		return 0, err
	}

	return sidecarServer, nil
}

func CloseListener(listener net.Listener) {
	_ = listener.Close()
}
