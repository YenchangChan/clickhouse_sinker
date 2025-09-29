/*Copyright [2019] housepower

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	cm "github.com/housepower/clickhouse_sinker/config_manager"
	"github.com/housepower/clickhouse_sinker/mvc"
	"github.com/housepower/clickhouse_sinker/task"
	"github.com/housepower/clickhouse_sinker/util"
	"go.uber.org/zap"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// goreleaser fills the following info per https://goreleaser.com/customization/build/.
	version = "None"
	commit  = "None"
	date    = "None"
	builtBy = "None"

	cmdOps      util.CmdOptions
	httpAddr    string
	httpMetrics = promhttp.Handler()
	runner      *task.Sinker
	server      *mvc.Service
	v           = util.VersionInfo{
		Version:   version,
		BuildTime: date,
		Commit:    commit,
		GoVersion: runtime.Version(),
	}
)

func initCmdOptions() {
	// 1. Set options to default value.
	cmdOps = util.CmdOptions{
		LogLevel:      "info",
		LogPaths:      "stdout,clickhouse_sinker.log",
		PushInterval:  10,
		LocalCfgFile:  "/etc/clickhouse_sinker.hjson",
		NacosAddr:     "127.0.0.1:8848",
		NacosGroup:    "DEFAULT_GROUP",
		NacosUsername: "nacos",
		NacosPassword: "nacos",
		Encrypt:       "",
	}

	// 2. Replace options with the corresponding env variable if present.
	util.EnvBoolVar(&cmdOps.ShowVer, "v")
	util.EnvStringVar(&cmdOps.Encrypt, "e")
	util.EnvStringVar(&cmdOps.LogLevel, "log-level")
	util.EnvStringVar(&cmdOps.LogPaths, "log-paths")
	util.EnvIntVar(&cmdOps.HTTPPort, "http-port")
	util.EnvStringVar(&cmdOps.HTTPHost, "http-host")
	util.EnvStringVar(&cmdOps.PushGatewayAddrs, "metric-push-gateway-addrs")
	util.EnvIntVar(&cmdOps.PushInterval, "push-interval")
	util.EnvStringVar(&cmdOps.LocalCfgFile, "local-cfg-file")

	util.EnvStringVar(&cmdOps.NacosAddr, "nacos-addr")
	util.EnvStringVar(&cmdOps.NacosUsername, "nacos-username")
	util.EnvStringVar(&cmdOps.NacosPassword, "nacos-password")
	util.EnvStringVar(&cmdOps.NacosNamespaceID, "nacos-namespace-id")
	util.EnvStringVar(&cmdOps.NacosGroup, "nacos-group")
	util.EnvStringVar(&cmdOps.NacosDataID, "nacos-dataid")
	util.EnvStringVar(&cmdOps.NacosCommonNamespaceID, "nacos-common-namespace-id")
	util.EnvStringVar(&cmdOps.NacosCommonGroup, "nacos-common-group")
	util.EnvStringVar(&cmdOps.NacosCommonDataID, "nacos-common-dataid")
	util.EnvStringVar(&cmdOps.NacosServiceName, "nacos-service-name")

	util.EnvStringVar(&cmdOps.ClickhouseUsername, "clickhouse-username")
	util.EnvStringVar(&cmdOps.ClickhousePassword, "clickhouse-password")
	util.EnvStringVar(&cmdOps.KafkaUsername, "kafka-username")
	util.EnvStringVar(&cmdOps.KafkaPassword, "kafka-password")
	util.EnvStringVar(&cmdOps.KafkaGSSAPIUsername, "kafka-gssapi-username")
	util.EnvStringVar(&cmdOps.KafkaGSSAPIPassword, "kafka-gssapi-password")

	// 3. Replace options with the corresponding CLI parameter if present.
	flag.BoolVar(&cmdOps.ShowVer, "v", cmdOps.ShowVer, "show build version and quit")
	flag.StringVar(&cmdOps.Encrypt, "e", cmdOps.Encrypt, "encrypt password")
	flag.StringVar(&cmdOps.LogLevel, "log-level", cmdOps.LogLevel, "one of debug, info, warn, error, dpanic, panic, fatal")
	flag.StringVar(&cmdOps.LogPaths, "log-paths", cmdOps.LogPaths, "a list of comma-separated log file path. stdout means the console stdout")
	flag.IntVar(&cmdOps.HTTPPort, "http-port", cmdOps.HTTPPort, "http listen port")
	flag.StringVar(&cmdOps.HTTPHost, "http-host", cmdOps.HTTPHost, "http host to bind to")
	flag.StringVar(&cmdOps.PushGatewayAddrs, "metric-push-gateway-addrs", cmdOps.PushGatewayAddrs, "a list of comma-separated prometheus push gatway address")
	flag.IntVar(&cmdOps.PushInterval, "push-interval", cmdOps.PushInterval, "push interval in seconds")
	flag.StringVar(&cmdOps.LocalCfgFile, "local-cfg-file", cmdOps.LocalCfgFile, "local config file")

	flag.StringVar(&cmdOps.NacosAddr, "nacos-addr", cmdOps.NacosAddr, "a list of comma-separated nacos server addresses")
	flag.StringVar(&cmdOps.NacosUsername, "nacos-username", cmdOps.NacosUsername, "nacos username")
	flag.StringVar(&cmdOps.NacosPassword, "nacos-password", cmdOps.NacosPassword, "nacos password")
	flag.StringVar(&cmdOps.NacosNamespaceID, "nacos-namespace-id", cmdOps.NacosNamespaceID,
		`nacos namespace ID. Neither DEFAULT_NAMESPACE_ID("public") nor namespace name work! When namespace is 'public', fill in the blank string here!`)
	flag.StringVar(&cmdOps.NacosGroup, "nacos-group", cmdOps.NacosGroup, `nacos group name. Empty string doesn't work!`)
	flag.StringVar(&cmdOps.NacosDataID, "nacos-dataid", cmdOps.NacosDataID, "nacos dataid")
	flag.StringVar(&cmdOps.NacosCommonNamespaceID, "nacos-common-namespace-id", cmdOps.NacosCommonNamespaceID, "nacos common namespace id")
	flag.StringVar(&cmdOps.NacosCommonGroup, "nacos-common-group", cmdOps.NacosCommonGroup, "nacos common group")
	flag.StringVar(&cmdOps.NacosCommonDataID, "nacos-common-dataid", cmdOps.NacosCommonDataID, "nacos common dataid")
	flag.StringVar(&cmdOps.NacosServiceName, "nacos-service-name", cmdOps.NacosServiceName, "nacos service name")

	flag.StringVar(&cmdOps.ClickhouseUsername, "clickhouse-username", cmdOps.ClickhouseUsername, "clickhouse username")
	flag.StringVar(&cmdOps.ClickhousePassword, "clickhouse-password", cmdOps.ClickhousePassword, "clickhouse password")
	flag.StringVar(&cmdOps.KafkaUsername, "kafka-username", cmdOps.KafkaUsername, "kafka username")
	flag.StringVar(&cmdOps.KafkaPassword, "kafka-password", cmdOps.KafkaPassword, "kafka password")
	flag.StringVar(&cmdOps.KafkaGSSAPIUsername, "kafka-gssapi-username", cmdOps.KafkaGSSAPIUsername, "kafka GSSAPI username")
	flag.StringVar(&cmdOps.KafkaGSSAPIPassword, "kafka-gssapi-password", cmdOps.KafkaGSSAPIPassword, "kafka GSSAPI password")

	flag.Parse()
	if err := util.Gsypt.Unmarshal(&cmdOps); err != nil {
		util.Logger.Fatal("failed to decrypt password", zap.Error(err))
	}
}

func getVersion(v util.VersionInfo) string {
	return fmt.Sprintf("version %s, commit %s, date %s, go-version %s , pid %v", v.Version, v.Commit, v.BuildTime, v.GoVersion, os.Getpid())
}

func init() {
	initCmdOptions()
	if cmdOps.Encrypt != "" {
		fmt.Println(util.AesEncryptECB(cmdOps.Encrypt))
		os.Exit(0)
	}
	logPaths := strings.Split(cmdOps.LogPaths, ",")
	util.InitLogger(logPaths)
	util.SetLogLevel(cmdOps.LogLevel)
	util.Logger.Info(getVersion(v))
	if cmdOps.ShowVer {
		os.Exit(0)
	}
	util.Logger.Info("parsed command options:", zap.Reflect("opts", cmdOps))
}

func main() {
	util.Run("clickhouse_sinker", func() error {
		httpHost := cmdOps.HTTPHost
		if httpHost == "" {
			ip, err := util.GetOutboundIP()
			if err != nil {
				return fmt.Errorf("failed to determine outbound ip: %w", err)
			}
			httpHost = ip.String()
		}

		httpPort := cmdOps.HTTPPort
		if httpPort == 0 {
			httpPort = util.HttpPortBase
		}
		httpPort = util.GetSpareTCPPort(httpPort)

		var rcm cm.RemoteConfManager
		var properties map[string]interface{}
		logDir := "."
		logPaths := strings.Split(cmdOps.LogPaths, ",")
		for _, logPath := range logPaths {
			if logPath != "stdout" && logPath != "stderr" {
				logDir, _ = filepath.Split(logPath)
			}
		}
		logDir, _ = filepath.Abs(logDir)
		if cmdOps.NacosDataID != "" {
			util.Logger.Info(fmt.Sprintf("get config from nacos serverAddrs %s, namespaceId %s, group %s, dataId %s",
				cmdOps.NacosAddr, cmdOps.NacosNamespaceID, cmdOps.NacosGroup, cmdOps.NacosDataID))
			rcm = &cm.NacosConfManager{}
			properties = make(map[string]interface{}, 8)
			properties["serverAddrs"] = cmdOps.NacosAddr
			properties["username"] = cmdOps.NacosUsername
			properties["password"] = cmdOps.NacosPassword
			properties["namespaceId"] = cmdOps.NacosNamespaceID
			properties["group"] = cmdOps.NacosGroup
			properties["dataId"] = cmdOps.NacosDataID
			properties["commonNamespaceId"] = cmdOps.NacosCommonNamespaceID
			properties["commonGroup"] = cmdOps.NacosCommonGroup
			properties["commonDataId"] = cmdOps.NacosCommonDataID
			properties["serviceName"] = cmdOps.NacosServiceName
			properties["logDir"] = logDir
		} else {
			util.Logger.Info(fmt.Sprintf("get config from local file %s", cmdOps.LocalCfgFile))
		}
		if rcm != nil {
			if err := rcm.Init(properties); err != nil {
				util.Logger.Fatal("rcm.Init failed", zap.Error(err))
			}
			if cmdOps.NacosServiceName != "" {
				if err := rcm.Register(httpHost, httpPort); err != nil {
					util.Logger.Fatal("rcm.Init failed", zap.Error(err))
				}
			}
		}
		runner = task.NewSinker(rcm, httpAddr, &cmdOps)
		// cmdOps.HTTPPort=0: disable the http server
		if cmdOps.HTTPPort > 0 {
			ops := cmdOps
			if rcm == nil {
				ops.LocalCfgFile = ""
			}
			server = mvc.NewService(ops, runner, httpHost, httpPort, v)
			err := server.Start()
			if err != nil {
				return fmt.Errorf("failed to start http server: %w", err)
			}
		}
		return runner.Init()
	}, func() error {
		runner.Run()
		return nil
	}, func() error {
		runner.Close()
		if server != nil {
			server.Stop()
		}
		return nil
	})
}
