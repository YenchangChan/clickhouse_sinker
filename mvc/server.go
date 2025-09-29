package mvc

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/housepower/clickhouse_sinker/mvc/model"
	"github.com/housepower/clickhouse_sinker/task"
	"github.com/housepower/clickhouse_sinker/util"
	"go.uber.org/zap"
)

//go:embed dist/*
var staticFiles embed.FS

type Service struct {
	cmdOps  util.CmdOptions
	runner  *task.Sinker
	host    string
	port    int
	svr     *http.Server
	version util.VersionInfo
}

func NewService(cmdOps util.CmdOptions, runner *task.Sinker, httpHost string, httpPort int, v util.VersionInfo) *Service {
	return &Service{
		cmdOps:  cmdOps,
		runner:  runner,
		host:    httpHost,
		port:    httpPort,
		version: v,
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

	// 静态文件服务 - 使用不冲突的方式配置
	s.setupStaticRoutes(r)
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

func (s *Service) setupStaticRoutes(r *gin.Engine) {
	// 配置静态文件服务到内部路径，但不直接暴露给用户
	// 尝试使用嵌入的静态文件
	distFS, err := fs.Sub(staticFiles, "dist")
	if err == nil {
		// 注册静态文件系统，但不直接挂载到根路径
		staticFS := http.FS(distFS)

		// 为根路径添加处理函数，返回index.html
		r.GET("/", func(c *gin.Context) {
			f, err := staticFS.Open("index.html")
			if err != nil {
				c.String(http.StatusNotFound, "Not found")
				return
			}
			defer f.Close()

			c.DataFromReader(http.StatusOK, -1, "text/html; charset=utf-8", f, nil)
		})

		// 为所有其他路径提供静态文件服务，但避免与API和pprof路由冲突
		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			// 避免处理API和pprof相关路径
			if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/metrics") ||
				strings.HasPrefix(path, "/pprof/") || strings.HasPrefix(path, "/debug/") {
				c.String(http.StatusNotFound, "Not found")
				return
			}

			// 尝试打开请求的文件
			f, err := staticFS.Open(path[1:]) // 去掉前导斜杠
			if err != nil {
				// 如果文件不存在，返回index.html以支持单页应用路由
				f, err := staticFS.Open("index.html")
				if err != nil {
					c.String(http.StatusNotFound, "Not found")
					return
				}
				defer f.Close()
				c.DataFromReader(http.StatusOK, -1, "text/html; charset=utf-8", f, nil)
				return
			}
			defer f.Close()

			// 获取文件信息以确定内容类型
			fi, err := f.Stat()
			if err != nil {
				c.String(http.StatusInternalServerError, "Internal server error")
				return
			}

			// 设置适当的内容类型
			contentType := "application/octet-stream"
			ext := filepath.Ext(path)
			switch ext {
			case ".html":
				contentType = "text/html; charset=utf-8"
			case ".css":
				contentType = "text/css"
			case ".js":
				contentType = "application/javascript"
			case ".json":
				contentType = "application/json"
			case ".png":
				contentType = "image/png"
			case ".jpg", ".jpeg":
				contentType = "image/jpeg"
			case ".gif":
				contentType = "image/gif"
			case ".svg":
				contentType = "image/svg+xml"
			}

			c.DataFromReader(http.StatusOK, fi.Size(), contentType, f, nil)
		})

		util.Logger.Info("Using embedded static files")
	} else {
		// 回退到文件系统
		staticDir := http.Dir("./mvc/static")

		// 为根路径添加处理函数，返回index.html
		r.GET("/", func(c *gin.Context) {
			f, err := staticDir.Open("index.html")
			if err != nil {
				c.String(http.StatusNotFound, "Not found")
				return
			}
			defer f.Close()

			c.DataFromReader(http.StatusOK, -1, "text/html; charset=utf-8", f, nil)
		})

		// 为所有其他路径提供静态文件服务，但避免与API和pprof路由冲突
		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			// 避免处理API和pprof相关路径
			if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/metrics") ||
				strings.HasPrefix(path, "/pprof/") || strings.HasPrefix(path, "/debug/") {
				c.String(http.StatusNotFound, "Not found")
				return
			}

			// 尝试打开请求的文件
			f, err := staticDir.Open(path[1:]) // 去掉前导斜杠
			if err != nil {
				// 如果文件不存在，返回index.html以支持单页应用路由
				f, err := staticDir.Open("index.html")
				if err != nil {
					c.String(http.StatusNotFound, "Not found")
					return
				}
				defer f.Close()
				c.DataFromReader(http.StatusOK, -1, "text/html; charset=utf-8", f, nil)
				return
			}
			defer f.Close()

			// 获取文件信息以确定内容类型
			fi, err := f.Stat()
			if err != nil {
				c.String(http.StatusInternalServerError, "Internal server error")
				return
			}

			// 设置适当的内容类型
			contentType := "application/octet-stream"
			ext := filepath.Ext(path)
			switch ext {
			case ".html":
				contentType = "text/html; charset=utf-8"
			case ".css":
				contentType = "text/css"
			case ".js":
				contentType = "application/javascript"
			case ".json":
				contentType = "application/json"
			case ".png":
				contentType = "image/png"
			case ".jpg", ".jpeg":
				contentType = "image/jpeg"
			case ".gif":
				contentType = "image/gif"
			case ".svg":
				contentType = "image/svg+xml"
			}

			c.DataFromReader(http.StatusOK, fi.Size(), contentType, f, nil)
		})

		util.Logger.Info("Using filesystem static files")
	}
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
