package constant

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ==================== GetContainerTypes Tests ====================

func TestGetContainerTypes(t *testing.T) {
	// Act
	types := GetContainerTypes()

	// Assert
	assert.Len(t, types, 3)
	assert.Contains(t, types, Module)
	assert.Contains(t, types, Sidecar)
	assert.Contains(t, types, Management)
	assert.Equal(t, []string{Module, Sidecar, Management}, types)
}

// ==================== GetTenantTypes Tests ====================

func TestGetTenantTypes(t *testing.T) {
	// Act
	types := GetTenantTypes()

	// Assert
	assert.Len(t, types, 2)
	assert.ElementsMatch(t, []TenantType{Central, Member}, types)
}

func TestGetTenantTypes_DoesNotIncludeAll(t *testing.T) {
	// Act
	types := GetTenantTypes()

	// Assert
	assert.NotContains(t, types, All)
}

func TestGetTenantTypes_DoesNotIncludeDefault(t *testing.T) {
	// Act
	types := GetTenantTypes()

	// Assert
	assert.NotContains(t, types, Default)
}

// ==================== GetTokenTypes Tests ====================

func TestGetTokenTypes(t *testing.T) {
	// Act
	types := GetTokenTypes()

	// Assert
	assert.Len(t, types, 3)
	assert.Contains(t, types, DefaultToken)
	assert.Contains(t, types, MasterCustomToken)
	assert.Contains(t, types, MasterAdminCLIToken)
	assert.Equal(t, []string{DefaultToken, MasterCustomToken, MasterAdminCLIToken}, types)
}

// ==================== GetNamespaces Tests ====================

func TestGetNamespaces(t *testing.T) {
	// Act
	namespaces := GetNamespaces()

	// Assert
	assert.Len(t, namespaces, 3)
	assert.Contains(t, namespaces, SnapshotNamespace)
	assert.Contains(t, namespaces, ReleaseNamespace)
	assert.Contains(t, namespaces, LocalNamespace)
	assert.Equal(t, []string{SnapshotNamespace, ReleaseNamespace, LocalNamespace}, namespaces)
}

func TestGetNamespaces_OrderPreserved(t *testing.T) {
	// Act
	namespaces := GetNamespaces()

	// Assert - Verify order is preserved
	assert.Equal(t, SnapshotNamespace, namespaces[0])
	assert.Equal(t, ReleaseNamespace, namespaces[1])
	assert.Equal(t, LocalNamespace, namespaces[2])
}

func TestGetNamespaces_ExpectedValues(t *testing.T) {
	// Act
	namespaces := GetNamespaces()

	// Assert - Verify actual string values
	assert.Equal(t, "folioci", namespaces[0])
	assert.Equal(t, "folioorg", namespaces[1])
	assert.Equal(t, "foliolocal", namespaces[2])
}

// ==================== GetInitialRequiredContainers Tests ====================

func TestGetInitialRequiredContainers(t *testing.T) {
	// Act
	containers := GetInitialRequiredContainers()

	// Assert
	assert.Len(t, containers, 7)
	assert.Contains(t, containers, PostgreSQLContainer)
	assert.Contains(t, containers, KafkaContainer)
	assert.Contains(t, containers, KafkaToolsContainer)
	assert.Contains(t, containers, VaultContainer)
	assert.Contains(t, containers, KeycloakProxyContainer)
	assert.Contains(t, containers, KeycloakContainer)
	assert.Contains(t, containers, KongContainer)
}

func TestGetInitialRequiredContainers_OrderPreserved(t *testing.T) {
	// Act
	containers := GetInitialRequiredContainers()

	// Assert - Verify order is preserved
	assert.Equal(t, PostgreSQLContainer, containers[0])
	assert.Equal(t, KafkaContainer, containers[1])
	assert.Equal(t, KafkaToolsContainer, containers[2])
	assert.Equal(t, VaultContainer, containers[3])
	assert.Equal(t, KeycloakProxyContainer, containers[4])
	assert.Equal(t, KeycloakContainer, containers[5])
	assert.Equal(t, KongContainer, containers[6])
}

// ==================== GetProfiles Tests ====================

func TestGetProfiles(t *testing.T) {
	// Act
	profiles := GetProfiles()

	// Assert
	assert.Len(t, profiles, 9)
	assert.Contains(t, profiles, CombinedProfile)
	assert.Contains(t, profiles, CombinedNativeProfile)
	assert.Contains(t, profiles, ExportProfile)
	assert.Contains(t, profiles, SearchProfile)
	assert.Contains(t, profiles, EdgeProfile)
	assert.Contains(t, profiles, ECSProfile)
	assert.Contains(t, profiles, ECSSingleProfile)
	assert.Contains(t, profiles, ECSMigrationProfile)
	assert.Contains(t, profiles, ImportProfile)
}

func TestGetProfiles_OrderPreserved(t *testing.T) {
	// Act
	profiles := GetProfiles()

	// Assert
	assert.Equal(t, CombinedProfile, profiles[0])
	assert.Equal(t, CombinedNativeProfile, profiles[1])
	assert.Equal(t, ExportProfile, profiles[2])
	assert.Equal(t, SearchProfile, profiles[3])
	assert.Equal(t, EdgeProfile, profiles[4])
	assert.Equal(t, ECSProfile, profiles[5])
	assert.Equal(t, ECSSingleProfile, profiles[6])
	assert.Equal(t, ECSMigrationProfile, profiles[7])
	assert.Equal(t, ImportProfile, profiles[8])
}

// ==================== GetDefaultProfile Tests ====================

func TestGetDefaultProfile(t *testing.T) {
	// Act
	profile := GetDefaultProfile()

	// Assert
	assert.Equal(t, CombinedProfile, profile)
	assert.Equal(t, "combined", profile)
}

func TestGetDefaultProfile_IncludedInProfiles(t *testing.T) {
	// Arrange
	profiles := GetProfiles()

	// Act
	defaultProfile := GetDefaultProfile()

	// Assert
	assert.Contains(t, profiles, defaultProfile)
}

// ==================== Constant Values Tests ====================

func TestTenantTypeConstants(t *testing.T) {
	// Assert - Testing constant values
	assert.Equal(t, "", string(All))
	assert.Equal(t, "default", string(Default))
	assert.Equal(t, "central", string(Central))
	assert.Equal(t, "member", string(Member))
}

func TestKeycloakGrantTypeConstants(t *testing.T) {
	// Assert - Testing constant values
	assert.Equal(t, "client_credentials", string(ClientCredentials))
	assert.Equal(t, "password", string(Password))
}

func TestTokenTypeConstants(t *testing.T) {
	// Assert
	assert.Equal(t, "tenant", DefaultToken)
	assert.Equal(t, "master-custom", MasterCustomToken)
	assert.Equal(t, "master-admin-cli", MasterAdminCLIToken)
}

func TestContainerTypeConstants(t *testing.T) {
	// Assert
	assert.Equal(t, "management", Management)
	assert.Equal(t, "module", Module)
	assert.Equal(t, "sidecar", Sidecar)
}

func TestProfileConstants(t *testing.T) {
	// Assert
	assert.Equal(t, "combined", CombinedProfile)
	assert.Equal(t, "combined-native", CombinedNativeProfile)
	assert.Equal(t, "export", ExportProfile)
	assert.Equal(t, "search", SearchProfile)
	assert.Equal(t, "edge", EdgeProfile)
	assert.Equal(t, "ecs", ECSProfile)
	assert.Equal(t, "ecs-single", ECSSingleProfile)
	assert.Equal(t, "import", ImportProfile)
}
