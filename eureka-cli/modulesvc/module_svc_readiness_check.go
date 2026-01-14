package modulesvc

import (
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/j011195/eureka-setup/eureka-cli/constant"
	"github.com/j011195/eureka-setup/eureka-cli/errors"
	"github.com/j011195/eureka-setup/eureka-cli/helpers"
)

// ModuleReadinessChecker defines the interface for module readiness check operations
type ModuleReadinessChecker interface {
	CheckModuleReadiness(wg *sync.WaitGroup, errCh chan<- error, moduleName string, port int)
}

func (ms *ModuleSvc) CheckModuleReadiness(wg *sync.WaitGroup, errCh chan<- error, moduleName string, port int) {
	defer wg.Done()

	slog.Info(ms.Action.Name, "text", "Preparing module readiness check", "module", moduleName, "port", port)
	maxRetries := helpers.DefaultInt(ms.ReadinessMaxRetries, constant.ModuleReadinessMaxRetries)
	waitDuration := helpers.DefaultDuration(ms.ReadinessWait, constant.ModuleReadinessWait)
	requestURL := ms.Action.GetRequestURL(strconv.Itoa(port), "/admin/health")
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
