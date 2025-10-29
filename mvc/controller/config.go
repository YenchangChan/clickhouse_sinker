package controller

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"

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
	}
	model.WrapMsg(ctx, model.E_SUCCESS, resp)
}

func (c *ConfigContorller) GetCmdLine(ctx *gin.Context) {
	model.WrapMsg(ctx, model.E_SUCCESS, c.cmdOps)
}

func (c *ConfigContorller) GetLog(ctx *gin.Context) {
	fileName := "clickhouse_sinker.log"
	fromParam := ctx.Query("from")
	errorLevel := ctx.Query("error")
	errlvl := false
	from := 1
	if fromParam != "" {
		from, _ = strconv.Atoi(fromParam)
	}
	if errorLevel == "true" {
		errlvl = true
	}
	to := from + 100
	lines, total, err := ReadLines(fileName, from, to, errlvl)
	if err != nil {
		model.WrapMsg(ctx, model.E_CONFIG_FAILED, err)
		return
	}
	resp := model.LogResp{
		Total: total,
		Lines: lines,
	}
	model.WrapMsg(ctx, model.E_SUCCESS, resp)
}

func ReadLines(filename string, from, to int, errlvl bool) (lines []string, total int, err error) {
	var fi *os.File
	fi, err = os.Open(filename)
	if err != nil {
		return
	}
	defer fi.Close()
	// 输出从from到to行的内容，并计算所有行号
	scanner := bufio.NewScanner(fi)
	total = 0
	for scanner.Scan() {
		line := scanner.Text()
		if errlvl {
			var m map[string]interface{}
			_ = json.Unmarshal([]byte(line), &m)
			lvl, ok := m["level"].(string)
			if ok && (lvl == "warn" || lvl == "error" || lvl == "fatal") {
				total++
				if total >= from && total <= to {
					lines = append(lines, line)
				}
			}
		} else {
			total++
			if total >= from && total <= to {
				lines = append(lines, line)
			}
		}
	}
	return
}
