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

// ModuleReadinessChecker defines the interface for module readiness check operations
type ModuleReadinessChecker interface {
	CheckModuleReadiness(wg *sync.WaitGroup, errCh chan<- error, moduleName string, port int)
}

func (ms *ModuleSvc) CheckModuleReadiness(wg *sync.WaitGroup, errCh chan<- error, moduleName string, port int) {
	requestURL := ms.Action.GetRequestURL(strconv.Itoa(port), "/admin/health")
	ms.checkReadiness(wg, errCh, moduleName, requestURL)
}

func (ms *ModuleSvc) CheckModuleReadinessByURL(wg *sync.WaitGroup, errCh chan<- error, moduleName string, baseURL string) {
	requestURL := strings.TrimRight(baseURL, "/") + "/admin/health"
	ms.checkReadiness(wg, errCh, moduleName, requestURL)
}

func (ms *ModuleSvc) checkReadiness(wg *sync.WaitGroup, errCh chan<- error, moduleName string, requestURL string) {
	defer wg.Done()

	slog.Info(ms.Action.Name, "text", "Preparing module readiness check", "module", moduleName, "url", requestURL)
	maxRetries := helpers.DefaultInt(ms.ReadinessMaxRetries, constant.ModuleReadinessMaxRetries)
	waitDuration := helpers.DefaultDuration(ms.ReadinessWait, constant.ModuleReadinessWait)
	for retryCount := range maxRetries {
		statusCode, _ := ms.HTTPClient.Ping(requestURL)
		if statusCode == http.StatusOK {
			slog.Info(ms.Action.Name, "text", "Module is ready", "module", moduleName)
			return
		}

		slog.Warn(ms.Action.Name, "text", "Module is unready", "module", moduleName, "count", retryCount, "max", maxRetries)
		time.Sleep(waitDuration)
	}

	select {
	case errCh <- errors.ModuleNotReady(moduleName):
	default:
	}
}
