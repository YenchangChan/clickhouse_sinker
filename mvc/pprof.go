package mvc

import (
	"github.com/gin-gonic/gin"
	"github.com/housepower/clickhouse_sinker/mvc/controller"
)

func PprofRouter(r *gin.RouterGroup) {
	c := controller.NewPprofController()
	r.GET("/metrics", c.Metrics)
	r.GET("/pprof/", c.Index)
	r.GET("/pprof/cmdline", c.Cmdline)
	r.GET("/pprof/profile", c.Profile)
	r.GET("/pprof/symbol", c.Symbol)
	r.GET("/pprof/trace", c.Trace)
	r.GET("/pprof/ready", c.ReadyEndpoint)
	r.GET("/pprof/live", c.LiveEndpoint)
}
