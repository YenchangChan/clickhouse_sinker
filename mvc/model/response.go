package model

import (
	"github.com/housepower/clickhouse_sinker/config"
	"github.com/housepower/clickhouse_sinker/model"
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
	DbKey         map[string]*model.DbState
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
	Version        string
	BuildTime      string
	Commit         string
	GoVersion      string
	RecordPoolSize int64
	Goroutines     int
	CPU            float64
	Memory         uint64
	StartTime      int64
	Uptime         int64
}
