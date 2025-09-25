package controller

import (
	"net/http"
	"net/http/pprof"

	"github.com/gin-gonic/gin"
	"github.com/housepower/clickhouse_sinker/health"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PprofController struct {
}

func NewPprofController() *PprofController {
	return &PprofController{}
}

func (c *PprofController) Metrics(ctx *gin.Context) {
	promhttp.Handler().ServeHTTP(ctx.Writer, ctx.Request)
}

func (c *PprofController) ReadyEndpoint(ctx *gin.Context) {
	health.Health.ReadyEndpoint(ctx.Writer, ctx.Request)
}

func (c *PprofController) LiveEndpoint(ctx *gin.Context) {
	health.Health.LiveEndpoint(ctx.Writer, ctx.Request)
}

func (c *PprofController) Index(ctx *gin.Context) {
	pprof.Index(ctx.Writer, ctx.Request)
}

func (c *PprofController) Cmdline(ctx *gin.Context) {
	pprof.Cmdline(ctx.Writer, ctx.Request)
}

func (c *PprofController) Profile(ctx *gin.Context) {
	pprof.Profile(ctx.Writer, ctx.Request)
}

func (c *PprofController) Symbol(ctx *gin.Context) {
	pprof.Symbol(ctx.Writer, ctx.Request)
}

func (c *PprofController) Trace(ctx *gin.Context) {
	pprof.Trace(ctx.Writer, ctx.Request)
}

func (c *PprofController) Vars(ctx *gin.Context) {
	http.DefaultServeMux.ServeHTTP(ctx.Writer, ctx.Request)
}
