package mvc

import (
	"github.com/gin-gonic/gin"
	"github.com/housepower/clickhouse_sinker/mvc/controller"
)

func PprofRouter(r *gin.RouterGroup) {
	c := controller.NewPprofController()
	r.GET("/metrics", c.Metrics)

	// 标准pprof路径
	r.GET("/pprof/", c.Index)
	r.GET("/pprof/cmdline", c.Cmdline)
	r.GET("/pprof/profile", c.Profile)
	r.GET("/pprof/symbol", c.Symbol)
	r.GET("/pprof/trace", c.Trace)
	r.GET("/pprof/ready", c.ReadyEndpoint)
	r.GET("/pprof/live", c.LiveEndpoint)

	// 添加debug路径兼容性
	debug := r.Group("/debug")
	debug.GET("/pprof/", c.Index)
	debug.GET("/pprof/cmdline", c.Cmdline)
	debug.GET("/pprof/profile", c.Profile)
	debug.GET("/pprof/symbol", c.Symbol)
	debug.GET("/pprof/trace", c.Trace)
	debug.GET("/pprof/heap", c.Heap)
	debug.GET("/pprof/goroutine", c.Goroutine)
}
