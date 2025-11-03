package modulesvc

import (
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/errors"
	"github.com/folio-org/eureka-cli/httpclient"
)

// ModuleReadinessChecker defines the interface for module readiness check operations
type ModuleReadinessChecker interface {
	CheckModuleReadiness(wg *sync.WaitGroup, errCh chan<- error, moduleName string, port int)
}

func (ms *ModuleSvc) CheckModuleReadiness(wg *sync.WaitGroup, errCh chan<- error, moduleName string, port int) {
	defer wg.Done()

	slog.Info(ms.Action.Name, "text", "Waiting module on port", "module", moduleName, "port", port)
	requestURL := ms.Action.GetRequestURL(strconv.Itoa(port), "/admin/health")
	for retryCount := range constant.ModuleReadinessMaxRetries {
		ready, _ := ms.checkContainerStatusCode(requestURL)
		if ready {
			slog.Info(ms.Action.Name, "text", "Module is ready", "module", moduleName)
			return
		}

		if retryCount == constant.ModuleReadinessMaxRetries {
			select {
			case errCh <- errors.ModuleNotReady(moduleName):
			default:
			}
			return
		}

		slog.Info(ms.Action.Name, "text", "Module is unready", "module", moduleName, "retryCount", retryCount, "maxRetries", constant.ModuleReadinessMaxRetries)
		time.Sleep(constant.ModuleReadinessWait)
	}
}

func (ms *ModuleSvc) checkContainerStatusCode(requestURL string) (bool, error) {
	httpResponse, err := ms.HTTPClient.GetReturnResponse(requestURL, map[string]string{})
	if err != nil {
		return false, err
	}
	if httpResponse == nil {
		return false, nil
	}
	defer httpclient.CloseResponse(httpResponse)

	return httpResponse.StatusCode == http.StatusOK, nil
}
