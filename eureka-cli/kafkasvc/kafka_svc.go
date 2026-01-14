package kafkasvc

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/j011195/eureka-setup/eureka-cli/constant"
	"github.com/j011195/eureka-setup/eureka-cli/errors"
	"github.com/j011195/eureka-setup/eureka-cli/execsvc"
	"github.com/j011195/eureka-setup/eureka-cli/helpers"
)

// KafkaProcessor defines the interface for Kafka service operations
type KafkaProcessor interface {
	CheckBrokerReadiness() error
	PollConsumerGroup(tenantName string) error
}

// KafkaSvc provides functionality for Kafka operations including health checks and consumer lag monitoring
type KafkaSvc struct {
	Action           *action.Action
	ExecSvc          execsvc.CommandRunner
	RebalanceRetries int
	PollMaxRetries   int
	RebalanceWait    time.Duration
	PollWait         time.Duration
	TimeoutWait      time.Duration
}

// New creates a new KafkaSvc instance
func New(action *action.Action, execSvc execsvc.CommandRunner) *KafkaSvc {
	return &KafkaSvc{
		Action:  action,
		ExecSvc: execSvc,
	}
}

func (ks *KafkaSvc) CheckBrokerReadiness() error {
	kafkaCmd := fmt.Sprintf("timeout 30s kafka-broker-api-versions.sh --bootstrap-server %s", constant.KafkaTCP)
	stdout, stderr, err := ks.ExecSvc.ExecReturnOutput(exec.Command("docker", "exec", "-i", "kafka-tools", "bash", "-c", kafkaCmd))
	if err != nil || stderr.Len() > 0 {
		return errors.KafkaNotReady(err)
	}
	if stdout.Len() == 0 {
		return errors.KafkaBrokerAPIFailed()
	}
	slog.Info(ks.Action.Name, "text", "Broker is ready and accessible")

	return nil
}

func (ks *KafkaSvc) PollConsumerGroup(tenantName string) error {
	slog.Info(ks.Action.Name, "text", "Preparing broker readiness check")
	if err := ks.CheckBrokerReadiness(); err != nil {
		slog.Warn(ks.Action.Name, "text", "Broker is not fully ready", "error", err)
	}

	consumerGroup := fmt.Sprintf("%s-%s", ks.Action.ConfigEnvFolio, constant.ConsumerGroupSuffix)
	slog.Info(ks.Action.Name, "text", "Polling consumer group", "consumerGroup", consumerGroup, "tenant", tenantName)

	var lag int
	rebalanceRetryCount := 0
	rebalanceMaxRetries := helpers.DefaultInt(ks.RebalanceRetries, constant.ConsumerGroupRebalanceRetries)
	pollMaxRetries := helpers.DefaultInt(ks.PollMaxRetries, constant.ConsumerGroupPollMaxRetries)
	rebalanceWait := helpers.DefaultDuration(ks.RebalanceWait, constant.AttachCapabilitySetsRebalanceWait)
	pollWait := helpers.DefaultDuration(ks.PollWait, constant.AttachCapabilitySetsPollWait)
	for pollRetryCount := range pollMaxRetries {
		lag, err := ks.getConsumerGroupLag(tenantName, consumerGroup, lag)
		if err != nil {
			rebalanceRetryCount++
			if rebalanceRetryCount >= rebalanceMaxRetries {
				return errors.ConsumerGroupRebalanceTimeout(consumerGroup, err)
			}

			slog.Warn(ks.Action.Name, "text", "Waiting for consumer group to rebalance", "count", rebalanceRetryCount, "max", rebalanceMaxRetries)
			time.Sleep(rebalanceWait)
			continue
		}

		rebalanceRetryCount = 0
		if lag == 0 {
			slog.Info(ks.Action.Name, "text", "Consumer group has no new message to process", "consumerGroup", consumerGroup)
			return nil
		}

		slog.Warn(ks.Action.Name, "text", "Waiting for consumer group", "consumerGroup", consumerGroup, "lag", lag, "count", pollRetryCount, "max", pollMaxRetries)
		time.Sleep(pollWait)
	}

	return errors.ConsumerGroupPollTimeout(consumerGroup, pollMaxRetries)
}

func (ks *KafkaSvc) getConsumerGroupLag(tenant string, consumerGroup string, initialLag int) (lag int, err error) {
	rebalanceWait := helpers.DefaultDuration(ks.RebalanceWait, constant.AttachCapabilitySetsRebalanceWait)
	timeoutWait := helpers.DefaultDuration(ks.TimeoutWait, constant.AttachCapabilitySetsTimeoutWait)

	kafkaCmd := fmt.Sprintf("timeout 30s kafka-consumer-groups.sh --bootstrap-server %s --describe --group %s | grep %s | awk '{print $6}'", constant.KafkaTCP, consumerGroup, tenant)
	stdout, stderr, err := ks.ExecSvc.ExecReturnOutput(exec.Command("docker", "exec", "-i", "kafka-tools", "bash", "-c", kafkaCmd))
	if err != nil {
		return initialLag, err
	}

	if stderr.Len() > 0 {
		stderrText := stderr.String()
		if strings.Contains(stderrText, constant.ErrNoActiveMembers) ||
			strings.Contains(stderrText, constant.ErrRebalancing) {
			time.Sleep(rebalanceWait)
			return initialLag, nil
		}
		if strings.Contains(stderrText, constant.ErrTimeoutException) {
			time.Sleep(timeoutWait)
			return initialLag, nil
		}

		return initialLag, errors.ContainerCommandFailed(stderrText)
	}

	lag, err = strconv.Atoi(helpers.GetKafkaConsumerLagFromLogLine(stdout))
	if err != nil {
		slog.Error(ks.Action.Name, "error", err)
		return initialLag, nil
	}

	return lag, nil
}
