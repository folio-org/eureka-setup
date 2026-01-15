package awssvc

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/errors"
)

// AWSProcessor defines the interface for AWS service operations
type AWSProcessor interface {
	GetAuthorizationToken() (string, error)
	GetECRNamespace() string
	IsECRConfigured() bool
}

// AWSSvc provides functionality for AWS operations including ECR
type AWSSvc struct {
	Action *action.Action
}

// New creates a new AWSSvc instance
func New(action *action.Action) *AWSSvc {
	return &AWSSvc{Action: action}
}

func (as *AWSSvc) GetAuthorizationToken() (string, error) {
	if !as.IsECRConfigured() {
		return "", nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), constant.ContextTimeoutAWSConfig)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", errors.AWSConfigLoadFailed(err)
	}

	ecrClient := ecr.NewFromConfig(cfg)
	input := &ecr.GetAuthorizationTokenInput{}

	authToken, err := ecrClient.GetAuthorizationToken(ctx, input)
	if err != nil {
		return "", errors.ECRAuthFailed(err)
	}
	if len(authToken.AuthorizationData) == 0 {
		return "", errors.ECRNoAuthData()
	}

	authData := authToken.AuthorizationData[0]
	if authData.AuthorizationToken == nil {
		return "", errors.ECRTokenNil()
	}
	if authData.ProxyEndpoint != nil {
		slog.Debug(as.Action.Name, "text", "Using proxy endpoint", "endpoint", *authData.ProxyEndpoint)
	}
	if authData.ExpiresAt != nil {
		slog.Debug(as.Action.Name, "text", "Using expires at", "expiresAt", authData.ExpiresAt.String())
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(*authData.AuthorizationToken)
	if err != nil {
		return "", errors.ECRTokenDecodeFailed(err)
	}

	authCreds := strings.SplitN(string(decodedBytes), ":", 2)
	if len(authCreds) != 2 {
		return "", errors.ECRTokenDecodeFailed(fmt.Errorf("invalid authorization token format"))
	}
	authConfig := map[string]string{
		"username": authCreds[0],
		"password": authCreds[1],
	}

	payload, err := json.Marshal(authConfig)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal auth config")
	}
	encodedAuth := base64.StdEncoding.EncodeToString(payload)
	slog.Info(as.Action.Name, "text", "Successfully retrieved ECR authorization token")

	return encodedAuth, nil
}

func (as *AWSSvc) GetECRNamespace() string {
	namespace := os.Getenv(constant.ECRRepositoryEnv)
	if namespace != "" {
		slog.Info(as.Action.Name, "text", "Using AWS ECR registry namespace", "namespace", namespace)
		return namespace
	}

	return ""
}

func (as *AWSSvc) IsECRConfigured() bool {
	return os.Getenv(constant.ECRRepositoryEnv) != ""
}
