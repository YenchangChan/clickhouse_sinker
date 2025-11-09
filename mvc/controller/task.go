package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/housepower/clickhouse_sinker/config"
	cm "github.com/housepower/clickhouse_sinker/config_manager"
	model2 "github.com/housepower/clickhouse_sinker/model"
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
		task := model.Task{
			Name:          task.Name,
			Cluster:       conf.Clickhouse.Cluster,
			Table:         task.TableName,
			Topic:         task.Topic,
			ConsumerGroup: task.ConsumerGroup,
			Type:          tskType,
			ColPolicy:     colPolicy,
			Status:        state,
			Lag:           statelags[task.Name].Lag,
		}

		consumer := c.runner.Consumer(task.ConsumerGroup)
		if consumer != nil {
			task.DbKey = consumer.DbMap()
			if task.DbKey == nil {
				task.DbKey = make(map[string]*model2.DbState)
			}
		}
		service := consumer.GetTask(task.Name)
		if service != nil {
			task.Rate = int64(service.Meter().Rate1())
		}

		tasks = append(tasks, task)
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

	consumer := c.runner.Consumer(t.ConsumerGroup)
	var resp []model.DbKeyResp
	if consumer != nil {
		dbKey := consumer.DbMap()
		for _, v := range dbKey {
			resp = append(resp, model.DbKeyResp{
				IdxSerID:   v.IdxSerID,
				Name:       v.DB,
				NumDims:    v.NumDims,
				PrepareSQL: v.PrepareSQL,
				Processed:  v.Processed,
				PromSerSQL: v.PromSerSQL,
			})
		}
	}
	model.WrapMsg(ctx, model.E_SUCCESS, resp)
}
