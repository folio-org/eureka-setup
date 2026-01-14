package helpers_test

import (
	"errors"
	"testing"

	"github.com/j011195/eureka-setup/eureka-cli/constant"
	apperrors "github.com/j011195/eureka-setup/eureka-cli/errors"
	"github.com/j011195/eureka-setup/eureka-cli/helpers"
	"github.com/stretchr/testify/assert"
)

func TestStripModuleVersion_WithDashAndSuffix(t *testing.T) {
	// Arrange - Removes everything after LAST dash
	moduleName := "mod-users-19.3.0"

	// Act
	result := helpers.StripModuleVersion(moduleName)

	// Assert
	assert.Equal(t, "mod-users", result) // Removes "-19.3.0"
}

func TestStripModuleVersion_WithMultipleDashes(t *testing.T) {
	// Arrange - Removes everything after LAST dash
	moduleName := "mod-inventory-storage-25.0.0"

	// Act
	result := helpers.StripModuleVersion(moduleName)

	// Assert
	assert.Equal(t, "mod-inventory-storage", result) // Removes "-25.0.0"
}

func TestStripModuleVersion_NoDash(t *testing.T) {
	// Arrange
	moduleName := "module"

	// Act
	result := helpers.StripModuleVersion(moduleName)

	// Assert
	assert.Equal(t, "module", result)
}

func TestStripModuleVersion_EmptyString(t *testing.T) {
	// Arrange
	moduleName := ""

	// Act
	result := helpers.StripModuleVersion(moduleName)

	// Assert
	assert.Equal(t, "", result)
}

func TestStripModuleVersion_OnlyDash(t *testing.T) {
	// Arrange
	moduleName := "-"

	// Act
	result := helpers.StripModuleVersion(moduleName)

	// Assert
	assert.Equal(t, "", result)
}

func TestFilterEmptyLines_EmptyString(t *testing.T) {
	// Arrange
	input := ""

	// Act
	result := helpers.FilterEmptyLines(input)

	// Assert
	assert.Equal(t, "", result)
}

func TestFilterEmptyLines_SingleLine(t *testing.T) {
	// Arrange
	input := "single line"

	// Act
	result := helpers.FilterEmptyLines(input)

	// Assert
	assert.Equal(t, "single line", result)
}

func TestFilterEmptyLines_MultipleLines(t *testing.T) {
	// Arrange
	input := "line1\nline2\nline3"

	// Act
	result := helpers.FilterEmptyLines(input)

	// Assert
	assert.Equal(t, "line1\nline2\nline3", result)
}

func TestFilterEmptyLines_WithEmptyLines(t *testing.T) {
	// Arrange
	input := "line1\n\nline2\n\n\nline3"

	// Act
	result := helpers.FilterEmptyLines(input)

	// Assert
	assert.Equal(t, "line1\nline2\nline3", result)
}

func TestFilterEmptyLines_WithWhitespaceLines(t *testing.T) {
	// Arrange
	input := "line1\n   \nline2\n\t\nline3"

	// Act
	result := helpers.FilterEmptyLines(input)

	// Assert
	assert.Equal(t, "line1\nline2\nline3", result)
}

func TestFilterEmptyLines_OnlyEmptyLines(t *testing.T) {
	// Arrange
	input := "\n\n\n"

	// Act
	result := helpers.FilterEmptyLines(input)

	// Assert
	assert.Equal(t, "", result)
}

func TestFilterEmptyLines_OnlyWhitespaceLines(t *testing.T) {
	// Arrange
	input := "   \n\t\t\n  \n"

	// Act
	result := helpers.FilterEmptyLines(input)

	// Assert
	assert.Equal(t, "", result)
}

func TestFilterEmptyLines_LeadingAndTrailingEmpty(t *testing.T) {
	// Arrange
	input := "\n\nline1\nline2\n\n"

	// Act
	result := helpers.FilterEmptyLines(input)

	// Assert
	assert.Equal(t, "line1\nline2", result)
}

func TestFilterEmptyLines_PreservesIndentation(t *testing.T) {
	// Arrange
	input := "  indented line\n\tnext line\nregular line"

	// Act
	result := helpers.FilterEmptyLines(input)

	// Assert
	assert.Equal(t, "  indented line\n\tnext line\nregular line", result)
}

func TestFilterEmptyLines_MixedContent(t *testing.T) {
	// Arrange
	input := `
out: Listen on ipv4:

Address         Port        Address         Port


--------------- ----------  --------------- ----------

192.168.1.100   8081        127.0.0.1       9000

`

	// Act
	result := helpers.FilterEmptyLines(input)

	// Assert
	expected := `out: Listen on ipv4:
Address         Port        Address         Port
--------------- ----------  --------------- ----------
192.168.1.100   8081        127.0.0.1       9000`
	assert.Equal(t, expected, result)
}

func TestFilterEmptyLines_WindowsLineEndings(t *testing.T) {
	// Arrange
	input := "line1\r\n\r\nline2\r\n"

	// Act
	result := helpers.FilterEmptyLines(input)

	// Assert - Note: strings.Split on "\n" will leave "\r" in the lines
	assert.Contains(t, result, "line1")
	assert.Contains(t, result, "line2")
}

// ==================== IsVersionGreater Tests ====================

func TestIsVersionGreater_ValidSemanticVersions(t *testing.T) {
	tests := []struct {
		name     string
		version1 string
		version2 string
		expected bool
	}{
		{
			name:     "TestIsVersionGreater_ValidSemanticVersions_MajorVersionDifference",
			version1: "2.0.0",
			version2: "1.0.0",
			expected: true,
		},
		{
			name:     "TestIsVersionGreater_ValidSemanticVersions_MinorVersionDifference",
			version1: "1.5.0",
			version2: "1.3.0",
			expected: true,
		},
		{
			name:     "TestIsVersionGreater_ValidSemanticVersions_PatchVersionDifference",
			version1: "1.0.5",
			version2: "1.0.3",
			expected: true,
		},
		{
			name:     "TestIsVersionGreater_ValidSemanticVersions_EqualVersions",
			version1: "1.0.0",
			version2: "1.0.0",
			expected: false,
		},
		{
			name:     "TestIsVersionGreater_ValidSemanticVersions_FirstVersionLower",
			version1: "1.0.0",
			version2: "2.0.0",
			expected: false,
		},
		{
			name:     "TestIsVersionGreater_ValidSemanticVersions_ComplexVersions",
			version1: "10.25.100",
			version2: "9.30.200",
			expected: true,
		},
		{
			name:     "TestIsVersionGreater_ValidSemanticVersions_LeadingZeros",
			version1: "1.02.3",
			version2: "1.02.2",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - test data provided in table

			// Act
			result := helpers.IsVersionGreater(tt.version1, tt.version2)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsVersionGreater_WithPreReleaseVersions(t *testing.T) {
	tests := []struct {
		name     string
		version1 string
		version2 string
		expected bool
	}{
		{
			name:     "TestIsVersionGreater_WithPreReleaseVersions_ReleaseVsPreRelease",
			version1: "1.0.0",
			version2: "1.0.0-alpha",
			expected: true,
		},
		{
			name:     "TestIsVersionGreater_WithPreReleaseVersions_AlphaVsBeta",
			version1: "1.0.0-beta",
			version2: "1.0.0-alpha",
			expected: true,
		},
		{
			name:     "TestIsVersionGreater_WithPreReleaseVersions_BetaVsRC",
			version1: "1.0.0-rc.1",
			version2: "1.0.0-beta.2",
			expected: true,
		},
		{
			name:     "TestIsVersionGreater_WithPreReleaseVersions_SamePreReleaseType",
			version1: "1.0.0-alpha.2",
			version2: "1.0.0-alpha.1",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - test data provided in table

			// Act
			result := helpers.IsVersionGreater(tt.version1, tt.version2)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsVersionGreater_WithBuildMetadata(t *testing.T) {
	tests := []struct {
		name     string
		version1 string
		version2 string
		expected bool
	}{
		{
			name:     "TestIsVersionGreater_WithBuildMetadata_DifferentBuilds",
			version1: "1.0.0+build.2",
			version2: "1.0.0+build.1",
			expected: false, // Build metadata should be ignored per semver spec
		},
		{
			name:     "TestIsVersionGreater_WithBuildMetadata_WithAndWithout",
			version1: "1.0.1",
			version2: "1.0.0+build.1",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - test data provided in table

			// Act
			result := helpers.IsVersionGreater(tt.version1, tt.version2)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsVersionGreater_InvalidSemanticVersions_FallbackToStringComparison(t *testing.T) {
	tests := []struct {
		name     string
		version1 string
		version2 string
		expected bool
	}{
		{
			name:     "TestIsVersionGreater_InvalidSemanticVersions_BothInvalid_Alphabetical",
			version1: "invalid-v2",
			version2: "invalid-v1",
			expected: true, // String comparison: "invalid-v2" > "invalid-v1"
		},
		{
			name:     "TestIsVersionGreater_InvalidSemanticVersions_FirstInvalid",
			version1: "not-a-version",
			version2: "1.0.0",
			expected: true, // String comparison: "not-a-version" > "1.0.0"
		},
		{
			name:     "TestIsVersionGreater_InvalidSemanticVersions_SecondInvalid",
			version1: "1.0.0",
			version2: "not-a-version",
			expected: false, // String comparison: "1.0.0" < "not-a-version"
		},
		{
			name:     "TestIsVersionGreater_InvalidSemanticVersions_EmptyStrings",
			version1: "",
			version2: "",
			expected: false,
		},
		{
			name:     "TestIsVersionGreater_InvalidSemanticVersions_FirstEmpty",
			version1: "",
			version2: "1.0.0",
			expected: false,
		},
		{
			name:     "TestIsVersionGreater_InvalidSemanticVersions_SecondEmpty",
			version1: "1.0.0",
			version2: "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - test data provided in table

			// Act
			result := helpers.IsVersionGreater(tt.version1, tt.version2)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsVersionGreater_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		version1 string
		version2 string
		expected bool
	}{
		{
			name:     "TestIsVersionGreater_EdgeCases_WithVPrefix",
			version1: "v2.0.0",
			version2: "v1.0.0",
			expected: true,
		},
		{
			name:     "TestIsVersionGreater_EdgeCases_MixedVPrefix",
			version1: "v2.0.0",
			version2: "1.0.0",
			expected: true,
		},
		{
			name:     "TestIsVersionGreater_EdgeCases_SnapshotVersions",
			version1: "1.0.0-SNAPSHOT",
			version2: "1.0.0",
			expected: false,
		},
		{
			name:     "TestIsVersionGreater_EdgeCases_VeryLongVersion",
			version1: "1.2.3.4.5.6.7.8.9.10",
			version2: "1.2.3.4.5.6.7.8.9.9",
			expected: false, // String comparison: "1.2.3.4.5.6.7.8.9.10" < "1.2.3.4.5.6.7.8.9.9" (lexicographically, "10" < "9")
		},
		{
			name:     "TestIsVersionGreater_EdgeCases_SingleDigit",
			version1: "2",
			version2: "1",
			expected: true, // String comparison: "2" > "1"
		},
		{
			name:     "TestIsVersionGreater_EdgeCases_TwoPartVersion",
			version1: "2.5",
			version2: "2.4",
			expected: true, // String comparison
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - test data provided in table

			// Act
			result := helpers.IsVersionGreater(tt.version1, tt.version2)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsVersionGreater_RealWorldModuleVersions(t *testing.T) {
	tests := []struct {
		name     string
		version1 string
		version2 string
		expected bool
	}{
		{
			name:     "TestIsVersionGreater_RealWorldModuleVersions_ModUsers",
			version1: "19.3.0",
			version2: "19.2.0",
			expected: true,
		},
		{
			name:     "TestIsVersionGreater_RealWorldModuleVersions_ModInventoryStorage",
			version1: "25.1.0",
			version2: "25.0.0",
			expected: true,
		},
		{
			name:     "TestIsVersionGreater_RealWorldModuleVersions_MajorUpgrade",
			version1: "20.0.0",
			version2: "19.99.99",
			expected: true,
		},
		{
			name:     "TestIsVersionGreater_RealWorldModuleVersions_PatchRelease",
			version1: "1.2.4",
			version2: "1.2.3",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - test data provided in table

			// Act
			result := helpers.IsVersionGreater(tt.version1, tt.version2)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ==================== IncrementSnapshotVersion Tests ====================

func TestIsSnapshot_ValidSnapshotVersion(t *testing.T) {
	// Arrange
	version := "13.1.0-SNAPSHOT.1093"

	// Act
	result := helpers.IsSnapshot(version)

	// Assert
	assert.True(t, result)
}

func TestIsSnapshot_SingleDigitBuildNumber(t *testing.T) {
	// Arrange
	version := "1.0.0-SNAPSHOT.5"

	// Act
	result := helpers.IsSnapshot(version)

	// Assert
	assert.True(t, result)
}

func TestIsSnapshot_ZeroBuildNumber(t *testing.T) {
	// Arrange
	version := "2.3.4-SNAPSHOT.0"

	// Act
	result := helpers.IsSnapshot(version)

	// Assert
	assert.True(t, result)
}

func TestIsSnapshot_EdgeOrdersVersion(t *testing.T) {
	// Arrange
	version := "3.3.0-SNAPSHOT.88"

	// Act
	result := helpers.IsSnapshot(version)

	// Assert
	assert.True(t, result)
}

func TestIsSnapshot_ModUsersVersion(t *testing.T) {
	// Arrange
	version := "19.3.0-SNAPSHOT.456"

	// Act
	result := helpers.IsSnapshot(version)

	// Assert
	assert.True(t, result)
}

func TestIsSnapshot_ReleaseVersion(t *testing.T) {
	// Arrange
	version := "1.0.0"

	// Act
	result := helpers.IsSnapshot(version)

	// Assert
	assert.False(t, result)
}

func TestIsSnapshot_SnapshotWithoutBuildNumber(t *testing.T) {
	// Arrange
	version := "1.0.0-SNAPSHOT"

	// Act
	result := helpers.IsSnapshot(version)

	// Assert
	assert.False(t, result)
}

func TestIsSnapshot_EmptyString(t *testing.T) {
	// Arrange
	version := ""

	// Act
	result := helpers.IsSnapshot(version)

	// Assert
	assert.False(t, result)
}

func TestIsSnapshot_PrereleaseVersion(t *testing.T) {
	// Arrange
	version := "1.0.0-alpha.1"

	// Act
	result := helpers.IsSnapshot(version)

	// Assert
	assert.False(t, result)
}

func TestIsSnapshot_BetaVersion(t *testing.T) {
	// Arrange
	version := "2.0.0-beta.3"

	// Act
	result := helpers.IsSnapshot(version)

	// Assert
	assert.False(t, result)
}

func TestIsSnapshot_RCVersion(t *testing.T) {
	// Arrange
	version := "3.0.0-rc.1"

	// Act
	result := helpers.IsSnapshot(version)

	// Assert
	assert.False(t, result)
}

func TestIsSnapshot_TableDrivenTests(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{
			name:     "TestIsSnapshot_TableDrivenTests_StandardSnapshot",
			version:  "13.1.0-SNAPSHOT.1093",
			expected: true,
		},
		{
			name:     "TestIsSnapshot_TableDrivenTests_EdgeOrders",
			version:  "3.3.0-SNAPSHOT.88",
			expected: true,
		},
		{
			name:     "TestIsSnapshot_TableDrivenTests_ModInventory",
			version:  "1.0.0-SNAPSHOT.123",
			expected: true,
		},
		{
			name:     "TestIsSnapshot_TableDrivenTests_ModUsers",
			version:  "19.3.0-SNAPSHOT.456",
			expected: true,
		},
		{
			name:     "TestIsSnapshot_TableDrivenTests_LargeBuildNumber",
			version:  "25.1.0-SNAPSHOT.9999",
			expected: true,
		},
		{
			name:     "TestIsSnapshot_TableDrivenTests_ReleaseVersion",
			version:  "1.0.0",
			expected: false,
		},
		{
			name:     "TestIsSnapshot_TableDrivenTests_SnapshotWithoutDot",
			version:  "1.0.0-SNAPSHOT",
			expected: false,
		},
		{
			name:     "TestIsSnapshot_TableDrivenTests_EmptyVersion",
			version:  "",
			expected: false,
		},
		{
			name:     "TestIsSnapshot_TableDrivenTests_AlphaPrerelease",
			version:  "1.0.0-alpha.1",
			expected: false,
		},
		{
			name:     "TestIsSnapshot_TableDrivenTests_BetaPrerelease",
			version:  "2.0.0-beta.3",
			expected: false,
		},
		{
			name:     "TestIsSnapshot_TableDrivenTests_RCPrerelease",
			version:  "3.0.0-rc.1",
			expected: false,
		},
		{
			name:     "TestIsSnapshot_TableDrivenTests_VersionWithBuildMetadata",
			version:  "1.0.0+build.123",
			expected: false,
		},
		{
			name:     "TestIsSnapshot_TableDrivenTests_ComplexVersion",
			version:  "10.25.100",
			expected: false,
		},
		{
			name:     "TestIsSnapshot_TableDrivenTests_VPrefixVersion",
			version:  "v1.0.0",
			expected: false,
		},
		{
			name:     "TestIsSnapshot_TableDrivenTests_InvalidVersion",
			version:  "invalid-version",
			expected: false,
		},
		{
			name:     "TestIsSnapshot_TableDrivenTests_SnapshotInMiddle",
			version:  "1.0-SNAPSHOT.123-alpha",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - test data provided in table

			// Act
			result := helpers.IsSnapshot(tt.version)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ==================== IsFolioNamespace Tests ====================

func TestIsFolioNamespace_SnapshotNamespace(t *testing.T) {
	// Arrange
	namespace := constant.SnapshotNamespace // "folioci"

	// Act
	result := helpers.IsFolioNamespace(namespace)

	// Assert
	assert.True(t, result)
}

func TestIsFolioNamespace_ReleaseNamespace(t *testing.T) {
	// Arrange
	namespace := constant.ReleaseNamespace // "folioorg"

	// Act
	result := helpers.IsFolioNamespace(namespace)

	// Assert
	assert.True(t, result)
}

func TestIsFolioNamespace_LocalNamespace(t *testing.T) {
	// Arrange
	namespace := constant.LocalNamespace // "foliolocal"

	// Act
	result := helpers.IsFolioNamespace(namespace)

	// Assert
	assert.False(t, result)
}

func TestIsFolioNamespace_CustomNamespace(t *testing.T) {
	// Arrange
	namespace := "docker.dev.folio.org"

	// Act
	result := helpers.IsFolioNamespace(namespace)

	// Assert
	assert.False(t, result)
}

func TestIsFolioNamespace_EmptyString(t *testing.T) {
	// Arrange
	namespace := ""

	// Act
	result := helpers.IsFolioNamespace(namespace)

	// Assert
	assert.False(t, result)
}

func TestIsFolioNamespace_DockerHubNamespace(t *testing.T) {
	// Arrange
	namespace := "mycompany"

	// Act
	result := helpers.IsFolioNamespace(namespace)

	// Assert
	assert.False(t, result)
}

func TestIsFolioNamespace_LocalhostRegistry(t *testing.T) {
	// Arrange
	namespace := "localhost:5000"

	// Act
	result := helpers.IsFolioNamespace(namespace)

	// Assert
	assert.False(t, result)
}

func TestIsFolioNamespace_GHCRNamespace(t *testing.T) {
	// Arrange
	namespace := "ghcr.io/folio-org"

	// Act
	result := helpers.IsFolioNamespace(namespace)

	// Assert
	assert.False(t, result)
}

func TestIsFolioNamespace_CaseVariation(t *testing.T) {
	// Arrange - Test that comparison is case-sensitive
	namespace := "FolioCI"

	// Act
	result := helpers.IsFolioNamespace(namespace)

	// Assert
	assert.False(t, result)
}

func TestIsFolioNamespace_TableDrivenTests(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		expected  bool
	}{
		{
			name:      "TestIsFolioNamespace_TableDrivenTests_SnapshotNamespace",
			namespace: "folioci",
			expected:  true,
		},
		{
			name:      "TestIsFolioNamespace_TableDrivenTests_ReleaseNamespace",
			namespace: "folioorg",
			expected:  true,
		},
		{
			name:      "TestIsFolioNamespace_TableDrivenTests_LocalNamespace",
			namespace: "foliolocal",
			expected:  false,
		},
		{
			name:      "TestIsFolioNamespace_TableDrivenTests_CustomNamespace",
			namespace: "docker.dev.folio.org",
			expected:  false,
		},
		{
			name:      "TestIsFolioNamespace_TableDrivenTests_EmptyString",
			namespace: "",
			expected:  false,
		},
		{
			name:      "TestIsFolioNamespace_TableDrivenTests_RandomNamespace",
			namespace: "mycompany",
			expected:  false,
		},
		{
			name:      "TestIsFolioNamespace_TableDrivenTests_LocalhostRegistry",
			namespace: "localhost:5000",
			expected:  false,
		},
		{
			name:      "TestIsFolioNamespace_TableDrivenTests_GHCR",
			namespace: "ghcr.io/folio-org",
			expected:  false,
		},
		{
			name:      "TestIsFolioNamespace_TableDrivenTests_CaseSensitive",
			namespace: "FolioCI",
			expected:  false,
		},
		{
			name:      "TestIsFolioNamespace_TableDrivenTests_WithSpaces",
			namespace: " folioci ",
			expected:  false,
		},
		{
			name:      "TestIsFolioNamespace_TableDrivenTests_PartialMatch",
			namespace: "folioci-test",
			expected:  false,
		},
		{
			name:      "TestIsFolioNamespace_TableDrivenTests_DockerIO",
			namespace: "docker.io/folioorg",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - test data provided in table

			// Act
			result := helpers.IsFolioNamespace(tt.namespace)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIncrementSnapshotVersion_ValidSnapshot(t *testing.T) {
	// Arrange
	version := "13.1.0-SNAPSHOT.1093"

	// Act
	result, err := helpers.IncrementSnapshotVersion(version)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "13.1.0-SNAPSHOT.1094", result)
}

func TestIncrementSnapshotVersion_SingleDigitBuildNumber(t *testing.T) {
	// Arrange
	version := "1.0.0-SNAPSHOT.5"

	// Act
	result, err := helpers.IncrementSnapshotVersion(version)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "1.0.0-SNAPSHOT.6", result)
}

func TestIncrementSnapshotVersion_ZeroBuildNumber(t *testing.T) {
	// Arrange
	version := "2.3.4-SNAPSHOT.0"

	// Act
	result, err := helpers.IncrementSnapshotVersion(version)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "2.3.4-SNAPSHOT.1", result)
}

func TestIncrementSnapshotVersion_LargeBuildNumber(t *testing.T) {
	// Arrange
	version := "19.3.0-SNAPSHOT.9999"

	// Act
	result, err := helpers.IncrementSnapshotVersion(version)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "19.3.0-SNAPSHOT.10000", result)
}

func TestIncrementSnapshotVersion_ComplexSemanticVersion(t *testing.T) {
	// Arrange
	version := "25.1.0-SNAPSHOT.456"

	// Act
	result, err := helpers.IncrementSnapshotVersion(version)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "25.1.0-SNAPSHOT.457", result)
}

func TestIncrementSnapshotVersion_EmptyString(t *testing.T) {
	// Arrange
	version := ""

	// Act
	result, err := helpers.IncrementSnapshotVersion(version)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "", result)
	assert.True(t, errors.Is(err, apperrors.ErrInvalidInput))
	assert.Contains(t, err.Error(), "version cannot be empty")
}

func TestIncrementSnapshotVersion_NotSnapshotVersion(t *testing.T) {
	// Arrange
	version := "1.0.0"

	// Act
	result, err := helpers.IncrementSnapshotVersion(version)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "", result)
	assert.True(t, errors.Is(err, apperrors.ErrInvalidInput))
	assert.Contains(t, err.Error(), "not a SNAPSHOT version with build number")
}

func TestIncrementSnapshotVersion_SnapshotWithoutBuildNumber(t *testing.T) {
	// Arrange
	version := "1.0.0-SNAPSHOT"

	// Act
	result, err := helpers.IncrementSnapshotVersion(version)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "", result)
	assert.True(t, errors.Is(err, apperrors.ErrInvalidInput))
	assert.Contains(t, err.Error(), "not a SNAPSHOT version with build number")
}

func TestIncrementSnapshotVersion_InvalidBuildNumber(t *testing.T) {
	// Arrange
	version := "1.0.0-SNAPSHOT.abc"

	// Act
	result, err := helpers.IncrementSnapshotVersion(version)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "", result)
	assert.True(t, errors.Is(err, apperrors.ErrInvalidInput))
	assert.Contains(t, err.Error(), "invalid build number")
}

func TestIncrementSnapshotVersion_NegativeBuildNumber(t *testing.T) {
	// Arrange - Negative numbers are technically valid integers, so they increment
	version := "1.0.0-SNAPSHOT.-5"

	// Act
	result, err := helpers.IncrementSnapshotVersion(version)

	// Assert - strconv.Atoi accepts negative numbers, so -5 becomes -4
	assert.NoError(t, err)
	assert.Equal(t, "1.0.0-SNAPSHOT.-4", result)
}

func TestIncrementSnapshotVersion_MultipleSnapshotDelimiters(t *testing.T) {
	// Arrange
	version := "1.0.0-SNAPSHOT.123-SNAPSHOT.456"

	// Act
	result, err := helpers.IncrementSnapshotVersion(version)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "", result)
	assert.True(t, errors.Is(err, apperrors.ErrInvalidInput))
	assert.Contains(t, err.Error(), "invalid SNAPSHOT version format")
}

func TestIncrementSnapshotVersion_BuildNumberWithLeadingZeros(t *testing.T) {
	// Arrange
	version := "1.0.0-SNAPSHOT.0099"

	// Act
	result, err := helpers.IncrementSnapshotVersion(version)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "1.0.0-SNAPSHOT.100", result)
}

func TestIncrementSnapshotVersion_TableDrivenTests(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		expected    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "TestIncrementSnapshotVersion_TableDrivenTests_StandardSnapshot",
			version:     "13.1.0-SNAPSHOT.1093",
			expected:    "13.1.0-SNAPSHOT.1094",
			expectError: false,
		},
		{
			name:        "TestIncrementSnapshotVersion_TableDrivenTests_EdgeOrders",
			version:     "3.3.0-SNAPSHOT.88",
			expected:    "3.3.0-SNAPSHOT.89",
			expectError: false,
		},
		{
			name:        "TestIncrementSnapshotVersion_TableDrivenTests_ModUsers",
			version:     "19.3.0-SNAPSHOT.456",
			expected:    "19.3.0-SNAPSHOT.457",
			expectError: false,
		},
		{
			name:        "TestIncrementSnapshotVersion_TableDrivenTests_ReleaseVersion",
			version:     "1.0.0",
			expected:    "",
			expectError: true,
			errorMsg:    "not a SNAPSHOT version",
		},
		{
			name:        "TestIncrementSnapshotVersion_TableDrivenTests_InvalidFormat",
			version:     "1.0.0-SNAPSHOT",
			expected:    "",
			expectError: true,
			errorMsg:    "not a SNAPSHOT version",
		},
		{
			name:        "TestIncrementSnapshotVersion_TableDrivenTests_EmptyVersion",
			version:     "",
			expected:    "",
			expectError: true,
			errorMsg:    "version cannot be empty",
		},
		{
			name:        "TestIncrementSnapshotVersion_TableDrivenTests_NonNumericBuild",
			version:     "1.0.0-SNAPSHOT.xyz",
			expected:    "",
			expectError: true,
			errorMsg:    "invalid build number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - test data provided in table

			// Act
			result, err := helpers.IncrementSnapshotVersion(tt.version)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expected, result)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
