package parser

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGjsonInt(t *testing.T) {
	pp := NewParserPool("gjson", nil, "", nil)
	parser := pp.Get()
	defer pp.Put(parser)
	metric, _ := parser.Parse(jsonSample)

	var exp, act int64
	exp = 1536813227
	act = metric.GetInt("its", false).(int64)
	require.Equal(t, exp, act)

	exp = 0
	act = metric.GetInt("not_exist", false).(int64)
	require.Equal(t, exp, act)

	actual := metric.GetInt("not_exist", true)
	require.Nil(t, actual, "err should be nothing")

	exp = 0
	act = metric.GetInt("bool_false", false).(int64)
	require.Equal(t, exp, act)

	exp = 1
	act = metric.GetInt("bool_true", false).(int64)
	require.Equal(t, exp, act)
}

func TestGjsonFloat(t *testing.T) {
	pp := NewParserPool("gjson", nil, "", nil)
	parser := pp.Get()
	defer pp.Put(parser)
	metric, _ := parser.Parse(jsonSample)

	var exp, act float64
	exp = 0.11
	act = metric.GetFloat("percent", false).(float64)
	require.Equal(t, exp, act)

	exp = 0.0
	act = metric.GetFloat("not_exist", false).(float64)
	require.Equal(t, exp, act)

	actual := metric.GetFloat("not_exist", true)
	require.Nil(t, actual, "err should be nothing")
}

func TestGjsonString(t *testing.T) {
	pp := NewParserPool("gjson", nil, "", nil)
	parser := pp.Get()
	defer pp.Put(parser)
	metric, _ := parser.Parse(jsonSample)

	var exp, act string
	exp = "ws"
	act = metric.GetString("channel", false).(string)
	require.Equal(t, exp, act)

	exp = ""
	act = metric.GetString("not_exist", false).(string)
	require.Equal(t, exp, act)

	actual := metric.GetString("not_exist", true)
	require.Nil(t, actual, "err should be nothing")
}

func TestGjsonDate(t *testing.T) {
	pp := NewParserPool("gjson", nil, "", TSLayoutStd)
	parser := pp.Get()
	defer pp.Put(parser)
	metric, _ := parser.Parse(jsonSample)

	var exp, act time.Time
	exp = time.Date(2019, 12, 16, 0, 0, 0, 0, time.Local)
	act = metric.GetDate("date1", false).(time.Time)
	require.Equal(t, exp, act)

	exp = time.Time{}
	act = metric.GetDate("not_exist", false).(time.Time)
	require.Equal(t, exp, act)

	actual := metric.GetDate("not_exist", true)
	require.Nil(t, actual, "err should be nothing")
}

func TestGjsonDateTimeStd(t *testing.T) {
	pp := NewParserPool("gjson", nil, "", TSLayoutStd)
	parser := pp.Get()
	defer pp.Put(parser)
	metric, _ := parser.Parse(jsonSample)

	var exp, act time.Time
	exp = time.Date(2019, 12, 16, 12, 10, 30, 0, time.UTC)
	act = metric.GetDateTime("time_sec_rfc3339_1", false).(time.Time)
	require.Equal(t, exp, act)

	exp = time.Date(2019, 12, 16, 12, 10, 30, 0, time.FixedZone("CST", 8*60*60)).In(time.UTC)
	act = metric.GetDateTime("time_sec_rfc3339_2", false).(time.Time).In(time.UTC)
	require.Equal(t, exp, act)

	exp = time.Time{}
	act = metric.GetDateTime("not_exist", false).(time.Time)
	require.Equal(t, exp, act)

	actual := metric.GetDateTime("not_exist", true)
	require.Nil(t, actual, "err should be nothing")
}

func TestGjsonDateTimeCh(t *testing.T) {
	pp := NewParserPool("gjson", nil, "", TSLayoutCh)
	parser := pp.Get()
	defer pp.Put(parser)
	metric, _ := parser.Parse(jsonSample)

	var exp, act time.Time
	exp = time.Date(2019, 12, 16, 12, 10, 30, 0, time.Local).In(time.UTC)
	act = metric.GetDateTime("time_sec_clickhouse_1", false).(time.Time).In(time.UTC)
	require.Equal(t, exp, act)
}

func TestGjsonDateTime64Std(t *testing.T) {
	pp := NewParserPool("gjson", nil, "", TSLayoutStd)
	parser := pp.Get()
	defer pp.Put(parser)
	metric, _ := parser.Parse(jsonSample)

	var exp, act time.Time
	exp = time.Date(2019, 12, 16, 12, 10, 30, 123000000, time.UTC)
	act = metric.GetDateTime64("time_ms_rfc3339_1", false).(time.Time)
	require.Equal(t, exp, act)

	exp = time.Date(2019, 12, 16, 12, 10, 30, 123000000, time.FixedZone("CST", 8*60*60)).In(time.UTC)
	act = metric.GetDateTime64("time_ms_rfc3339_2", false).(time.Time).In(time.UTC)
	require.EqualValues(t, exp, act)

	exp = time.Time{}
	act = metric.GetDateTime64("not_exist", false).(time.Time)
	require.Equal(t, exp, act)

	actual := metric.GetDateTime64("not_exist", true)
	require.Nil(t, actual, "err should be nothing")
}

func TestGjsonDateTime64Ch(t *testing.T) {
	pp := NewParserPool("gjson", nil, "", TSLayoutCh)
	parser := pp.Get()
	defer pp.Put(parser)
	metric, _ := parser.Parse(jsonSample)

	var exp, act time.Time
	exp = time.Date(2019, 12, 16, 12, 10, 30, 123000000, time.Local).In(time.UTC)
	act = metric.GetDateTime64("time_ms_clickhouse_1", false).(time.Time).In(time.UTC)
	require.Equal(t, exp, act)
}

func TestGjsonElasticDateTime(t *testing.T) {
	pp := NewParserPool("gjson", nil, "", TSLayoutStd)
	parser := pp.Get()
	defer pp.Put(parser)
	metric, _ := parser.Parse(jsonSample)

	var exp, act int64
	// {"date": "2019-12-16T12:10:30Z"}
	// TZ=UTC date -d @1576498230 => Mon 16 Dec 2019 12:10:30 PM UTC
	exp = 1576498230
	act = metric.GetElasticDateTime("time_sec_rfc3339_1", false).(int64)
	require.Equal(t, exp, act)

	exp = -62135596800
	act = metric.GetElasticDateTime("not_exist", false).(int64)
	require.Equal(t, exp, act)

	actual := metric.GetElasticDateTime("not_exist", true)
	require.Nil(t, actual, "err should be nothing")
}

func TestGjsonArray(t *testing.T) {
	pp := NewParserPool("gjson", nil, "", nil)
	parser := pp.Get()
	defer pp.Put(parser)
	metric, _ := parser.Parse(jsonSample)

	actI := metric.GetArray("mp.i", "int").([]int64)
	expI := []int64{1, 2, 3}
	require.Equal(t, actI, expI)

	actF := metric.GetArray("mp.f", "float").([]float64)
	expF := []float64{1.1, 2.2, 3.3}
	require.Equal(t, expF, actF)

	actS := metric.GetArray("mp.s", "string").([]string)
	expS := []string{"aa", "bb", "cc"}
	require.Equal(t, expS, actS)

	actIE := metric.GetArray("mp.e", "int").([]int64)
	expIE := []int64{}
	require.Equal(t, expIE, actIE)

	actFE := metric.GetArray("mp.e", "float").([]float64)
	expFE := []float64{}
	require.Equal(t, expFE, actFE)

	actSE := metric.GetArray("mp.e", "string").([]string)
	expSE := []string{}
	require.Equal(t, expSE, actSE)
}
