package testhelpers

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/j011195/eureka-setup/eureka-cli/gitrepository"
	"github.com/stretchr/testify/mock"
)

// MockGitClient is a mock implementation of gitclient.GitClientRunner
type MockGitClient struct {
	mock.Mock
}

func (m *MockGitClient) KongRepository() (*gitrepository.GitRepository, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gitrepository.GitRepository), args.Error(1)
}

func (m *MockGitClient) KeycloakRepository() (*gitrepository.GitRepository, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gitrepository.GitRepository), args.Error(1)
}

func (m *MockGitClient) PlatformCompleteRepository(branch plumbing.ReferenceName) (*gitrepository.GitRepository, error) {
	args := m.Called(branch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gitrepository.GitRepository), args.Error(1)
}

func (m *MockGitClient) Clone(repository *gitrepository.GitRepository) error {
	args := m.Called(repository)
	return args.Error(0)
}

func (m *MockGitClient) ResetHardPullFromOrigin(repository *gitrepository.GitRepository) error {
	args := m.Called(repository)
	return args.Error(0)
}
