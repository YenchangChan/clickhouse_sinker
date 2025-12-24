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

package output

import (
	"encoding/json"
	"expvar"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/avast/retry-go/v4"
	"github.com/housepower/clickhouse_sinker/config"
	"github.com/housepower/clickhouse_sinker/model"
	"github.com/housepower/clickhouse_sinker/pool"
	"github.com/housepower/clickhouse_sinker/statistics"
	"github.com/housepower/clickhouse_sinker/util"
	"github.com/thanos-io/thanos/pkg/errors"
	"go.uber.org/zap"
)

var (
	ErrTblNotExist     = errors.Newf("table doesn't exist")
	selectSQLTemplate  = `select name, type, default_kind from system.columns where database = '%s' and table = '%s'`
	referedSQLTemplate = `SELECT 
    current_col.default_expression,
    referenced_col.type AS referenced_col_type,
    current_col.name,
    current_col.type
FROM 
    system.columns AS current_col
JOIN 
    system.columns AS referenced_col 
ON 
    current_col.database = referenced_col.database 
    AND current_col.table = referenced_col.table 
    AND current_col.default_expression = referenced_col.name 
WHERE 
    current_col.database = '%s' 
    AND
    current_col.table = '%s';`
	wrSeriesQuota int = 16384

	SeriesQuotas sync.Map

	distEngineReg = regexp.MustCompile(`(Distributed\s*\(\s*'[^']*',\s*')[^']*(')`)
)

// ClickHouse is an output service consumers from kafka messages
type ClickHouse struct {
	Base           *model.DbState
	KeyDim         model.ColumnWithType
	NameKey        string
	DimSerID       string
	DimMgmtID      string
	SortingKeys    []*model.ColumnWithType
	TableName      string
	cfg            *config.Config
	taskCfg        *config.TaskConfig
	seriesTbl      string
	distMetricTbls []string
	distSeriesTbls []string
	seriesQuota    *model.SeriesQuota
	numFlying      int32
	mux            sync.Mutex
	taskDone       *sync.Cond
}

type DistTblInfo struct {
	name    string
	cluster string
}

func init() {
	expvar.Publish("SeriesMap", expvar.Func(func() interface{} {
		var result = make(map[string]string)
		SeriesQuotas.Range(func(key, value interface{}) bool {
			if sq, ok := value.(*model.SeriesQuota); ok {
				sq.RLock()
				if bs, err := json.Marshal(sq); err == nil {
					result[key.(string)] = string(bs)
				}
				sq.RUnlock()
			}
			return true
		})
		return result
	}))
}

// NewClickHouse new a clickhouse instance
func NewClickHouse(cfg *config.Config, taskCfg *config.TaskConfig) *ClickHouse {
	ck := &ClickHouse{cfg: cfg, taskCfg: taskCfg}
	ck.taskDone = sync.NewCond(&ck.mux)
	return ck
}

// Init the clickhouse intance
func (c *ClickHouse) Init() (err error) {
	c.Base, err = c.initSchema(c.cfg.Clickhouse.DB)
	return err
}

// Drain drains flying batchs
func (c *ClickHouse) Drain() {
	c.mux.Lock()
	for c.numFlying != 0 {
		util.Logger.Debug("draining flying batches",
			zap.String("task", c.taskCfg.Name),
			zap.Int32("pending", c.numFlying))
		c.taskDone.Wait()
	}
	c.mux.Unlock()
}

// Send a batch to clickhouse
func (c *ClickHouse) Send(batch *model.Batch, state model.DbState) {
	sc := pool.GetShardConn(batch.BatchIdx)
	if err := sc.SubmitTask(func() {
		c.loopWrite(batch, sc, state)
		batch.Wg.Done()
		c.mux.Lock()
		c.numFlying--
		if c.numFlying == 0 {
			c.taskDone.Broadcast()
		}
		c.mux.Unlock()
		statistics.WritingPoolBacklog.WithLabelValues(c.taskCfg.Name).Dec()
	}); err != nil {
		batch.Wg.Done()
		util.Rs.Dec(int64(batch.RealSize))
		statistics.RecordPoolSize.WithLabelValues().Sub(float64(batch.RealSize))
		return
	}

	c.mux.Lock()
	c.numFlying++
	c.mux.Unlock()
	statistics.WritingPoolBacklog.WithLabelValues(c.taskCfg.Name).Inc()
}

func (c *ClickHouse) AllowWriteSeries(sid, mid int64) (allowed bool) {
	c.seriesQuota.Lock()
	defer c.seriesQuota.Unlock()
	mid2, loaded := c.seriesQuota.BmSeries[sid]
	if !loaded {
		util.Logger.Debug("found new series", zap.Int64("mid", mid), zap.Int64("sid", sid))
		allowed = true
		statistics.WriteSeriesAllowNew.WithLabelValues(c.taskCfg.Name).Inc()
	} else if mid != mid2 {
		util.Logger.Debug("found new series map", zap.Int64("mid", mid), zap.Int64("sid", sid))
		if c.seriesQuota.WrSeries < wrSeriesQuota {
			c.seriesQuota.WrSeries++
			allowed = true
		} else {
			now := time.Now()
			if now.After(c.seriesQuota.NextResetQuota) {
				c.seriesQuota.NextResetQuota = now.Add(10 * time.Second)
				c.seriesQuota.WrSeries = 1
				allowed = true
			}
		}
		if allowed {
			statistics.WriteSeriesAllowChanged.WithLabelValues(c.taskCfg.Name).Inc()
		} else {
			statistics.WriteSeriesDropQuota.WithLabelValues(c.taskCfg.Name).Inc()
		}
	} else {
		statistics.WriteSeriesDropUnchanged.WithLabelValues(c.taskCfg.Name).Inc()
	}
	return
}

func (c *ClickHouse) writeSeries(promSerSQL string, idxSerID, numDims int, rows model.Rows, conn *pool.Conn) (err error) {
	var seriesRows model.Rows
	for _, row := range rows {
		if len(*row) != numDims {
			continue
		}
		seriesRows = append(seriesRows, row)
	}
	if len(seriesRows) != 0 {
		begin := time.Now()
		var numBad int
		if numBad, err = writeRows(promSerSQL, seriesRows, idxSerID, numDims, conn); err != nil {
			return
		}
		// update c.bmSeries **after** writing series
		c.seriesQuota.Lock()
		for _, row := range seriesRows {
			sid := (*row)[idxSerID].(int64)
			mid := (*row)[idxSerID+1].(int64)
			if _, loaded := c.seriesQuota.BmSeries[sid]; loaded {
				c.seriesQuota.WrSeries--
			}
			if c.seriesQuota.BmSeries == nil {
				c.seriesQuota.BmSeries = make(map[int64]int64)
			}
			c.seriesQuota.BmSeries[sid] = mid
		}
		c.seriesQuota.Unlock()
		util.Logger.Info("ClickHouse.writeSeries succeeded", zap.Int("series", len(seriesRows)), zap.String("task", c.taskCfg.Name))
		statistics.WriteSeriesSucceed.WithLabelValues(c.taskCfg.Name).Add(float64(len(seriesRows)))
		if numBad != 0 {
			statistics.ParseMsgsErrorTotal.WithLabelValues(c.taskCfg.Name).Add(float64(numBad))
		}
		statistics.WritingDurations.WithLabelValues(c.taskCfg.Name, c.seriesTbl).Observe(time.Since(begin).Seconds())
	}
	return
}

// Write a batch to clickhouse
func (c *ClickHouse) write(batch *model.Batch, sc *pool.ShardConn, dbVer *int, state model.DbState) (err error) {
	if len(*batch.Rows) == 0 {
		return
	}
	var conn *pool.Conn
	if conn, *dbVer, err = sc.NextGoodReplica(c.cfg.Clickhouse.Ctx, *dbVer); err != nil {
		return
	}
	util.Logger.Debug("writing batch", zap.String("task", c.taskCfg.Name), zap.String("replica", sc.GetReplica()), zap.Int("dbVer", *dbVer))

	//row[:c.IdxSerID+1] is for metric table
	//row[c.IdxSerID:] is for series table
	numDims := state.NumDims
	if c.taskCfg.PrometheusSchema {
		numDims = state.IdxSerID + 1
		if err = c.writeSeries(state.PromSerSQL, state.IdxSerID, state.NumDims, *batch.Rows, conn); err != nil {
			return
		}
	}
	begin := time.Now()
	var numBad int
	util.Logger.Info("write to clickhouse", zap.Int("rows", len(*batch.Rows)),
		zap.Int("numDims", numDims), zap.String("task", c.taskCfg.Name),
		zap.String("dbkey", state.DB), zap.String("replica", sc.GetReplica()))
	if numBad, err = writeRows(state.PrepareSQL, *batch.Rows, 0, numDims, conn); err != nil {
		return
	}
	statistics.WritingDurations.WithLabelValues(c.taskCfg.Name, c.TableName).Observe(time.Since(begin).Seconds())
	if numBad != 0 {
		statistics.ParseMsgsErrorTotal.WithLabelValues(c.taskCfg.Name).Add(float64(numBad))
	}
	statistics.FlushMsgsTotal.WithLabelValues(c.taskCfg.Name, state.DB).Add(float64(batch.RealSize))
	return
}

// LoopWrite will dead loop to write the records
func (c *ClickHouse) loopWrite(batch *model.Batch, sc *pool.ShardConn, state model.DbState) {
	var retrycount int
	var dbVer int

	defer func() {
		util.Rs.Dec(int64(batch.RealSize))
		statistics.RecordPoolSize.WithLabelValues().Sub(float64(batch.RealSize))
	}()
	times := c.cfg.Clickhouse.RetryTimes
	if times <= 0 {
		times = 0
	}
	if err := retry.Do(
		func() error { return c.write(batch, sc, &dbVer, state) },
		retry.LastErrorOnly(true),
		retry.Attempts(uint(times)),
		retry.Delay(10*time.Second),
		retry.MaxDelay(1*time.Minute),
		retry.OnRetry(func(n uint, err error) {
			retrycount++
			util.Logger.Error("flush batch failed",
				zap.String("task", c.taskCfg.Name),
				zap.String("group", batch.GroupId),
				zap.Int("try", int(retrycount)),
				zap.Error(err))
			statistics.FlushMsgsErrorTotal.WithLabelValues(c.taskCfg.Name).Add(float64(batch.RealSize))
		}),
	); err != nil {
		util.Logger.Fatal("ClickHouse.loopWrite failed", zap.String("task", c.taskCfg.Name), zap.Error(err))
	}
}

func (c *ClickHouse) getSeriesDims(dims []*model.ColumnWithType, conn *pool.Conn) {
	for _, dim := range dims {
		if strings.Contains(dim.Name, "series_id") {
			c.DimSerID = dim.Name
		}
		if strings.Contains(dim.Name, "mgmt_id") {
			c.DimMgmtID = dim.Name
		}
	}
}

func (c *ClickHouse) initSeriesSchema(conn *pool.Conn, database string, state *model.DbState) (err error) {
	if !c.taskCfg.PrometheusSchema {
		state.IdxSerID = -1
		return
	}

	// Add string columns from series table
	if c.seriesTbl == "" {
		c.seriesTbl = c.TableName + "_series"
	}
	var keyDim model.ColumnWithType
	var seriesDims []*model.ColumnWithType
	if seriesDims, keyDim, err = getDims(database, c.seriesTbl, nil, c.cfg.Clickhouse.DbKey, c.taskCfg.Parser, conn); err != nil {
		if errors.Is(err, ErrTblNotExist) {
			err = errors.Wrapf(err, "Please create series table for %s.%s", database, c.TableName)
			return
		}
		return
	}
	if c.cfg.Clickhouse.DbKey != "" && keyDim.IsDbKey {
		c.KeyDim = keyDim
	}

	c.getSeriesDims(seriesDims, conn)

	// Move column "__series_id__" to the last.
	var dimSerID *model.ColumnWithType
	for i := 0; i < len(state.Dims); {
		dim := state.Dims[i]
		if dim.Name == c.DimSerID && dim.Type.Type == model.Int64 {
			dimSerID = dim
			state.Dims = append(state.Dims[:i], state.Dims[i+1:]...)
			break
		} else {
			i++
		}
	}
	if dimSerID == nil {
		err = errors.Newf("Metric table %s.%s shall have column `%s Int64`.", database, c.TableName, c.DimSerID)
		return
	}
	state.IdxSerID = len(state.Dims)
	state.Dims = append(state.Dims, dimSerID)

	expSeriesDims := []*model.ColumnWithType{
		{Name: c.DimSerID, Type: &model.TypeInfo{Type: model.Int64}},
		{Name: c.DimMgmtID, Type: &model.TypeInfo{Type: model.Int64}},
		{Name: "labels", Type: &model.TypeInfo{Type: model.String}},
	}

	var badFirst bool
	if len(seriesDims) < len(expSeriesDims) {
		badFirst = true
	} else {
		for i := range expSeriesDims {
			if seriesDims[i].Name != expSeriesDims[i].Name ||
				seriesDims[i].Type.Type != expSeriesDims[i].Type.Type {
				badFirst = true
				break
			}
		}
	}
	if badFirst {
		err = errors.Newf(`First columns of %s are expect to be %s Int64, %s Int64, labels String".`, c.seriesTbl, c.DimSerID, c.DimMgmtID)
		return
	}
	c.NameKey = "__name__" // prometheus uses internal "__name__" label for metric name
	for i := len(expSeriesDims); i < len(seriesDims); i++ {
		serDim := seriesDims[i]
		if serDim.Type.Type == model.String {
			c.NameKey = serDim.Name // opentsdb uses "metric" tag for metric name
			break
		}
	}
	state.Dims = append(state.Dims, seriesDims[1:]...)

	// Generate SQL for series INSERT
	if c.cfg.Clickhouse.Protocol == clickhouse.HTTP.String() {
		serDimsQuoted := make([]string, len(seriesDims))
		for i, serDim := range seriesDims {
			serDimsQuoted[i] = fmt.Sprintf("`%s`", serDim.Name)
		}
		var params = make([]string, len(seriesDims))
		for i := range params {
			params[i] = "?"
		}

		state.PromSerSQL = "INSERT INTO " + database + "." + c.seriesTbl + " (" + strings.Join(serDimsQuoted, ",") + ") " +
			"VALUES (" + strings.Join(params, ",") + ")"
	} else {
		serDimsQuoted := make([]string, len(seriesDims))
		for i, serDim := range seriesDims {
			serDimsQuoted[i] = fmt.Sprintf("`%s`", serDim.Name)
		}
		state.PromSerSQL = fmt.Sprintf("INSERT INTO `%s`.`%s` (%s)",
			database,
			c.seriesTbl,
			strings.Join(serDimsQuoted, ","))
	}
	util.Logger.Info(fmt.Sprintf("promSer sql=> %s", state.PromSerSQL), zap.String("task", c.taskCfg.Name))

	// Check distributed series table
	if chCfg := &c.cfg.Clickhouse; chCfg.Cluster != "" {
		withDistTable := false
		info, e := c.getDistTbls(database, c.seriesTbl, chCfg.Cluster)
		if e != nil {
			return e
		}
		c.distSeriesTbls = make([]string, 0)
		for _, i := range info {
			c.distSeriesTbls = append(c.distSeriesTbls, i.name)
			if i.cluster == c.cfg.Clickhouse.Cluster {
				withDistTable = true
			}
		}
		if !withDistTable {
			err = errors.Newf("Please create distributed table for %s in cluster '%s'.", c.seriesTbl, c.cfg.Clickhouse.Cluster)
			return
		}
	}

	// seriesQuota 使用全局的，前提是__series_id__多租户不会重复
	// 仅第一次初始化的时候需要初始化SeriesQuota
	seriesQuotaKey := state.DB
	sq, _ := SeriesQuotas.LoadOrStore(c.GetSeriesQuotaKey(seriesQuotaKey),
		&model.SeriesQuota{
			NextResetQuota: time.Now().Add(10 * time.Second),
			Birth:          time.Now(),
		})
	c.seriesQuota = sq.(*model.SeriesQuota)
	return
}

func (c *ClickHouse) initSchema(database string) (state *model.DbState, err error) {
	state = &model.DbState{}
	if idx := strings.Index(c.taskCfg.TableName, "."); idx > 0 {
		c.TableName = c.taskCfg.TableName[idx+1:]
		if c.cfg.Clickhouse.DbKey != "" {
			state.DB = database
		} else {
			state.DB = c.taskCfg.TableName[0:idx]
		}
	} else {
		c.TableName = c.taskCfg.TableName[idx+1:]
		state.DB = database
	}
	c.seriesTbl = c.taskCfg.SeriesTableName

	sc := pool.GetShardConn(0)
	var conn *pool.Conn
	if conn, _, err = sc.NextGoodReplica(c.cfg.Clickhouse.Ctx, 0); err != nil {
		return
	}
	// Check distributed metric table
	if chCfg := &c.cfg.Clickhouse; chCfg.Cluster != "" {
		withDistTable := false
		info, e := c.getDistTbls(database, c.TableName, chCfg.Cluster)
		if e != nil {
			return state, e
		}
		c.distMetricTbls = make([]string, 0)
		for _, i := range info {
			c.distMetricTbls = append(c.distMetricTbls, i.name)
			if i.cluster == c.cfg.Clickhouse.Cluster {
				withDistTable = true
			}
		}
		if !withDistTable {
			err = errors.Newf("Please create distributed table for %s in cluster '%s'.", c.TableName, c.cfg.Clickhouse.Cluster)
			return
		}
	}
	if err = c.ensureShardingkey(conn, database, c.TableName, c.taskCfg.Parser); err != nil {
		return
	}
	if c.taskCfg.AutoSchema {
		if state.Dims, c.KeyDim, err = getDims(database, c.TableName, c.taskCfg.ExcludeColumns, c.cfg.Clickhouse.DbKey, c.taskCfg.Parser, conn); err != nil {
			return
		}
	} else {
		state.Dims = make([]*model.ColumnWithType, 0, len(c.taskCfg.Dims))
		for _, dim := range c.taskCfg.Dims {
			state.Dims = append(state.Dims, &model.ColumnWithType{
				Name:       dim.Name,
				Type:       model.WhichType(dim.Type),
				SourceName: dim.SourceName,
			})
		}
	}
	if err = c.initSeriesSchema(conn, database, state); err != nil {
		return
	}
	state.NumDims = len(state.Dims)
	// Generate SQL for INSERT
	if c.cfg.Clickhouse.Protocol == clickhouse.HTTP.String() {
		numDims := state.NumDims
		if c.taskCfg.PrometheusSchema {
			numDims = state.IdxSerID + 1
		}
		quotedDms := make([]string, numDims)
		for i := 0; i < numDims; i++ {
			quotedDms[i] = fmt.Sprintf("`%s`", state.Dims[i].Name)
		}
		var params = make([]string, numDims)
		for i := range params {
			params[i] = "?"
		}
		state.PrepareSQL = "INSERT INTO " + database + "." + c.TableName + " (" + strings.Join(quotedDms, ",") + ") " +
			"VALUES (" + strings.Join(params, ",") + ")"
	} else {
		numDims := state.NumDims
		if c.taskCfg.PrometheusSchema {
			numDims = state.IdxSerID + 1
		}
		quotedDims := make([]string, numDims)
		for i := 0; i < numDims; i++ {
			quotedDims[i] = fmt.Sprintf("`%s`", state.Dims[i].Name)
		}
		state.PrepareSQL = fmt.Sprintf("INSERT INTO `%s`.`%s` (%s)",
			database,
			c.TableName,
			strings.Join(quotedDims, ","))
	}
	util.Logger.Info(fmt.Sprintf("Prepare sql=> %s", state.PrepareSQL), zap.String("task", c.taskCfg.Name))
	return
}

func (c *ClickHouse) ChangeSchema(state *model.DbState, newKeys *sync.Map) (err error) {
	var onCluster string
	taskCfg := c.taskCfg
	chCfg := &c.cfg.Clickhouse
	if chCfg.Cluster != "" {
		onCluster = fmt.Sprintf("ON CLUSTER `%s`", chCfg.Cluster)
	}
	maxDims := math.MaxInt16
	if taskCfg.DynamicSchema.MaxDims > 0 {
		maxDims = taskCfg.DynamicSchema.MaxDims
	}
	newKeysQuota := maxDims - len(state.Dims)
	if newKeysQuota <= 0 {
		util.Logger.Warn("number of columns reaches upper limit", zap.Int("limit", maxDims), zap.Int("current", len(state.Dims)))
		return
	}

	var i int
	var alterSeries, alterMetric []string
	newKeys.Range(func(key, value interface{}) bool {
		i++
		if i > newKeysQuota {
			util.Logger.Warn("number of columns reaches upper limit", zap.Int("limit", maxDims), zap.Int("current", i))
			return false
		}
		strKey, _ := key.(string)
		intVal := value.(int)
		var strVal string
		switch intVal {
		case model.Bool:
			strVal = "Bool"
		case model.Int64:
			strVal = "Int64"
		case model.Float64:
			strVal = "Float64"
		case model.String:
			strVal = "String"
		case model.DateTime:
			strVal = "DateTime64(3)"
		case model.Object:
			strVal = model.GetTypeName(intVal)
		default:
			err = errors.Newf("%s: BUG: unsupported column type %s", taskCfg.Name, model.GetTypeName(intVal))
			return false
		}

		if !taskCfg.DynamicSchema.NotNullable {
			strVal = fmt.Sprintf("Nullable(%v)", strVal)
		}

		if c.taskCfg.PrometheusSchema && intVal == model.String {
			alterSeries = append(alterSeries, fmt.Sprintf("ADD COLUMN IF NOT EXISTS `%s` %s", strKey, strVal))
		} else {
			if c.taskCfg.PrometheusSchema {
				if intVal > model.String {
					util.Logger.Fatal("unsupported metric value type", zap.String("type", strVal), zap.String("name", strKey), zap.String("task", c.taskCfg.Name))
				} else if intVal == model.Float64 || (intVal == model.Int64 && strKey != c.DimMgmtID) {
					// 多指标仅支持float64和int64
					alterMetric = append(alterMetric, fmt.Sprintf("ADD COLUMN IF NOT EXISTS `%s` %s", strKey, strVal))
				}
			} else {
				alterMetric = append(alterMetric, fmt.Sprintf("ADD COLUMN IF NOT EXISTS `%s` %s", strKey, strVal))
			}
		}
		return true
	})
	if err != nil {
		return
	}

	sc := pool.GetShardConn(0)
	var conn *pool.Conn
	if conn, _, err = sc.NextGoodReplica(c.cfg.Clickhouse.Ctx, 0); err != nil {
		return
	}

	var version string
	if err = conn.QueryRow("SELECT version()").Scan(&version); err != nil {
		version = "1.0.0.0"
	}
	alterTable := func(tbl, col string) error {
		query := fmt.Sprintf("ALTER TABLE `%s`.`%s` %s %s", state.DB, tbl, onCluster, col)
		if util.CompareClickHouseVersion(version, "23.3") >= 0 {
			query += " SETTINGS alter_sync = 0"
		}
		util.Logger.Info(fmt.Sprintf("executing sql=> %s", query), zap.String("task", taskCfg.Name))
		return conn.Exec(query)
	}

	if len(alterSeries) != 0 {
		sort.Strings(alterSeries)
		columns := strings.Join(alterSeries, ",")
		if err = alterTable(c.seriesTbl, columns); err != nil {
			return err
		}
		for _, distTbl := range c.distSeriesTbls {
			if err = alterTable(distTbl, columns); err != nil {
				return err
			}
		}
	}
	if len(alterMetric) != 0 {
		sort.Strings(alterMetric)
		columns := strings.Join(alterMetric, ",")
		if err = alterTable(c.TableName, columns); err != nil {
			return err
		}
		for _, distTbl := range c.distMetricTbls {
			if err = alterTable(distTbl, columns); err != nil {
				return err
			}
		}
	}

	return
}

func (c *ClickHouse) getDistTbls(database, table, clusterName string) (distTbls []DistTblInfo, err error) {
	taskCfg := c.taskCfg
	sc := pool.GetShardConn(0)
	var conn *pool.Conn
	if conn, _, err = sc.NextGoodReplica(c.cfg.Clickhouse.Ctx, 0); err != nil {
		return
	}
	query := fmt.Sprintf(`SELECT name, (extractAllGroups(engine_full, '(Distributed\\(\')(.*)\',\\s+\'(.*)\',\\s+\'(.*)\'(.*)')[1])[2] AS cluster
	 FROM system.tables WHERE engine='Distributed' AND database='%s' AND match(engine_full, 'Distributed\(\'.*\', \'%s\', \'%s\'.*\)')`,
		database, database, table)
	util.Logger.Info(fmt.Sprintf("executing sql=> %s", query), zap.String("task", taskCfg.Name))
	var rows *pool.Rows
	if rows, err = conn.Query(query); err != nil {
		err = errors.Wrapf(err, "")
		return
	}
	defer rows.Close()
	var curInfo DistTblInfo
	for rows.Next() {
		var name, cluster string
		if err = rows.Scan(&name, &cluster); err != nil {
			err = errors.Wrapf(err, "")
			return
		}
		if cluster == clusterName {
			// distributed table
			curInfo = DistTblInfo{name: name, cluster: cluster}
		} else {
			// logic table
			distTbls = append(distTbls, DistTblInfo{name: name, cluster: cluster})
		}
	}
	// dist table always in the end
	distTbls = append(distTbls, curInfo)
	return
}

func (c *ClickHouse) GetSeriesQuotaKey(db string) string {
	if db == "" {
		db = c.cfg.Clickhouse.DB
	}
	if c.taskCfg.PrometheusSchema {
		if c.cfg.Clickhouse.Cluster != "" {
			return db + "." + c.distSeriesTbls[len(c.distSeriesTbls)-1]
		} else {
			return db + "." + c.seriesTbl
		}
	}
	return ""
}

func (c *ClickHouse) GetMetricTable() string {
	if c.taskCfg.PrometheusSchema {
		if c.cfg.Clickhouse.Cluster != "" {
			return c.distMetricTbls[len(c.distMetricTbls)-1]
		} else {
			return c.TableName
		}
	}
	return ""
}

func (c *ClickHouse) ensureShardingkey(conn *pool.Conn, database, tblName string, parser string) (err error) {
	if c.taskCfg.ShardingKey != "" {
		return
	}
	if c.taskCfg.PrometheusSchema {
		return
	}
	// get engine
	query := fmt.Sprintf("SELECT engine FROM system.tables WHERE database = '%s' AND table = '%s'",
		database, tblName)
	util.Logger.Info(fmt.Sprintf("executing sql=> %s", query), zap.String("task", c.taskCfg.Name))
	var engine string
	err = conn.QueryRow(query).Scan(&engine)
	if err != nil {
		return
	}
	//get sortingkey
	if strings.Contains(engine, "Replacing") {
		query = fmt.Sprintf("SELECT name, type FROM system.columns WHERE (database = '%s') AND (table = '%s') AND (is_in_sorting_key = 1)",
			database, tblName)
		util.Logger.Info(fmt.Sprintf("executing sql=> %s", query), zap.String("task", c.taskCfg.Name))
		rows, _ := conn.Query(query)
		sortingKeys := make([]string, 0)
		for rows.Next() {
			var sortingkey, typ string
			err = rows.Scan(&sortingkey, &typ)
			if err != nil {
				return
			}
			sortingKeys = append(sortingKeys, sortingkey)
			c.SortingKeys = append(c.SortingKeys, &model.ColumnWithType{
				Name:       sortingkey,
				Type:       model.WhichType(typ),
				SourceName: util.GetSourceName(parser, sortingkey),
			})
		}
		rows.Close()
		util.Logger.Info(fmt.Sprintf("sortingKeys: %v", sortingKeys),
			zap.String("db", database),
			zap.String("table", tblName),
			zap.String("task", c.taskCfg.Name))
		var version string
		if err = conn.QueryRow("SELECT version()").Scan(&version); err != nil {
			version = "1.0.0.0"
		}
		var onCluster string
		if c.cfg.Clickhouse.Cluster != "" {
			onCluster = fmt.Sprintf("ON CLUSTER `%s`", c.cfg.Clickhouse.Cluster)
		}
		query = fmt.Sprintf("ALTER TABLE `%s`.`%s` %s ADD COLUMN IF NOT EXISTS `__shardingkey` Int64",
			database, tblName, onCluster)
		if util.CompareClickHouseVersion(version, "23.3") >= 0 {
			query += " SETTINGS alter_sync = 0"
		}
		util.Logger.Info(fmt.Sprintf("executing sql=> %s", query), zap.String("task", c.taskCfg.Name))
		if err = conn.Exec(query); err != nil {
			return
		}

		for _, distTbl := range c.distMetricTbls {
			query := fmt.Sprintf("ALTER TABLE `%s`.`%s` %s ADD COLUMN IF NOT EXISTS `__shardingkey` Int64",
				database, distTbl, onCluster)
			util.Logger.Info(fmt.Sprintf("executing sql=> %s", query), zap.String("task", c.taskCfg.Name))
			if err = conn.Exec(query); err != nil {
				return
			}
		}
	}
	return
}

func (c *ClickHouse) EnsureSchema(state *model.DbState) (err error) {
	if c.cfg.Clickhouse.DbKey == "" {
		return
	}
	sc := pool.GetShardConn(0)
	var conn *pool.Conn
	if conn, _, err = sc.NextGoodReplica(c.cfg.Clickhouse.Ctx, 0); err != nil {
		util.Logger.Error("no connection in clickhouse", zap.String("task", c.taskCfg.Name), zap.Error(err))
		return
	}
	// check if table exists
	var tableExists bool
	query := fmt.Sprintf("SELECT count() FROM system.tables WHERE (database = '%s') AND (table = '%s')",
		state.DB, c.TableName)
	util.Logger.Info(fmt.Sprintf("executing sql=> %s", query), zap.String("task", c.taskCfg.Name))
	var count uint64
	if err = conn.QueryRow(query).Scan(&count); err == nil {
		if count > 0 {
			tableExists = true
		} else {
			util.Logger.Info("table not exists, creating tables", zap.Uint64("count", count))
		}
	} else {
		util.Logger.Error("executing sql=> %s", zap.String("task", c.taskCfg.Name), zap.Error(err))
	}

	if !tableExists {
		util.Logger.Info("table not exists, creating tables")
		query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s ON CLUSTER `%s`", state.DB, c.cfg.Clickhouse.Cluster)
		util.Logger.Info(fmt.Sprintf("executing sql=> %s", query), zap.String("task", c.taskCfg.Name))
		if err = conn.Exec(query); err != nil {
			util.Logger.Error(fmt.Sprintf("executing sql=> %s", query), zap.String("task", c.taskCfg.Name), zap.Error(err))
			return
		}
		var createSqls []string
		var tables []string
		tables = append(tables, c.TableName)
		if c.seriesTbl != "" {
			tables = append(tables, c.seriesTbl)
		}

		if len(c.distMetricTbls) > 0 {
			tables = append(tables, c.distMetricTbls...)
		}
		if len(c.distSeriesTbls) > 0 {
			tables = append(tables, c.distSeriesTbls...)
		}
		util.Logger.Info("tables to create", zap.Int("count", len(tables)), zap.Any("tables", tables))
		for _, tbl := range tables {
			var createSql string
			createSql, err = genCreateSql(conn, c.Base.DB, tbl, state.DB, c.cfg.Clickhouse.Cluster)
			if err != nil {
				util.Logger.Error("failed to gen create sql", zap.String("task", c.taskCfg.Name), zap.Error(err))
				return
			}
			createSqls = append(createSqls, createSql)
		}
		for _, createSql := range createSqls {
			util.Logger.Info(fmt.Sprintf("executing sql=> %s", createSql), zap.String("task", c.taskCfg.Name))
			if err = conn.Exec(createSql); err != nil {
				util.Logger.Error(fmt.Sprintf("executing sql=> %s", createSql), zap.String("task", c.taskCfg.Name), zap.Error(err))
			}
		}
	}
	newState, err := c.initSchema(state.DB)
	if err != nil {
		util.Logger.Error("failed to init schema", zap.String("task", c.taskCfg.Name), zap.String("database", state.DB), zap.Error(err))
		return
	}
	state.PrepareSQL = newState.PrepareSQL
	state.PromSerSQL = newState.PromSerSQL
	state.IdxSerID = newState.IdxSerID
	state.NumDims = newState.NumDims
	state.Dims = make([]*model.ColumnWithType, len(newState.Dims))
	copy(state.Dims, newState.Dims)

	return
}

func genCreateSql(conn *pool.Conn, database, table, target, cluster string) (string, error) {
	query := fmt.Sprintf(`SELECT replaceRegexpAll(replaceRegexpOne(create_table_query, 'CREATE TABLE( IF NOT EXISTS)?\\s+\\w+\\.\\w+', 'CREATE TABLE IF NOT EXISTS %s.%s ON CLUSTER %s'), '/clickhouse/tables/\\{cluster\\}/%s/', '/clickhouse/tables/{cluster}/%s/') AS create_sql
FROM system.tables
WHERE (database = '%s') AND (table = '%s')`,
		target, table, cluster, database, target, database, table)
	util.Logger.Info(fmt.Sprintf("executing sql=> %s", query))
	var createSql string
	if err := conn.QueryRow(query).Scan(&createSql); err != nil {
		return "", err
	}
	if strings.Contains(createSql, "Distributed") {
		createSql = strings.ReplaceAll(createSql,
			fmt.Sprintf("Distributed('%s', '%s'", cluster, database),
			fmt.Sprintf("Distributed('%s', '%s'", cluster, target))
	}
	createSql = distEngineReg.ReplaceAllString(createSql, fmt.Sprintf("${1}%s${2}", target))
	return createSql, nil
}
