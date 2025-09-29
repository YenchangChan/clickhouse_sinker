package model

import (
	"github.com/housepower/clickhouse_sinker/config"
	"github.com/housepower/clickhouse_sinker/util"
)

type ConfigResp struct {
	ClickHouse *config.ClickHouseConfig
	Kafka      *config.KafkaConfig
	Discovery  *config.Discovery
	Tasks      int
}

type CmdLineResp struct {
	util.CmdOptions
}

type Task struct {
	Name          string
	Cluster       string
	Table         string
	Topic         string
	ConsumerGroup string
	Type          string
	ColPolicy     string
	Status        string
	Rate          int
	Lag           int64
	LastUpdate    int64
}

type TaskResp struct {
	Tasks []Task
	Total int
}

type TaskDetailResp struct {
	config.TaskConfig
}

type ProcInfoResp struct {
	Version    string
	BuildTime  string
	Commit     string
	GoVersion  string
	Goroutines int
	CPU        float64
	Memory     uint64
	StartTime  int64
	Uptime     int64
}
