package mvc

import (
	"context"
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

type Service struct {
	cmdOps util.CmdOptions
	runner *task.Sinker
	host   string
	port   int
	svr    *http.Server
}

func NewService(cmdOps util.CmdOptions, runner *task.Sinker, httpHost string, httpPort int) *Service {
	return &Service{
		cmdOps: cmdOps,
		runner: runner,
		host:   httpHost,
		port:   httpPort,
	}
}

func (s *Service) Start() (err error) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(ginLoggerToFile())
	r.Use(gin.CustomRecoveryWithWriter(nil, handlePanic))

	PprofRouter(r.Group("/"))

	groupApi := r.Group("/api")
	groupV1 := groupApi.Group("/v1")
	InitRouterV1(groupV1, s.cmdOps, s.runner)
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
