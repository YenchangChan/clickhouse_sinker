package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/housepower/clickhouse_sinker/mvc/model"
	"github.com/housepower/clickhouse_sinker/task"
	"github.com/housepower/clickhouse_sinker/util"
)

type ConfigContorller struct {
	cmdOps util.CmdOptions
	runner *task.Sinker
}

func NewConfigController(cmdOps util.CmdOptions, runner *task.Sinker) *ConfigContorller {
	return &ConfigContorller{
		cmdOps: cmdOps,
		runner: runner,
	}
}

func (c *ConfigContorller) GetConfig(ctx *gin.Context) {
	if c.runner == nil {
		model.WrapMsg(ctx, model.E_CONFIG_FAILED, nil)
		return
	}
	conf := c.runner.GetCurrentConfig()
	if conf == nil {
		model.WrapMsg(ctx, model.E_CONFIG_FAILED, nil)
		return
	}
	resp := model.ConfigResp{
		ClickHouse: &conf.Clickhouse,
		Discovery:  &conf.Discovery,
		Kafka:      &conf.Kafka,
		Tasks:      len(conf.Tasks),
	}
	model.WrapMsg(ctx, model.E_SUCCESS, resp)
}

func (c *ConfigContorller) GetCmdLine(ctx *gin.Context) {
	model.WrapMsg(ctx, model.E_SUCCESS, c.cmdOps)
}
