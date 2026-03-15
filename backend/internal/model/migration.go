package model

import "time"

// @PublicValueInstance
type Migration struct {
	Version   uint
	Name      string
	Query     string
	AppliedAt time.Time
}
