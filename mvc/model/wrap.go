package model

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/housepower/clickhouse_sinker/util"
	"go.uber.org/zap"
)

type ResponseBody struct {
	RetCode string      `json:"retCode"`
	RetMsg  string      `json:"retMsg"`
	Entity  interface{} `json:"entity"`
}

func WrapMsg(c *gin.Context, retCode string, entity interface{}) {
	c.Status(http.StatusOK)
	c.Header("Content-Type", "application/json; charset=utf-8")

	retMsg := GetMsg(c, retCode)
	if retCode != E_SUCCESS {
		util.Logger.Error(retMsg, zap.String("method", c.Request.Method), zap.String("uri", c.Request.RequestURI), zap.String("code", retCode), zap.Any("entity", entity))
		if err, ok := entity.(error); ok {
			retMsg += ": " + err.Error()
		} else if s, ok := entity.(string); ok {
			retMsg += ": " + s
		}
		entity = nil
	}

	resp := ResponseBody{
		RetCode: retCode,
		RetMsg:  retMsg,
		Entity:  entity,
	}
	jsonBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		util.Logger.Error("marshal response body fail", zap.Error(err))
		return
	}

	util.Logger.Debug("[response]", zap.String("host", c.Request.Host), zap.String("method", c.Request.Method), zap.String("url", c.Request.URL.String()), zap.String("body", string(jsonBytes)))

	_, err = c.Writer.Write(jsonBytes)
	if err != nil {
		util.Logger.Error("write response body fail", zap.Error(err))
		return
	}
}
