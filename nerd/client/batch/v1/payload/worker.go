package v1payload

import "time"

//WorkerCondition describes the worker status at a point in time
type WorkerCondition struct {
	ProbeTime time.Time `json:"probe_time"`
	Type      string    `json:"type"`
}

//WorkerSummary is a small version
type WorkerSummary struct {
	WorkerID   string             `json:"worker_id"`
	Status     string             `json:"status"`
	Conditions []*WorkerCondition `json:"conditions"`
}
