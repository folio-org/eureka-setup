package managementsvc_test

import (
	"encoding/json"
	"strings"
	"testing"

	apperrors "github.com/folio-org/eureka-setup/eureka-cli/errors"
	"github.com/folio-org/eureka-setup/eureka-cli/internal/testhelpers"
	"github.com/folio-org/eureka-setup/eureka-cli/managementsvc"
	"github.com/folio-org/eureka-setup/eureka-cli/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func embeddedDescriptorExtract() *models.RegistryExtract {
	backendVersion := "1.11.0-SNAPSHOT.278"
	frontendVersion := "3.0.0"
	return &models.RegistryExtract{
		Modules: &models.ProxyModulesByRegistry{
			FolioModules: []*models.ProxyModule{
				{ID: "mod-ill-1.11.0-SNAPSHOT.278", Metadata: models.ProxyModuleMetadata{Name: "mod-ill", Version: &backendVersion, SidecarName: "mod-ill-sc"}},
				{ID: "k-int_ill-ui-3.0.0", Metadata: models.ProxyModuleMetadata{Name: "k-int_ill-ui", Version: &frontendVersion}},
			},
			EurekaModules: []*models.ProxyModule{},
			ModuleDescriptors: map[string]any{
				"mod-ill-1.11.0-SNAPSHOT.278": map[string]any{"id": "mod-ill-1.11.0-SNAPSHOT.278", "provides": []any{}},
				"k-int_ill-ui-3.0.0":          map[string]any{"id": "k-int_ill-ui-3.0.0"},
			},
		},
		BackendModules:    map[string]models.BackendModule{"mod-ill": {DeployModule: true, PrivatePort: 8080}},
		FrontendModules:   map[string]models.FrontendModule{"k-int_ill-ui": {DeployModule: true}},
		ModuleDescriptors: map[string]any{},
	}
}

// stubApplicationLifecycle answers the existence check with 404, accepts the registration and the discovery,
// and returns the decoded registration payload once CreateApplication has run
func stubApplicationLifecycle(t *testing.T, mockHTTP *testhelpers.MockHTTPClient) map[string]any {
	t.Helper()
	payload := map[string]any{}
	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(url string) bool { return strings.Contains(url, "/applications/") && !strings.Contains(url, "?") }),
		mock.Anything, mock.Anything).Once().Return(apperrors.ErrHTTP404NotFound)
	mockHTTP.On("PostReturnStruct",
		mock.MatchedBy(func(url string) bool { return strings.Contains(url, "/applications") }),
		mock.Anything, mock.Anything, mock.AnythingOfType("*models.ApplicationDescriptor")).
		Run(func(args mock.Arguments) {
			require.NoError(t, json.Unmarshal(args.Get(1).([]byte), &payload))
			args.Get(3).(*models.ApplicationDescriptor).ID = "app-ill-1.0.0"
		}).Return(nil)
	mockHTTP.On("PostReturnStruct",
		mock.MatchedBy(func(url string) bool { return strings.Contains(url, "/modules/discovery") }),
		mock.Anything, mock.Anything, mock.AnythingOfType("*models.ModuleDiscoveryResponse")).
		Run(func(args mock.Arguments) {
			args.Get(3).(*models.ModuleDiscoveryResponse).TotalRecords = 1
		}).Return(nil)

	return payload
}

func newDescriptorManagementSvc(fetchDescriptors bool) (*managementsvc.ManagementSvc, *testhelpers.MockHTTPClient) {
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakMasterAccessToken = "test-token"
	action.ConfigApplicationID = "app-ill-1.0.0"
	action.ConfigApplicationName = "app-ill"
	action.ConfigApplicationVersion = "1.0.0"
	action.ConfigRegistryURL = "https://folio-registry.k-int.com"
	action.ConfigApplicationFetchDescriptors = fetchDescriptors
	return managementsvc.New(action, mockHTTP, &MockTenantSvc{}), mockHTTP
}

// payloadEntries returns the objects under a list key of the registration payload
func payloadEntries(payload map[string]any, key string) []map[string]any {
	raw, _ := payload[key].([]any)
	entries := make([]map[string]any, 0, len(raw))
	for _, entry := range raw {
		entries = append(entries, entry.(map[string]any))
	}

	return entries
}

func assertEmbeddedDescriptorsInlined(t *testing.T, payload map[string]any) {
	t.Helper()
	modules, uiModules := payloadEntries(payload, "modules"), payloadEntries(payload, "uiModules")
	descriptors, uiDescriptors := payloadEntries(payload, "moduleDescriptors"), payloadEntries(payload, "uiModuleDescriptors")
	require.Len(t, modules, 1)
	require.Len(t, uiModules, 1)
	require.Len(t, descriptors, 1)
	require.Len(t, uiDescriptors, 1)
	assert.NotContains(t, modules[0], "url")
	assert.NotContains(t, uiModules[0], "url")
	assert.Equal(t, "mod-ill-1.11.0-SNAPSHOT.278", descriptors[0]["id"])
	assert.Equal(t, "k-int_ill-ui-3.0.0", uiDescriptors[0]["id"])
}

func TestCreateApplication_ApplicationDescriptor_TableDrivenTests(t *testing.T) {
	const descriptorURL = "https://registry.example.com/descriptors/k-int_ill-ui-3.0.0.json"
	tests := []struct {
		name             string
		fetchDescriptors bool
		arrange          func(t *testing.T, extract *models.RegistryExtract, mockHTTP *testhelpers.MockHTTPClient)
		assert           func(t *testing.T, payload map[string]any)
	}{
		{
			name:   "embedded descriptors are inlined without fetch-descriptors",
			assert: assertEmbeddedDescriptorsInlined,
		},
		{
			name:             "embedded descriptors skip the registry fetch with fetch-descriptors",
			fetchDescriptors: true,
			assert:           assertEmbeddedDescriptorsInlined,
		},
		{
			name: "module without embedded descriptor keeps the registry URL",
			arrange: func(t *testing.T, extract *models.RegistryExtract, mockHTTP *testhelpers.MockHTTPClient) {
				delete(extract.Modules.ModuleDescriptors, "k-int_ill-ui-3.0.0")
			},
			assert: func(t *testing.T, payload map[string]any) {
				uiModules := payloadEntries(payload, "uiModules")
				require.Len(t, uiModules, 1)
				assert.Equal(t, "https://folio-registry.k-int.com/_/proxy/modules/k-int_ill-ui-3.0.0", uiModules[0]["url"])
				assert.Len(t, payloadEntries(payload, "moduleDescriptors"), 1)
				assert.Empty(t, payloadEntries(payload, "uiModuleDescriptors"))
			},
		},
		{
			name: "descriptor URL wins over the registry URL",
			arrange: func(t *testing.T, extract *models.RegistryExtract, mockHTTP *testhelpers.MockHTTPClient) {
				delete(extract.Modules.ModuleDescriptors, "k-int_ill-ui-3.0.0")
				extract.Modules.ModuleDescriptorURLs = map[string]string{"k-int_ill-ui-3.0.0": descriptorURL}
			},
			assert: func(t *testing.T, payload map[string]any) {
				uiModules := payloadEntries(payload, "uiModules")
				require.Len(t, uiModules, 1)
				assert.Equal(t, descriptorURL, uiModules[0]["url"])
				assert.Empty(t, payloadEntries(payload, "uiModuleDescriptors"))
			},
		},
		{
			name:             "descriptor URL is used for the fetch",
			fetchDescriptors: true,
			arrange: func(t *testing.T, extract *models.RegistryExtract, mockHTTP *testhelpers.MockHTTPClient) {
				delete(extract.Modules.ModuleDescriptors, "k-int_ill-ui-3.0.0")
				extract.Modules.ModuleDescriptorURLs = map[string]string{"k-int_ill-ui-3.0.0": descriptorURL}
				mockHTTP.On("GetRetryReturnStruct", descriptorURL, mock.Anything, mock.Anything).Once().
					Run(func(args mock.Arguments) {
						*args.Get(2).(*any) = map[string]any{"id": "k-int_ill-ui-3.0.0", "fetched": true}
					}).Return(nil)
			},
			assert: func(t *testing.T, payload map[string]any) {
				uiModules, uiDescriptors := payloadEntries(payload, "uiModules"), payloadEntries(payload, "uiModuleDescriptors")
				require.Len(t, uiModules, 1)
				require.Len(t, uiDescriptors, 1)
				assert.NotContains(t, uiModules[0], "url")
				assert.Equal(t, true, uiDescriptors[0]["fetched"])
			},
		},
		{
			name: "pinned version ignores the descriptor URL and the embedded descriptor",
			arrange: func(t *testing.T, extract *models.RegistryExtract, mockHTTP *testhelpers.MockHTTPClient) {
				pinnedVersion := "3.1.0"
				extract.FrontendModules["k-int_ill-ui"] = models.FrontendModule{DeployModule: true, ModuleVersion: &pinnedVersion}
				extract.Modules.ModuleDescriptorURLs = map[string]string{"k-int_ill-ui-3.0.0": descriptorURL}
			},
			assert: func(t *testing.T, payload map[string]any) {
				uiModules := payloadEntries(payload, "uiModules")
				require.Len(t, uiModules, 1)
				assert.Equal(t, "k-int_ill-ui-3.1.0", uiModules[0]["id"])
				assert.Equal(t, "https://folio-registry.k-int.com/_/proxy/modules/k-int_ill-ui-3.1.0", uiModules[0]["url"])
				assert.Empty(t, payloadEntries(payload, "uiModuleDescriptors"))
			},
		},
		{
			name: "local-descriptor-path wins over the embedded descriptor",
			arrange: func(t *testing.T, extract *models.RegistryExtract, mockHTTP *testhelpers.MockHTTPClient) {
				descriptorPath := testhelpers.CreateTempJSONFile(t, map[string]any{"id": "mod-ill-1.11.0-SNAPSHOT.278", "local": true})
				extract.BackendModules["mod-ill"] = models.BackendModule{DeployModule: true, PrivatePort: 8080, LocalDescriptorPath: descriptorPath}
			},
			assert: func(t *testing.T, payload map[string]any) {
				modules, descriptors := payloadEntries(payload, "modules"), payloadEntries(payload, "moduleDescriptors")
				require.Len(t, modules, 1)
				require.Len(t, descriptors, 1)
				assert.NotContains(t, modules[0], "url")
				assert.Equal(t, true, descriptors[0]["local"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			svc, mockHTTP := newDescriptorManagementSvc(tt.fetchDescriptors)
			extract := embeddedDescriptorExtract()
			if tt.arrange != nil {
				tt.arrange(t, extract, mockHTTP)
			}
			payload := stubApplicationLifecycle(t, mockHTTP)

			// Act
			err := svc.CreateApplication(extract)

			// Assert
			assert.NoError(t, err)
			mockHTTP.AssertExpectations(t)
			mockHTTP.AssertNotCalled(t, "GetRetryReturnStruct",
				mock.MatchedBy(func(url string) bool { return strings.Contains(url, "/_/proxy/modules/") }),
				mock.Anything, mock.Anything)
			tt.assert(t, payload)
		})
	}
}
