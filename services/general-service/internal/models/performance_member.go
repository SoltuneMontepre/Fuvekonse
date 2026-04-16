package models

// PerformanceMemberInfo is one named participant (performer, panelist, etc.).
type PerformanceMemberInfo struct {
	Name   string `json:"name"`
	Detail string `json:"detail,omitempty"`
}
