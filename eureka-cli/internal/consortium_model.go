package internal

import (
	"fmt"
	"strings"
)

type ConsortiumTenant struct {
	Consortium string
	Tenant     string
	IsCentral  int
}

func (c ConsortiumTenant) String() string {
	if c.IsCentral == 1 {
		return fmt.Sprintf("%s (central)", c.Tenant)
	}

	return c.Tenant
}

type ConsortiumTenants []*ConsortiumTenant

func (c ConsortiumTenants) String() string {
	var builder strings.Builder
	for idx, value := range c {
		builder.WriteString(value.Tenant)
		if idx+1 < len(c) {
			builder.WriteString(", ")
		}
	}

	return builder.String()
}
