package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/housepower/clickhouse_sinker/config"
	"github.com/housepower/clickhouse_sinker/mvc/model"
	"github.com/housepower/clickhouse_sinker/task"
)

type TaskController struct {
	runner *task.Sinker
}

func NewTaskController(runner *task.Sinker) *TaskController {
	return &TaskController{
		runner: runner,
	}
}

func (c *TaskController) GetAllTasks(ctx *gin.Context) {
	if c.runner == nil {
		model.WrapMsg(ctx, model.E_CONFIG_FAILED, nil)
		return
	}
	conf := c.runner.GetCurrentConfig()
	if conf == nil {
		model.WrapMsg(ctx, model.E_CONFIG_FAILED, nil)
		return
	}
	resp := model.TaskResp{
		Tasks: conf.Tasks,
		Total: len(conf.Tasks),
	}
	model.WrapMsg(ctx, model.E_SUCCESS, resp)
}

func (c *TaskController) GetTaskByName(ctx *gin.Context) {
	taskName := ctx.Param("taskname")
	if c.runner == nil {
		model.WrapMsg(ctx, model.E_CONFIG_FAILED, nil)
		return
	}
	conf := c.runner.GetCurrentConfig()
	if conf == nil {
		model.WrapMsg(ctx, model.E_CONFIG_FAILED, nil)
		return
	}
	var t *config.TaskConfig
	for _, task := range conf.Tasks {
		if task.Name == taskName {
			t = task
			break
		}
	}
	model.WrapMsg(ctx, model.E_SUCCESS, t)
}

func (c *TaskController) GetTaskByCondition(ctx *gin.Context) {
}

func (c *TaskController) GetDbKeyByTask(ctx *gin.Context) {
	if c.runner == nil {
		model.WrapMsg(ctx, model.E_CONFIG_FAILED, nil)
		return
	}
}
