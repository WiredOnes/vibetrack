package model

// @Enum
type HealthStatus uint8

const (
	HealthStatusUnknown HealthStatus = iota
	HealthStatusServing
	HealthStatusNotServing
)
