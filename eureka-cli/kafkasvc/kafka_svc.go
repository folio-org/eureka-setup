package kafkasvc

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/errors"
	"github.com/folio-org/eureka-cli/helpers"
)

// KafkaProcessor defines the interface for Kafka service operations
type KafkaProcessor interface {
	CheckReadiness() error
	GetConsumerGroupLag(tenant string, consumerGroup string, initialLag int) (lag int, err error)
}

// KafkaSvc provides functionality for Kafka operations including health checks and consumer lag monitoring
type KafkaSvc struct {
	Action *action.Action
}

// New creates a new KafkaSvc instance
func New(action *action.Action) *KafkaSvc {
	return &KafkaSvc{Action: action}
}

func (ks *KafkaSvc) CheckReadiness() error {
	kafkaCmd := fmt.Sprintf("timeout 10s kafka-broker-api-versions.sh --bootstrap-server %s", constant.KafkaTCP)
	stdout, stderr, err := helpers.ExecReturnOutput(exec.Command("docker", "exec", "-i", "kafka-tools", "bash", "-c", kafkaCmd))
	if err != nil || stderr.Len() > 0 {
		return errors.KafkaNotReady(err)
	}
	if stdout.Len() == 0 {
		return errors.KafkaBrokerAPIFailed()
	}
	slog.Info(ks.Action.Name, "text", "Kafka broker is ready and accessible")

	return nil
}

func (ks *KafkaSvc) GetConsumerGroupLag(tenant string, consumerGroup string, initialLag int) (lag int, err error) {
	kafkaCmd := fmt.Sprintf("timeout 30s kafka-consumer-groups.sh --bootstrap-server %s --describe --group %s | grep %s | awk '{print $6}'", constant.KafkaTCP, consumerGroup, tenant)
	stdout, stderr, err := helpers.ExecReturnOutput(exec.Command("docker", "exec", "-i", "kafka-tools", "bash", "-c", kafkaCmd))
	if err != nil {
		return initialLag, err
	}

	if stderr.Len() > 0 {
		stderrText := stderr.String()
		if strings.Contains(stderrText, constant.ErrNoActiveMembers) ||
			strings.Contains(stderrText, constant.ErrRebalancing) {
			time.Sleep(constant.AttachCapabilitySetsRebalanceWait)
			return initialLag, nil
		}
		if strings.Contains(stderrText, constant.ErrTimeoutException) {
			slog.Info(ks.Action.Name, "text", "Kafka timeout detected, waiting for Kafka to be ready")
			time.Sleep(constant.AttachCapabilitySetsTimeoutWait)
			return initialLag, nil
		}

		return 0, errors.ContainerCommandFailed(stderrText)
	}

	lag, err = strconv.Atoi(helpers.GetKafkaConsumerLagFromLogLine(stdout))
	if err != nil {
		slog.Error(ks.Action.Name, "error", err)
		return initialLag, nil
	}

	return lag, nil
}
