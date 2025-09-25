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

type TaskResp struct {
	Tasks []*config.TaskConfig
	Total int
}

type TaskDetailResp struct {
	config.TaskConfig
}

type ProcInfoResp struct {
	Version    string
	Goroutines int
	CPU        float64
	Memory     uint64
	StartTime  int64
	Uptime     int64
}
