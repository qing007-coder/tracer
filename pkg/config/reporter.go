package config

import "time"

type ReporterConfig struct {
	QueueSize uint          `json:"queue_size"`
	Duration  time.Duration `json:"duration"`
	AgentAddr string        `json:"agent_addr"`
}
