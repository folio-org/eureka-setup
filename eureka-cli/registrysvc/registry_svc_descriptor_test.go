package registrysvc_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	apperrors "github.com/folio-org/eureka-setup/eureka-cli/errors"
	"github.com/folio-org/eureka-setup/eureka-cli/internal/testhelpers"
	"github.com/folio-org/eureka-setup/eureka-cli/models"
	"github.com/folio-org/eureka-setup/eureka-cli/registrysvc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func sampleApplicationDescriptor() map[string]any {
	return map[string]any{
		"id":      "app-ill-1.0.0",
		"name":    "app-ill",
		"version": "1.0.0",
		"dependencies": []any{
			map[string]any{"name": "app-platform-minimal", "version": "^2.0.0", "optional": false},
		},
		"modules": []any{
			map[string]any{"id": "mod-ill-1.11.0-SNAPSHOT.278+1787670557413-d0e2d3d", "name": "mod-ill", "version": "1.11.0-SNAPSHOT.278+1787670557413-d0e2d3d"},
			map[string]any{"name": "mod-ill-keycloak", "version": "2.0.0", "url": "https://folio-registry.k-int.com/_/proxy/modules/mod-ill-keycloak-2.0.0"},
		},
		"uiModules": []any{
			map[string]any{"id": "k-int_ill-ui-3.0.0", "name": "k-int_ill-ui", "version": "3.0.0"},
		},
		"moduleDescriptors": []any{
			map[string]any{"id": "mod-ill-1.11.0-SNAPSHOT.278+1787670557413-d0e2d3d", "provides": []any{}},
		},
		"uiModuleDescriptors": []any{
			map[string]any{"id": "k-int_ill-ui-3.0.0"},
			map[string]any{"name": "descriptor-without-id"},
		},
	}
}

func newDescriptorSvc(t *testing.T, source string) (*registrysvc.RegistrySvc, *testhelpers.MockHTTPClient) {
	t.Helper()
	testhelpers.SetTempConfigDir(t)
	mockHTTP := &testhelpers.MockHTTPClient{}
	act := testhelpers.NewMockAction()
	act.ConfigFarURL = "http://far.example.com"
	act.ConfigApplicationName = "app-ill"
	act.ConfigApplicationDescriptor = source
	return registrysvc.New(act, mockHTTP, &MockAWSSvc{}), mockHTTP
}

func TestGetModules_ApplicationDescriptorSources_TableDrivenTests(t *testing.T) {
	const url = "https://registry.example.com/app-ill-1.0.0.json"
	tests := []struct {
		name         string
		source       func(t *testing.T, mockHTTP *testhelpers.MockHTTPClient) string
		skipRegistry bool
		forceRefresh bool
		viaHTTP      bool
	}{
		{
			name: "file path",
			source: func(t *testing.T, mockHTTP *testhelpers.MockHTTPClient) string {
				return testhelpers.CreateTempJSONFile(t, sampleApplicationDescriptor())
			},
			forceRefresh: true,
		},
		{
			name: "home-relative file path",
			source: func(t *testing.T, mockHTTP *testhelpers.MockHTTPClient) string {
				userHome, err := os.UserHomeDir()
				require.NoError(t, err)
				testhelpers.CreateJSONFileInDir(t, userHome, "app-ill-1.0.0.json", sampleApplicationDescriptor())
				return "~/app-ill-1.0.0.json"
			},
			forceRefresh: true,
		},
		{
			name: "URL",
			source: func(t *testing.T, mockHTTP *testhelpers.MockHTTPClient) string {
				data, err := json.Marshal(sampleApplicationDescriptor())
				require.NoError(t, err)
				mockHTTP.On("GetRetryReturnStruct", url, mock.Anything, mock.AnythingOfType("*models.ApplicationDescriptorSource")).
					Run(func(args mock.Arguments) {
						require.NoError(t, json.Unmarshal(data, args.Get(2).(*models.ApplicationDescriptorSource)))
					}).Return(nil)
				return url
			},
			viaHTTP: true,
		},
		{
			name: "file path with the skip-registry flag",
			source: func(t *testing.T, mockHTTP *testhelpers.MockHTTPClient) string {
				return testhelpers.CreateTempJSONFile(t, sampleApplicationDescriptor())
			},
			skipRegistry: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			svc, mockHTTP := newDescriptorSvc(t, "")
			svc.Action.ConfigApplicationDescriptor = tt.source(t, mockHTTP)
			svc.Action.Param.SkipRegistry = tt.skipRegistry

			// Act
			result, err := svc.GetModules(true, tt.forceRefresh)
			require.NoError(t, err)
			svc.ResolveModuleMetadata(result)

			// Assert
			require.Len(t, result.FolioModules, 2)
			buildMetadataModule := result.FolioModules[0]
			assert.Equal(t, "mod-ill-1.11.0-SNAPSHOT.278+1787670557413-d0e2d3d", buildMetadataModule.ID)
			assert.Equal(t, "enable", buildMetadataModule.Action)
			assert.Equal(t, "mod-ill", buildMetadataModule.Metadata.Name, "the ID pattern does not parse build metadata, the descriptor's name must be kept")
			require.NotNil(t, buildMetadataModule.Metadata.Version)
			assert.Equal(t, "1.11.0-SNAPSHOT.278+1787670557413-d0e2d3d", *buildMetadataModule.Metadata.Version)
			assert.Equal(t, "mod-ill-sc", buildMetadataModule.Metadata.SidecarName)
			assert.Equal(t, "k-int_ill-ui-3.0.0", result.FolioModules[1].ID)
			require.Len(t, result.EurekaModules, 1)
			synthesized := result.EurekaModules[0]
			assert.Equal(t, "mod-ill-keycloak-2.0.0", synthesized.ID, "ID is synthesized from name and version when absent")
			assert.Equal(t, "mod-ill-keycloak", synthesized.Metadata.Name)
			require.NotNil(t, synthesized.Metadata.Version)
			assert.Equal(t, "2.0.0", *synthesized.Metadata.Version)
			assert.Len(t, result.ModuleDescriptors, 2)
			assert.Contains(t, result.ModuleDescriptors, "mod-ill-1.11.0-SNAPSHOT.278+1787670557413-d0e2d3d")
			assert.Contains(t, result.ModuleDescriptors, "k-int_ill-ui-3.0.0")
			assert.Equal(t, map[string]string{"mod-ill-keycloak-2.0.0": "https://folio-registry.k-int.com/_/proxy/modules/mod-ill-keycloak-2.0.0"}, result.ModuleDescriptorURLs, "only modules with a url in the descriptor are keyed")
			assert.NoFileExists(t, modulesFilePath(t), "descriptor-backed inventory must not touch the modules cache")
			mockHTTP.AssertExpectations(t)
			if !tt.viaHTTP {
				mockHTTP.AssertNotCalled(t, "GetRetryReturnStruct", mock.Anything, mock.Anything, mock.Anything)
			}
		})
	}
}

func TestGetModules_ApplicationDescriptorWithLspComponents(t *testing.T) {
	// Arrange
	path := testhelpers.CreateTempJSONFile(t, sampleApplicationDescriptor())
	svc, mockHTTP := newDescriptorSvc(t, path)
	svc.Action.ConfigLspURL = "http://lsp.example.com/descriptor.json"
	descriptor := buildLSPResponse(
		[]models.PlatformApplication{{Name: "app-platform-minimal", Version: "2.0.0"}},
		nil, nil,
		[]models.PlatformApplication{{Name: "folio-module-sidecar", Version: "4.1.0-SNAPSHOT.2329"}, {Name: "mgr-applications", Version: "3.0.0"}},
	)
	stubLSP(mockHTTP, svc.Action.ConfigLspURL, descriptor)

	// Act
	result, err := svc.GetModules(false, true)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result.FolioModules, 2, "platform applications are not fetched from FAR")
	require.Len(t, result.EurekaModules, 3)
	assert.Equal(t, "folio-module-sidecar-4.1.0-SNAPSHOT.2329", result.EurekaModules[1].ID)
	assert.Equal(t, "mgr-applications-3.0.0", result.EurekaModules[2].ID)
	assert.NoFileExists(t, modulesFilePath(t))
	mockHTTP.AssertExpectations(t)
}

func TestGetModules_ApplicationDescriptorLspFetchError(t *testing.T) {
	// Arrange
	path := testhelpers.CreateTempJSONFile(t, sampleApplicationDescriptor())
	svc, mockHTTP := newDescriptorSvc(t, path)
	svc.Action.ConfigLspURL = "http://lsp.example.com/descriptor.json"
	mockHTTP.On("GetRetryReturnStruct", svc.Action.ConfigLspURL, mock.Anything, mock.Anything).Return(errors.New("connection refused"))

	// Act
	result, err := svc.GetModules(false, true)

	// Assert
	assert.Nil(t, result)
	assert.ErrorContains(t, err, "connection refused")
}

func TestGetModules_ApplicationDescriptorErrors_TableDrivenTests(t *testing.T) {
	const source = "https://registry.example.com/app-ill-1.0.0.json"
	tests := []struct {
		name         string
		source       func(t *testing.T, mockHTTP *testhelpers.MockHTTPClient) string
		wantContains []string
	}{
		{
			name: "URL fetch error",
			source: func(t *testing.T, mockHTTP *testhelpers.MockHTTPClient) string {
				mockHTTP.On("GetRetryReturnStruct", source, mock.Anything, mock.Anything).Return(errors.New("connection refused"))
				return source
			},
			wantContains: []string{source, "connection refused"},
		},
		{
			name: "file missing",
			source: func(t *testing.T, mockHTTP *testhelpers.MockHTTPClient) string {
				return filepath.Join(t.TempDir(), "missing.json")
			},
			wantContains: []string{"missing.json"},
		},
		{
			name: "no modules",
			source: func(t *testing.T, mockHTTP *testhelpers.MockHTTPClient) string {
				return testhelpers.CreateTempJSONFile(t, map[string]any{"id": "app-empty-1.0.0", "name": "app-empty", "version": "1.0.0"})
			},
			wantContains: []string{"no modules"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			svc, mockHTTP := newDescriptorSvc(t, "")
			svc.Action.ConfigApplicationDescriptor = tt.source(t, mockHTTP)

			// Act
			result, err := svc.GetModules(false, false)

			// Assert
			assert.Nil(t, result)
			assert.ErrorIs(t, err, apperrors.ErrInvalidInput)
			for _, want := range tt.wantContains {
				assert.ErrorContains(t, err, want)
			}
		})
	}
}
