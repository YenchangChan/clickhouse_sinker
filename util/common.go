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

package util

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/thanos-io/thanos/pkg/errors"
)

var (
	Logger       *zap.Logger
	logAtomLevel zap.AtomicLevel
	logPaths     []string

	startTime time.Time
)

func init() {
	startTime = time.Now()
}

type CmdOptions struct {
	ShowVer  bool
	LogLevel string // "debug", "info", "warn", "error", "dpanic", "panic", "fatal"
	LogPaths string // comma-separated paths. "stdout" means the console stdout

	// HTTPHost to bind to. If empty, outbound ip of machine
	// is automatically determined and used.
	HTTPHost string
	HTTPPort int // 0 means a randomly chosen port.

	PushGatewayAddrs string
	PushInterval     int
	LocalCfgFile     string
	NacosAddr        string
	NacosNamespaceID string
	NacosGroup       string
	NacosUsername    string
	NacosPassword    string
	NacosDataID      string
	NacosServiceName string // participate in assignment management if not empty
	Encrypt          string

	Credentials
	CommonConf
}

type CommonConf struct {
	NacosCommonNamespaceID string
	NacosCommonGroup       string
	NacosCommonDataID      string
}

type Credentials struct {
	ClickhouseUsername  string
	ClickhousePassword  string
	KafkaUsername       string
	KafkaPassword       string
	KafkaGSSAPIUsername string
	KafkaGSSAPIPassword string
}

// StringContains check if contains string in array
func StringContains(arr []string, str string) bool {
	for _, s := range arr {
		if s == str {
			return true
		}
	}
	return false
}

// GetSourceName returns the field name in message for the given ClickHouse column
func GetSourceName(parser, name string) (sourcename string) {
	if parser == "gjson" {
		sourcename = strings.Replace(name, ".", "\\.", -1)
	} else {
		sourcename = name
	}
	return
}

// GetShift returns the smallest `shift` which 1<<shift is no smaller than s
func GetShift(s int) (shift uint) {
	for shift = 0; (1 << shift) < s; shift++ {
	}
	return
}

// Refers to:
// https://medium.com/processone/using-tls-authentication-for-your-go-kafka-client-3c5841f2a625
// https://github.com/denji/golang-tls
// https://www.baeldung.com/java-keystore-truststore-difference
func NewTLSConfig(caCertFiles, clientCertFile, clientKeyFile string, insecureSkipVerify bool) (*tls.Config, error) {
	tlsConfig := tls.Config{}
	// Load client cert
	if clientCertFile != "" && clientKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
		if err != nil {
			err = errors.Wrapf(err, "")
			return &tlsConfig, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// Load CA cert if it exists.  Not needed for OS trusted certs
	if caCertFiles != "" {
		caCertPool := x509.NewCertPool()
		for _, caCertFile := range strings.Split(caCertFiles, ",") {
			caCert, err := os.ReadFile(caCertFile)
			if err != nil {
				err = errors.Wrapf(err, "")
				return &tlsConfig, err
			}
			caCertPool.AppendCertsFromPEM(caCert)
		}
		tlsConfig.RootCAs = caCertPool
	}
	tlsConfig.InsecureSkipVerify = insecureSkipVerify
	return &tlsConfig, nil
}

func EnvStringVar(value *string, key string) {
	realKey := strings.ReplaceAll(strings.ToUpper(key), "-", "_")
	val, found := os.LookupEnv(realKey)
	if found {
		*value = val
	}
}

func EnvIntVar(value *int, key string) {
	realKey := strings.ReplaceAll(strings.ToUpper(key), "-", "_")
	val, found := os.LookupEnv(realKey)
	if found {
		valInt, err := strconv.Atoi(val)
		if err == nil {
			*value = valInt
		}
	}
}

func EnvBoolVar(value *bool, key string) {
	realKey := strings.ReplaceAll(strings.ToUpper(key), "-", "_")
	if _, found := os.LookupEnv(realKey); found {
		*value = true
	}
}

// JksToPem converts JKS to PEM
// Refers to:
// https://serverfault.com/questions/715827/how-to-generate-key-and-crt-file-from-jks-file-for-httpd-apache-server
func JksToPem(jksPath, jksPassword string, overwrite bool) (certPemPath, keyPemPath string, err error) {
	dir, fn := filepath.Split(jksPath)
	certPemPath = filepath.Join(dir, fn+".cert.pem")
	keyPemPath = filepath.Join(dir, fn+".key.pem")
	pkcs12Path := filepath.Join(dir, fn+".p12")
	if overwrite {
		for _, fp := range []string{certPemPath, keyPemPath, pkcs12Path} {
			if err = os.RemoveAll(fp); err != nil {
				err = errors.Wrapf(err, "")
				return
			}
		}
	} else {
		for _, fp := range []string{certPemPath, keyPemPath, pkcs12Path} {
			if _, err = os.Stat(fp); err == nil {
				return
			}
		}
	}
	cmds := [][]string{
		{"keytool", "-importkeystore", "-srckeystore", jksPath, "-destkeystore", pkcs12Path, "-deststoretype", "PKCS12"},
		{"openssl", "pkcs12", "-in", pkcs12Path, "-nokeys", "-out", certPemPath, "-passin", "env:password"},
		{"openssl", "pkcs12", "-in", pkcs12Path, "-nodes", "-nocerts", "-out", keyPemPath, "-passin", "env:password"},
	}
	for _, cmd := range cmds {
		if Logger != nil {
			Logger.Info(strings.Join(cmd, " "))
		}
		exe := exec.Command(cmd[0], cmd[1:]...)
		if cmd[0] == "keytool" {
			exe.Stdin = bytes.NewReader([]byte(jksPassword + "\n" + jksPassword + "\n" + jksPassword))
		} else if cmd[0] == "openssl" {
			exe.Env = []string{fmt.Sprintf("password=%s", jksPassword)}
		}
		var out []byte
		out, err = exe.CombinedOutput()
		if Logger != nil {
			Logger.Info(string(out))
		}
		if err != nil {
			err = errors.Wrapf(err, "")
			return
		}
	}
	return
}

func InitLogger(newLogPaths []string) {
	if reflect.DeepEqual(logPaths, newLogPaths) {
		return
	}
	logAtomLevel = zap.NewAtomicLevel()
	logPaths = newLogPaths
	var syncers []zapcore.WriteSyncer
	for _, p := range logPaths {
		switch p {
		case "stdout":
			syncers = append(syncers, zapcore.AddSync(os.Stdout))
		case "stderr":
			syncers = append(syncers, zapcore.AddSync(os.Stderr))
		default:
			writeFile := zapcore.AddSync(&lumberjack.Logger{
				Filename:   p,
				MaxSize:    100, // megabytes
				MaxBackups: 10,
				LocalTime:  true,
			})
			syncers = append(syncers, writeFile)
		}
	}

	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(cfg),
		zapcore.NewMultiWriteSyncer(syncers...),
		logAtomLevel,
	)
	Logger = zap.New(core, zap.AddStacktrace(zap.ErrorLevel), zap.AddCaller())
}

func SetLogLevel(newLogLevel string) {
	if Logger != nil {
		var lvl zapcore.Level
		if err := lvl.Set(newLogLevel); err != nil {
			lvl = zap.InfoLevel
		}
		logAtomLevel.SetLevel(lvl)
	}
}

// set v2 to v1, if v1 didn't bind any value
// FIXME: how about v1 bind default value?
func TrySetValue(v1, v2 interface{}) bool {
	var ok bool
	rt := reflect.TypeOf(v1)
	rv := reflect.ValueOf(v1)

	if rt.Kind() != reflect.Ptr {
		return ok
	}
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		rv = rv.Elem()
	}

	if rv.IsValid() && rv.IsZero() {
		ok = true
		switch rt.Kind() {
		case reflect.Uint:
			v, _ := v2.(uint)
			rv.SetUint(uint64(v))
		case reflect.Uint8:
			v, _ := v2.(uint8)
			rv.SetUint(uint64(v))
		case reflect.Uint16:
			v, _ := v2.(uint16)
			rv.SetUint(uint64(v))
		case reflect.Uint32:
			v, _ := v2.(uint32)
			rv.SetUint(uint64(v))
		case reflect.Uint64:
			v, _ := v2.(uint64)
			rv.SetUint(uint64(v))
		case reflect.Int:
			v, _ := v2.(int)
			rv.SetInt(int64(v))
		case reflect.Int8:
			v, _ := v2.(int8)
			rv.SetInt(int64(v))
		case reflect.Int16:
			v, _ := v2.(int16)
			rv.SetInt(int64(v))
		case reflect.Int32:
			v, _ := v2.(int32)
			rv.SetInt(int64(v))
		case reflect.Int64:
			v, _ := v2.(int64)
			rv.SetInt(int64(v))
		case reflect.Float32:
			v, _ := v2.(float32)
			rv.SetFloat(float64(v))
		case reflect.Float64:
			v, _ := v2.(float64)
			rv.SetFloat(float64(v))
		case reflect.String:
			rv.SetString(v2.(string))
		case reflect.Bool:
			rv.SetBool(v2.(bool))
		default:
			ok = false
		}
	}
	return ok
}

func CompareClickHouseVersion(v1, v2 string) int {
	s1 := strings.Split(v1, ".")
	s2 := strings.Split(v2, ".")
	for i := 0; i < len(s1); i++ {
		if len(s2) <= i {
			break
		}
		if s1[i] == "x" || s2[i] == "x" {
			continue
		}
		f1, _ := strconv.Atoi(s1[i])
		f2, _ := strconv.Atoi(s2[i])
		if f1 > f2 {
			return 1
		} else if f1 < f2 {
			return -1
		}
	}
	return 0
}

func Key(s string) string {
	return "${" + s + "}"
}

func Value(val interface{}) string {
	s := fmt.Sprint(val)
	return strings.ReplaceAll(s, ".", "_")
}

func Replace(str, key string, val interface{}) string {
	placeholder := Key(key)
	if !strings.Contains(str, Key(key)) {
		return str
	}

	valueStr := fmt.Sprint(val)
	if strings.Contains(valueStr, ".") {
		valueStr = strings.ReplaceAll(valueStr, ".", "_")
	}

	return strings.ReplaceAll(str, placeholder, valueStr)
}

func Match(str, key string) bool {
	return strings.Contains(str, Key(key))
}

func ZeroValue(v interface{}) bool {
	switch t := v.(type) {
	case string:
		return t == ""
	case int, int8, int16, int32, int64:
		return t == 0
	case uint, uint8, uint16, uint32, uint64:
		return t == 0
	case float32, float64:
		return t == 0
	case bool:
		return !t
	default:
		return false
	}
}

type VersionInfo struct {
	Version   string
	Commit    string
	BuildTime string
	GoVersion string
}

func GetProcessStartTime() int64 {
	return startTime.Unix()
}

// GetCPUUsage returns the current CPU usage percentage
// This is a simple implementation that works on most Unix-like systems
func GetCPUUsage() (float64, error) {
	// For a simple implementation, we'll use a basic approach
	// In a production environment, you might want to use a more sophisticated method

	// Read /proc/stat for system-wide CPU usage (Linux)
	if _, err := os.Stat("/proc/stat"); err == nil {
		return getCPUUsageLinux()
	}

	// For other systems, return 0 for now
	// You could implement Windows/macOS specific methods here
	return 0.0, nil
}

// getCPUUsageLinux reads CPU usage from /proc/stat on Linux systems
func getCPUUsageLinux() (float64, error) {
	// Read /proc/stat twice with a small interval to calculate usage
	stat1, err := readProcStat()
	if err != nil {
		return 0.0, err
	}

	// Wait a short time
	time.Sleep(100 * time.Millisecond)

	stat2, err := readProcStat()
	if err != nil {
		return 0.0, err
	}

	// Calculate CPU usage percentage
	totalDiff := stat2.total - stat1.total
	idleDiff := stat2.idle - stat1.idle

	if totalDiff == 0 {
		return 0.0, nil
	}

	usage := 100.0 * (1.0 - float64(idleDiff)/float64(totalDiff))
	return usage, nil
}

type cpuStat struct {
	total uint64
	idle  uint64
}

// readProcStat reads CPU statistics from /proc/stat
func readProcStat() (*cpuStat, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty /proc/stat")
	}

	// Parse the first line which contains overall CPU stats
	fields := strings.Fields(lines[0])
	if len(fields) < 5 || fields[0] != "cpu" {
		return nil, fmt.Errorf("invalid /proc/stat format")
	}

	// CPU fields: user, nice, system, idle, iowait, irq, softirq, steal, guest, guest_nice
	var values []uint64
	for i := 1; i < len(fields) && i <= 10; i++ {
		val, err := strconv.ParseUint(fields[i], 10, 64)
		if err != nil {
			return nil, err
		}
		values = append(values, val)
	}

	if len(values) < 4 {
		return nil, fmt.Errorf("insufficient CPU stat fields")
	}

	// Calculate total and idle
	var total uint64
	for _, val := range values {
		total += val
	}

	idle := values[3] // idle is the 4th field (index 3)

	return &cpuStat{
		total: total,
		idle:  idle,
	}, nil
}
