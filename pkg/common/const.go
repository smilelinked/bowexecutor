// Package common used to store constants, data conversion functions, timers, etc
package common

const (
	// APIVersion description API version
	APIVersion = "v1"
)

// ErrKind define the error's type
type ErrKind string

type DeviceStatus string

const (
	StatusReady      DeviceStatus = "Ready"
	StatusSyncing    DeviceStatus = "Syncing"
	StatusExecucting DeviceStatus = "Executing"
)
