package awssvc

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
)

type AWSProcessor interface {
	GetAuthorizationToken() (string, error)
	GetECRNamespace() string
	IsECRConfigured() bool
}

type AWSSvc struct {
	Action *action.Action
}

func New(action *action.Action) *AWSSvc {
	return &AWSSvc{
		Action: action,
	}
}

func (as *AWSSvc) GetAuthorizationToken() (string, error) {
	if !as.IsECRConfigured() {
		return "", nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	ecrClient := ecr.NewFromConfig(cfg)
	input := &ecr.GetAuthorizationTokenInput{}

	authToken, err := ecrClient.GetAuthorizationToken(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to get ECR authorization token: %w", err)
	}

	if len(authToken.AuthorizationData) == 0 {
		return "", errors.New("no authorization data returned from ECR")
	}

	authData := authToken.AuthorizationData[0]
	if authData.AuthorizationToken == nil {
		return "", errors.New("authorization token is nil")
	}

	if authData.ProxyEndpoint != nil {
		slog.Debug(as.Action.Name, "ecr_endpoint", *authData.ProxyEndpoint)
	}

	if authData.ExpiresAt != nil {
		slog.Debug(as.Action.Name, "token_expires_at", authData.ExpiresAt.String())
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(*authData.AuthorizationToken)
	if err != nil {
		return "", fmt.Errorf("failed to decode authorization token: %w", err)
	}

	authCreds := strings.SplitN(string(decodedBytes), ":", 2)
	if len(authCreds) != 2 {
		return "", errors.New("invalid authorization token format")
	}

	authConfig := map[string]string{
		"username": authCreds[0],
		"password": authCreds[1],
	}

	payload, err := json.Marshal(authConfig)
	if err != nil {
		return "", fmt.Errorf("failed to marshal auth config: %w", err)
	}

	encodedAuth := base64.StdEncoding.EncodeToString(payload)

	slog.Info(as.Action.Name, "message", "Successfully retrieved ECR authorization token")

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
