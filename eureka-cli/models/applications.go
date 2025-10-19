package models

type Applications struct {
	ApplicationDescriptors []map[string]any `json:"applicationDescriptors"`
	TotalRecords           int              `json:"totalRecords"`
}
