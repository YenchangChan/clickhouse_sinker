package mvc

import (
	"context"
	"embed"
	"fmt"
	"net"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/housepower/clickhouse_sinker/mvc/model"
	"github.com/housepower/clickhouse_sinker/task"
	"github.com/housepower/clickhouse_sinker/util"
	"go.uber.org/zap"
)

//go:embed static
var embedFs embed.FS

type Service struct {
	cmdOps  util.CmdOptions
	runner  *task.Sinker
	host    string
	port    int
	svr     *http.Server
	version util.VersionInfo
	modTime time.Time
}

func NewService(cmdOps util.CmdOptions, runner *task.Sinker, httpHost string, httpPort int, v util.VersionInfo) *Service {
	modTime, err := time.Parse(time.RFC3339, v.BuildTime)
	if err != nil {
		modTime = time.Now()
	}
	return &Service{
		cmdOps:  cmdOps,
		runner:  runner,
		host:    httpHost,
		port:    httpPort,
		version: v,
		modTime: modTime,
	}
}

func (s *Service) Start() (err error) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(ginLoggerToFile())
	r.Use(gin.CustomRecoveryWithWriter(nil, handlePanic))

	// API路由
	groupApi := r.Group("/api")
	groupV1 := groupApi.Group("/v1")
	InitRouterV1(groupV1, s.cmdOps, s.runner, s.version)

	// pprof路由 - 直接挂载到根路径的特定子路径
	PprofRouter(r.Group(""))

	r.Use(Serve("/", EmbedFolder(embedFs, "static"), s.modTime))
	r.NoRoute(func(c *gin.Context) {
		data, err := embedFs.ReadFile("static/index.html")
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})
	bind := net.JoinHostPort(s.host, fmt.Sprintf("%d", s.port))
	s.svr = &http.Server{
		Addr:         bind,
		WriteTimeout: time.Second * 3600,
		ReadTimeout:  time.Second * 3600,
		IdleTimeout:  time.Second * 3600,
		Handler:      r,
	}

	go func() {
		if err := s.svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			util.Logger.Error("http.ListenAndServe failed", zap.Error(err))
		} else {
			util.Logger.Info(fmt.Sprintf("Run http server at http://%s/", bind))
		}
	}()

	return nil
}

func (s *Service) Stop() error {
	waitTimeout := time.Duration(time.Second * 10)
	ctx, cancel := context.WithTimeout(context.Background(), waitTimeout)
	defer cancel()
	return s.svr.Shutdown(ctx)
}

// Log runtime error stack to make debug easy.
func handlePanic(c *gin.Context, err interface{}) {
	util.Logger.Error("server panic", zap.Reflect("error", err),
		zap.String("stack", string(debug.Stack())))
	model.WrapMsg(c, model.E_UNKNOWN, err)
}

// Replace gin.Logger middleware to customize log format.
func ginLoggerToFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		// start time
		startTime := time.Now()
		// Processing request
		c.Next()
		// End time
		endTime := time.Now()
		// execution time
		latencyTime := endTime.Sub(startTime)
		// Request mode
		reqMethod := c.Request.Method
		// Request routing
		reqUri := c.Request.RequestURI
		// Status code
		statusCode := c.Writer.Status()
		// Request IP
		clientIP := c.ClientIP()
		// Log format
		msg := fmt.Sprintf("| %3d | %13v | %15s | %s | %s",
			statusCode,
			latencyTime,
			clientIP,
			reqMethod,
			reqUri)
		if statusCode == 200 {
			util.Logger.Info(msg)
		} else {
			util.Logger.Error(msg)
		}
	}
}
