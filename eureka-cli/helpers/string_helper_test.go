package helpers_test

import (
	"testing"

	"github.com/folio-org/eureka-cli/helpers"
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
