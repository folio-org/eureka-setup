package helpers

import (
	"fmt"
	"log/slog"
	"net"
	"slices"
	"strconv"
	"strings"

	"github.com/folio-org/eureka-cli/action"
)

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
