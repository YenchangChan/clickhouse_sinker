package mvc

import (
	"github.com/gin-gonic/gin"
	"github.com/housepower/clickhouse_sinker/mvc/controller"
	"github.com/housepower/clickhouse_sinker/task"
	"github.com/housepower/clickhouse_sinker/util"
)

func InitRouterV1(groupV1 *gin.RouterGroup, cmdOps util.CmdOptions, runner *task.Sinker, v util.VersionInfo) {
	cfgController := controller.NewConfigController(cmdOps, runner)
	taskController := controller.NewTaskController(runner)
	metricController := controller.NewMetricController(runner, v)

	groupV1.GET("/config", cfgController.GetConfig)
	groupV1.GET("/cmdline", cfgController.GetCmdLine)
	groupV1.GET("/log", cfgController.GetLog)
	groupV1.GET("/tasks", taskController.GetAllTasks)
	groupV1.GET("/task/:taskname", taskController.GetTaskByName)
	groupV1.GET("/dbkey/:taskname", taskController.GetDbKeyByTask)
	groupV1.POST("/tasks", taskController.GetTaskByCondition)
	groupV1.GET("/metrics/procinfo", metricController.GetProcSummary)
}
