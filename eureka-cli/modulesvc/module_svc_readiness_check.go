package modulesvc

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/errors"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
)

const (
	healthPath         = "/admin/health"
	livenessHealthPath = "/admin/health/live"
)

// ModuleReadinessChecker defines the interface for module readiness check operations
type ModuleReadinessChecker interface {
	CheckModuleReadiness(wg *sync.WaitGroup, errCh chan<- error, moduleName string, port int)
}

func (ms *ModuleSvc) CheckModuleReadiness(wg *sync.WaitGroup, errCh chan<- error, moduleName string, port int) {
	ms.checkReadiness(wg, errCh, moduleName, ms.healthURL("", port, healthPath))
}

func (ms *ModuleSvc) CheckModuleReadinessByURL(wg *sync.WaitGroup, errCh chan<- error, moduleName string, baseURL string) {
	ms.checkReadiness(wg, errCh, moduleName, ms.healthURL(baseURL, 0, healthPath))
}

func (ms *ModuleSvc) healthURL(baseURL string, port int, path string) string {
	if baseURL != "" {
		return strings.TrimRight(baseURL, "/") + path
	}

	return ms.Action.GetRequestURL(strconv.Itoa(port), path)
}

func (ms *ModuleSvc) checkReadiness(wg *sync.WaitGroup, errCh chan<- error, moduleName string, requestURL string) {
	defer wg.Done()

	if err := ms.waitUntilHealthy(moduleName, requestURL); err != nil {
		select {
		case errCh <- err:
		default:
		}
	}
}

func (ms *ModuleSvc) waitUntilHealthy(moduleName string, requestURL string) error {
	slog.Info(ms.Action.Name, "text", "Preparing module readiness check", "module", moduleName, "url", requestURL)
	maxRetries := helpers.DefaultInt(ms.ReadinessMaxRetries, constant.ModuleReadinessMaxRetries)
	waitDuration := helpers.DefaultDuration(ms.ReadinessWait, constant.ModuleReadinessWait)
	for retryCount := range maxRetries {
		statusCode, _ := ms.HTTPClient.Ping(requestURL)
		if statusCode == http.StatusOK {
			slog.Info(ms.Action.Name, "text", "Module is ready", "module", moduleName)
			return nil
		}

		slog.Warn(ms.Action.Name, "text", "Module is unready", "module", moduleName, "count", retryCount, "max", maxRetries)
		time.Sleep(waitDuration)
	}

	return errors.ModuleNotReady(moduleName)
}
