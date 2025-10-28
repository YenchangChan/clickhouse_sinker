package controller

import (
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/housepower/clickhouse_sinker/mvc/model"
	"github.com/housepower/clickhouse_sinker/task"
	"github.com/housepower/clickhouse_sinker/util"
)

type MetricController struct {
	runner  *task.Sinker
	version util.VersionInfo
}

func NewMetricController(runner *task.Sinker, v util.VersionInfo) *MetricController {
	return &MetricController{
		runner:  runner,
		version: v,
	}
}

func (c *MetricController) GetProcSummary(ctx *gin.Context) {
	if c.runner == nil {
		model.WrapMsg(ctx, model.E_CONFIG_FAILED, nil)
		return
	}

	// 获取系统信息
	resp := model.ProcInfoResp{
		Version:    c.version.Version, // 可以从build info获取
		GoVersion:  c.version.GoVersion,
		BuildTime:  c.version.BuildTime,
		Commit:     c.version.Commit,
		Goroutines: runtime.NumGoroutine(),
		CPU:        0.0, // 需要实现CPU使用率计算
		Memory:     0,   // 需要实现内存使用计算
	}

	// 获取内存统计
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	resp.Memory = m.Alloc

	// 获取CPU使用率
	if cpu, err := util.GetCPUUsage(); err == nil {
		resp.CPU = cpu
	}

	// 获取进程启动时间
	resp.StartTime = util.GetProcessStartTime()
	resp.Uptime = time.Now().Unix() - resp.StartTime

	conf := c.runner.GetCurrentConfig()
	if conf != nil {
		resp.Tasks = len(conf.Tasks)
	}

	model.WrapMsg(ctx, model.E_SUCCESS, resp)
}
