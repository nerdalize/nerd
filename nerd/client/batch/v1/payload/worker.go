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

//WorkerLogsInput is for fetching worker logs
type WorkerLogsInput struct {
	ProjectID  string `json:"project_id"`
	WorkloadID string `json:"workload_id"`
	WorkerID   string `json:"worker_id"`
}

//WorkerLogsOutput contains raw log data from the cluster
type WorkerLogsOutput struct {
	Data []byte `json:"data"`
}
