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

package task

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/housepower/clickhouse_sinker/config"
	"github.com/housepower/clickhouse_sinker/model"
	"github.com/housepower/clickhouse_sinker/output"
	"github.com/housepower/clickhouse_sinker/parser"
	"github.com/housepower/clickhouse_sinker/pool"
	"github.com/housepower/clickhouse_sinker/statistics"
	"github.com/housepower/clickhouse_sinker/util"
	"github.com/rcrowley/go-metrics"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

const (
	INVALID_COL_SEQ = -99
)

type ColKeys struct {
	dbkey      string
	knownKeys  sync.Map
	newKeys    sync.Map
	warnKeys   sync.Map
	cntNewKeys int32 // size of newKeys
}

// TaskService holds the configuration for each task
type Service struct {
	cfg               *config.Config
	clickhouse        *output.ClickHouse
	pp                *parser.Pool
	taskCfg           *config.TaskConfig
	whiteList         *regexp.Regexp
	blackList         *regexp.Regexp
	lblBlkList        *regexp.Regexp
	base              *ColKeys
	colKeys           map[string]*ColKeys
	dynamicSchemaLock sync.Mutex
	sharder           *Sharder
	limiter           *rate.Limiter //作用：控制打日志的频率
	offShift          int64
	consumer          *Consumer
	meter             metrics.Meter
}

// cloneTask create a new task by stealing members from s instead of creating a new one
func cloneTask(s *Service, newGroup *Consumer) (service *Service) {
	service = &Service{
		cfg:        s.cfg,
		clickhouse: s.clickhouse,
		pp:         s.pp,
		taskCfg:    s.taskCfg,
		consumer:   s.consumer,
		whiteList:  s.whiteList,
		blackList:  s.blackList,
		lblBlkList: s.lblBlkList,
		meter:      s.meter,
		colKeys:    make(map[string]*ColKeys),
		base:       s.base,
	}
	if newGroup != nil {
		service.consumer = newGroup
	}
	s.dynamicSchemaLock.Lock()
	for k, v := range s.colKeys {
		service.colKeys[k] = v
	}
	s.dynamicSchemaLock.Unlock()
	if err := service.Init(); err != nil {
		util.Logger.Fatal("failed to clone task", zap.String("group", service.taskCfg.ConsumerGroup), zap.String("task", service.taskCfg.Name), zap.Error(err))
	}

	return
}

// NewTaskService creates an instance of new tasks with kafka, clickhouse and paser instances
func NewTaskService(cfg *config.Config, taskCfg *config.TaskConfig, c *Consumer) (service *Service) {
	ck := output.NewClickHouse(cfg, taskCfg)
	pp, err := parser.NewParserPool(taskCfg.Parser, taskCfg.CsvFormat, taskCfg.Delimiter, taskCfg.TimeZone, taskCfg.TimeUnit, taskCfg.Fields)
	if err != nil {
		util.Logger.Fatal("failed to create task", zap.String("group", c.grpConfig.Name), zap.String("task", taskCfg.Name), zap.Error(err))
	}
	service = &Service{
		cfg:        cfg,
		clickhouse: ck,
		pp:         pp,
		taskCfg:    taskCfg,
		consumer:   c,
		base:       &ColKeys{},
		colKeys:    make(map[string]*ColKeys),
	}
	service.meter = metrics.NewMeter()
	metrics.GetOrRegister("rate.requests", service.meter)
	if taskCfg.DynamicSchema.WhiteList != "" {
		service.whiteList = regexp.MustCompile(taskCfg.DynamicSchema.WhiteList)
	}
	if taskCfg.DynamicSchema.BlackList != "" {
		service.blackList = regexp.MustCompile(taskCfg.DynamicSchema.BlackList)
	}
	if taskCfg.PromLabelsBlackList != "" {
		service.lblBlkList = regexp.MustCompile(taskCfg.PromLabelsBlackList)
	}
	return
}

func (service *Service) Meter() metrics.Meter {
	return service.meter
}

func (service *Service) copyColKeys(state *model.DbState) {
	colKey := &ColKeys{}
	if service.colKeys == nil {
		service.colKeys = make(map[string]*ColKeys)
	}
	for _, dims := range state.Dims {
		if _, ok := colKey.knownKeys.Load(dims.SourceName); !ok {
			colKey.knownKeys.Store(dims.SourceName, nil)
		}
	}
	for _, dims := range service.taskCfg.ExcludeColumns {
		if _, ok := colKey.knownKeys.Load(dims); !ok {
			colKey.knownKeys.Store(dims, nil)
		}
	}
	colKey.knownKeys.Store("", nil) // column name shall not be empty string
	colKey.newKeys = sync.Map{}
	atomic.StoreInt32(&colKey.cntNewKeys, 0)
	service.colKeys[state.DB] = colKey
}

// Init initializes the kafak and clickhouse task associated with this service
func (service *Service) Init() (err error) {
	taskCfg := service.taskCfg
	util.Logger.Info("task initializing", zap.String("task", taskCfg.Name))
	if err = service.clickhouse.Init(); err != nil {
		return
	}
	service.limiter = rate.NewLimiter(rate.Every(10*time.Second), 1)
	//service.offShift = int64(util.GetShift(taskCfg.BufferSize))
	service.offShift = int64(taskCfg.BufferSize)

	if len(service.clickhouse.SortingKeys) > 0 {
		service.taskCfg.ShardingKey = "__shardingkey"
		service.taskCfg.ShardingStripe = 1
	}

	if service.sharder, err = NewSharder(service); err != nil {
		return
	}
	service.clickhouse.Base.ShardingColSeq = service.sharder.policy.colSeq

	if taskCfg.DynamicSchema.Enable {
		for _, dim := range service.clickhouse.Base.Dims {
			service.base.knownKeys.Store(dim.SourceName, nil)
		}
		for _, dim := range taskCfg.ExcludeColumns {
			service.base.knownKeys.Store(dim, nil)
		}
		service.base.knownKeys.Store("", nil) // column name shall not be empty string
		service.base.newKeys = sync.Map{}
		atomic.StoreInt32(&service.base.cntNewKeys, 0)
	}
	service.consumer.addTask(service)

	return
}

func (service *Service) Put(msg *model.InputMessage, flushFn func()) error {
	taskCfg := service.taskCfg
	statistics.ConsumeMsgsTotal.WithLabelValues(taskCfg.Name).Inc()
	var err error
	var row *model.Row
	var state *model.DbState
	var foundNewKeys bool
	var metric model.Metric
	var colKey *ColKeys

	p, err := service.pp.Get()
	if err != nil {
		util.Logger.Fatal("error initializing json parser", zap.String("task", taskCfg.Name), zap.Error(err))
	}
	if metric, err = p.Parse(msg.Value); err != nil {
		// directly return, ignore the row with parsing errors
		statistics.ParseMsgsErrorTotal.WithLabelValues(taskCfg.Name).Inc()
		if service.limiter.Allow() {
			util.Logger.Error(fmt.Sprintf("failed to parse message(topic %v, partition %d, offset %v)",
				msg.Topic, msg.Partition, msg.Offset), zap.String("message value", string(msg.Value)), zap.String("task", taskCfg.Name), zap.Error(err))
		}
		return nil
	} else {
		state, row = service.metric2Row(metric, msg)
		if row == nil {
			return nil
		}
		if state == nil {
			return nil
		}
		if state.NewKey {
			service.dynamicSchemaLock.Lock()
			err = service.clickhouse.EnsureSchema(state)

			if err != nil {
				util.Logger.Error("failed to ensure schema", zap.String("task", taskCfg.Name), zap.Error(err))
			}
			service.dynamicSchemaLock.Unlock()
			state.NewKey = false
			policy, err := NewShardingPolicy(taskCfg.ShardingKey, taskCfg.ShardingStripe, state.Dims, pool.NumShard())
			if err == nil {
				state.ShardingColSeq = policy.colSeq
			} else {
				state.ShardingColSeq = INVALID_COL_SEQ
			}

			service.consumer.SetDbMap(state.DB, state)
			service.dynamicSchemaLock.Lock()
			service.copyColKeys(state)
			service.dynamicSchemaLock.Unlock()
		}
		if taskCfg.DynamicSchema.Enable {
			service.dynamicSchemaLock.Lock()
			colKey = service.colKeys[state.DB]
			if colKey == nil {
				service.copyColKeys(state)
				colKey = service.colKeys[state.DB]
			}
			foundNewKeys = metric.GetNewKeys(&colKey.knownKeys, &colKey.newKeys, &colKey.warnKeys, service.whiteList, service.blackList, msg.Partition, msg.Offset)
			service.dynamicSchemaLock.Unlock()
		} else {
			service.dynamicSchemaLock.Lock()
			colKey = service.colKeys[state.DB]
			if colKey == nil {
				service.copyColKeys(state)
				colKey = service.colKeys[state.DB]
			}
			service.dynamicSchemaLock.Unlock()
		}
	}
	// WARNNING: metric.GetXXX may depend on p. Don't call them after p been freed.
	service.pp.Put(p)

	if foundNewKeys {
		cntNewKeys := atomic.AddInt32(&colKey.cntNewKeys, 1)
		if cntNewKeys == 1 {
			// the first message which contains new keys triggers the following:
			// 1) restart the consumer group
			// 	 1) stop the consumer to prevent blocking other consumers, stop will process until ChangeSchema completed
			// 2) flush the shards
			// 3) apply the schema change.
			// 4) recreate the service
			if len(service.consumer.grpConfig.Configs) > 1 {
				util.Logger.Warn("new key detected, consumer is going to restart", zap.String("consumer group", service.taskCfg.ConsumerGroup), zap.Error(err))
				go service.consumer.restart()
			}
			flushFn()
			if err = service.clickhouse.ChangeSchema(state, &colKey.newKeys); err != nil {
				util.Logger.Fatal("clickhouse.ChangeSchema failed", zap.String("task", taskCfg.Name), zap.Error(err))
			}
			service.consumer.DelDbMap(state.DB)
			cloneTask(service, nil)
			util.Rs.Reset()
			return fmt.Errorf("consumer restart required due to new key")
		}
	}

	if colKey != nil && atomic.LoadInt32(&colKey.cntNewKeys) == 0 && service.consumer.state.Load() == util.StateRunning {
		msgRow := model.MsgRow{Msg: msg, Row: row}
		if service.sharder.policy != nil && state.ShardingColSeq < len(*row) {
			if msgRow.Shard, err = service.sharder.Calc(msgRow.Row, msg.Offset, state.ShardingColSeq); err != nil {
				util.Logger.Warn("shard number calculation failed, skip this message", zap.String("task", taskCfg.Name),
					zap.String("dbkey", state.DB), zap.String("topic", msg.Topic),
					zap.Int("partition", msg.Partition), zap.Int64("offset", msg.Offset),
					zap.Int("colseq", state.ShardingColSeq),
					zap.Reflect("row", msgRow.Row),
					zap.Error(err))
				return err
			}
		} else {
			msgRow.Shard = int(msgRow.Msg.Offset * (int64(msgRow.Msg.Partition + 1)) >> service.offShift % int64(service.sharder.shards))
		}
		service.sharder.PutElement(state.DB, &msgRow)
	}

	return nil
}

func (service *Service) GetDbKey(metric model.Metric) string {
	key := service.clickhouse.Base.DB //基础库的库名
	dim := service.clickhouse.KeyDim
	if dim.IsDbKey {
		val := model.GetValueByType(metric, &dim)
		if val != nil && !util.ZeroValue(val) {
			key = util.Replace(service.consumer.sinker.curCfg.Clickhouse.DbKey, dim.SourceName, val)
		} // 如果dbkey 没有设置，那么返回的还是基础库的库名
	}
	return key
}

func (service *Service) metric2Row(metric model.Metric, msg *model.InputMessage) (*model.DbState, *model.Row) {
	base := service.clickhouse.Base
	dims := base.Dims
	numDims := base.NumDims
	idxSerID := base.IdxSerID

	key := service.GetDbKey(metric)
	state, ok := service.consumer.GetDbMap(key)
	if ok {
		dims = state.Dims
		numDims = state.NumDims
		idxSerID = state.IdxSerID
	} else {
		// 此处需要原样复制，避免修改base的值, copy dims， 避免因为浅拷贝导致base被意外修改
		newDims := make([]*model.ColumnWithType, len(base.Dims))
		copy(newDims, base.Dims)
		state = &model.DbState{
			DB:             key, //使用获取到的key
			PrepareSQL:     base.PrepareSQL,
			PromSerSQL:     base.PromSerSQL,
			BufLength:      0,
			Processed:      0,
			NewKey:         true,
			Dims:           newDims,
			NumDims:        base.NumDims,
			IdxSerID:       base.IdxSerID,
			ShardingColSeq: base.ShardingColSeq,
		}
	}

	if idxSerID >= 0 {
		// If some labels are not Prometheus native, ETL shall calculate and pass "__series_id__" and "__mgmt_id__".
		val := metric.GetInt64(service.clickhouse.DimSerID, false)
		seriesID := val.(int64)
		val = metric.GetInt64(service.clickhouse.DimMgmtID, false)
		mgmtID := val.(int64)
		newSeries := service.clickhouse.AllowWriteSeries(seriesID, mgmtID)
		rowcount := idxSerID + 1 // including __series_id__
		if newSeries {
			// 啥意思？
			rowcount += (numDims - idxSerID + 3)
		}

		row := make(model.Row, 0, rowcount)
		for i := 0; i < idxSerID; i++ {
			row = append(row, model.GetValueByType(metric, dims[i]))
		}
		if idxSerID == 0 {
			util.Logger.Info("CATCH YOU!!!!!!!!!!!!!!!", zap.Reflect("state", state))
		}
		row = append(row, seriesID) // __series_id__
		if newSeries {
			var labels []string
			row = append(row, mgmtID, nil) // __mgmt_id__, labels
			for i := idxSerID + 3; i < numDims; i++ {
				dim := dims[i]
				val := model.GetValueByType(metric, dim)
				row = append(row, val)
				if val != nil && dim.Type.Type == model.String && dim.Name != service.clickhouse.NameKey && dim.Name != "le" && (service.lblBlkList == nil || !service.lblBlkList.MatchString(dim.Name)) {
					// "labels" JSON excludes "le", so that "labels" can be used as group key for histogram queries.
					if !(service.taskCfg.DynamicSchema.NotNullable && val == "") {
						labelVal := val.(string)
						labels = append(labels, fmt.Sprintf(`%s: %s`, strconv.Quote(dim.Name), strconv.Quote(labelVal)))
					}
				}
			}

			row[idxSerID+2] = fmt.Sprintf("{%s}", strings.Join(labels, ", "))
		}
		//util.Logger.Info("metric2Row2222", zap.String("key", state.Name))
		atomic.AddInt64(&state.BufLength, 1)
		atomic.AddInt64(&state.Processed, 1)
		service.consumer.SetDbMap(key, state)
		return state, &row
	} else {
		var shardingVal uint64
		if len(service.clickhouse.SortingKeys) > 0 {
			var sortingKeys []string
			for _, dim := range service.clickhouse.SortingKeys {
				sortingKeys = append(sortingKeys, fmt.Sprintf("%v", model.GetValueByType(metric, dim)))
			}

			shardingVal = xxhash.Sum64String(strings.Join(sortingKeys, "."))
		}
		row := make(model.Row, 0, len(dims))
		for _, dim := range dims {
			if strings.HasPrefix(dim.Name, "__kafka") {
				if strings.HasSuffix(dim.Name, "_topic") {
					row = append(row, msg.Topic)
				} else if strings.HasSuffix(dim.Name, "_partition") {
					row = append(row, msg.Partition)
				} else if strings.HasSuffix(dim.Name, "_offset") {
					row = append(row, msg.Offset)
				} else if strings.HasSuffix(dim.Name, "_key") {
					row = append(row, string(msg.Key))
				} else if strings.HasSuffix(dim.Name, "_timestamp") {
					row = append(row, *msg.Timestamp)
				} else {
					row = append(row, nil)
				}
			} else if dim.Name == "__shardingkey" {
				row = append(row, shardingVal)
			} else {
				val := model.GetValueByType(metric, dim)
				if dim.NotNullable && val == nil {
					// null 不能插入到非 nullbale字段中
					util.Logger.Warn("null value detected, throw this message",
						zap.String("dimension", dim.Name),
						zap.String("task", service.taskCfg.Name),
						zap.String("topic", msg.Topic),
						zap.Int("partition", msg.Partition),
						zap.Int64("offset", msg.Offset),
						zap.String("key", string(msg.Key)),
						zap.Time("timestamp", *msg.Timestamp))
					return state, nil
				}
				row = append(row, val)
			}
		}
		atomic.AddInt64(&state.BufLength, 1)
		atomic.AddInt64(&state.Processed, 1)
		service.consumer.SetDbMap(key, state)

		return state, &row
	}
}
