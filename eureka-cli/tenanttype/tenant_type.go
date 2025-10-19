package tenanttype

type TenantType string

const (
	All     TenantType = ""
	Default TenantType = "default"
	Central TenantType = "central"
	Member  TenantType = "member"
)

func Get() []TenantType {
	return []TenantType{Central, Member}
}
