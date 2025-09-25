package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/housepower/clickhouse_sinker/task"
)

type MetricController struct {
	runner *task.Sinker
}

func NewMetricController(runner *task.Sinker) *MetricController {
	return &MetricController{
		runner: runner,
	}
}

func (c *MetricController) GetProcSummary(ctx *gin.Context) {
}
