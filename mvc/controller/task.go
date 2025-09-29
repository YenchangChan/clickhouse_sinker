package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/housepower/clickhouse_sinker/config"
	cm "github.com/housepower/clickhouse_sinker/config_manager"
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

	statelags, err := cm.GetTaskStateAndLags(conf)
	if err != nil {
		model.WrapMsg(ctx, model.E_CONFIG_FAILED, nil)
		return
	}
	var tasks []model.Task
	for _, task := range conf.Tasks {
		tskType := "log"
		if task.PrometheusSchema {
			tskType = "metric"
		}
		colPolicy := "dims"
		if task.AutoSchema {
			colPolicy = "auto"
			if task.DynamicSchema.Enable {
				colPolicy = "dynamic"
			}
		}
		state := statelags[task.Name].State
		if state == "Preparing Rebalance" {
			state = "Rebalancing"
		} else if state == "Dead and Empty" {
			state = "Dead"
		}
		tasks = append(tasks, model.Task{
			Name:          task.Name,
			Cluster:       conf.Clickhouse.Cluster,
			Table:         task.TableName,
			Topic:         task.Topic,
			ConsumerGroup: task.ConsumerGroup,
			Type:          tskType,
			ColPolicy:     colPolicy,
			Status:        state,
			Lag:           statelags[task.Name].Lag,
		})
	}
	resp := model.TaskResp{
		Tasks: tasks,
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
