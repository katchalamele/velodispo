package domain

import "time"

type Status struct {
	StationID      string
	BikesAvailable int
	DocksAvailable int
	BikesDisabled  int
	DocksDisabled  int
	IsInstalled    bool
	IsRenting      bool
	IsReturning    bool
	LastReported   time.Time
}
