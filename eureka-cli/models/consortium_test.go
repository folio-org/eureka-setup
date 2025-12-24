package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ==================== SortedConsortiumTenant.String() Tests ====================

func TestSortedConsortiumTenant_String_Central(t *testing.T) {
	// Arrange
	tenant := SortedConsortiumTenant{
		Consortium: "consortium-1",
		Name:       "diku",
		IsCentral:  1,
	}

	// Act
	result := tenant.String()

	// Assert
	assert.Equal(t, "diku (central)", result)
}

func TestSortedConsortiumTenant_String_Member(t *testing.T) {
	// Arrange
	tenant := SortedConsortiumTenant{
		Consortium: "consortium-1",
		Name:       "college",
		IsCentral:  0,
	}

	// Act
	result := tenant.String()

	// Assert
	assert.Equal(t, "college", result)
}

// ==================== SortedConsortiumTenants.String() Tests ====================

func TestSortedConsortiumTenants_String_MultipleTenants(t *testing.T) {
	// Arrange
	tenants := SortedConsortiumTenants{
		{Consortium: "c1", Name: "diku", IsCentral: 1},
		{Consortium: "c1", Name: "college", IsCentral: 0},
		{Consortium: "c1", Name: "university", IsCentral: 0},
	}

	// Act
	result := tenants.String()

	// Assert
	assert.Equal(t, "diku, college, university", result)
}

func TestSortedConsortiumTenants_String_SingleTenant(t *testing.T) {
	// Arrange
	tenants := SortedConsortiumTenants{
		{Consortium: "c1", Name: "diku", IsCentral: 1},
	}

	// Act
	result := tenants.String()

	// Assert
	assert.Equal(t, "diku", result)
}

func TestSortedConsortiumTenants_String_Empty(t *testing.T) {
	// Arrange
	tenants := SortedConsortiumTenants{}

	// Act
	result := tenants.String()

	// Assert
	assert.Equal(t, "", result)
}

func TestSortedConsortiumTenants_String_OnlyMembers(t *testing.T) {
	// Arrange
	tenants := SortedConsortiumTenants{
		{Consortium: "c1", Name: "college", IsCentral: 0},
		{Consortium: "c1", Name: "university", IsCentral: 0},
	}

	// Act
	result := tenants.String()

	// Assert
	assert.Equal(t, "college, university", result)
}
