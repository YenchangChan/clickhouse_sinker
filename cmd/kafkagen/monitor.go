package main

import (
	"encoding/json"
	"fmt"
	"time"
)

/**
CREATE TABLE IF Not EXISTS sinker.metric_series ON CLUSTER abc (
    __series_id__ Int64,
    __mgmt_id__ Int64,
    labels String,
    __name__ String,
    __object_type__ String,
    __metric_type__ String,
    __series_key__ String,
    __field_key__ String,
    __object_id__ String,
    `__ttl__` DateTime DEFAULT now(),
    `__help__` Nullable(String),
    `__type__` Nullable(String)
    ) ENGINE = ReplicatedReplacingMergeTree()
    PARTITION BY xxHash64(__object_id__) % 5
    ORDER BY (__name__, __series_id__);
CREATE TABLE IF Not EXISTS sinker.dist_metric_series ON CLUSTER abc AS sinker.metric_series
    ENGINE = Distributed(abc, sinker, metric_series);

CREATE TABLE IF Not EXISTS sinker.metric ON CLUSTER abc (
    timestamp DateTime CODEC(DoubleDelta, LZ4),
    value Float64 CODEC(ZSTD(15)),
    __series_id__ Int64,
    __parse_start_time__ Nullable(DateTime) CODEC(DoubleDelta, LZ4),
    __parse_end_time__ Nullable(DateTime) CODEC(DoubleDelta, LZ4)
    )  ENGINE = ReplicatedReplacingMergeTree()
    PARTITION BY toYYYYMMDD(timestamp)
    ORDER BY (__series_id__, timestamp)
    TTL toDate(`timestamp`) + toIntervalDay(7) delete;

CREATE TABLE IF Not EXISTS sinker.dist_metric ON CLUSTER `abc AS sinker.metric
    ENGINE = Distributed(abc, sinker, metric)  ;


*/

var (
	MetricTypes = []string{"system_io", "system_cpu", "system_mem", "system_disk", "system_net", "system_process", "system_process_cpu", "system_process_mem", "system_process_io", "system_process_thread", "system_process_thread_cpu", "system_process_thread_mem", "system_process_thread_io", "system_process_thread_net", "system_process"}
	FiledKeys   = []string{"device", "device_name", "device_type", "disk_name", "disk_type", "interface_name", "interface_type", "process_name", "process_type", "thread_name", "thread_type", "thread_state", "thread_state_type", "thread_cpu_time", "thread_cpu_time_type", "thread_mem_size", "thread_mem_size"}
	SeriesKeys  = []string{"uin", "__object_id__", "__name__", "__metric_type__", "device", "device_name", "device_type", "disk_name", "disk_type", "interface_name", "interface_type", "process_name", "process_type", "thread_name", "thread_type", "thread_state", "thread_state_type", "thread_cpu_time"}
	MgmtIds     = []int64{1123074217849335582, 1123074217849335583, 1123074217849335584, 1123074217849335585, 1123074217849335586, 1123074217849335587, 1123074217849335588, 1123074217849335589, 1123074217849335590, 1123074217849335591, 1123074217849335592, 1123074217849335593, 1123074217849335594, 1123074217849335595, 1123074217849335596, 1123074217849335597, 1123074217849335598, 1123074217849335599, 1123074217849335600, 1123074217849335601, 1123074217849335602, 1123074217849335603}
	ObjectNames = []string{"dev-101-88", "dev-101-89", "dev-101-90", "dev-101-91", "dev-101-92", "dev-101-93", "dev-101-94", "dev-101-95", "dev-101-96", "dev-101-97", "dev-101-98", "dev-101-99", "dev-101-100"}
	MetricFlags = []string{"datadog", "influxdb", "prometheus", "telegraf", "open-falcon", "cloud-monitor", "cloud-monitor-v2", "cloud-monitor-v3", "cloud-monitor-v4", "cloud-monitor-v5", "cloud-monitor-v6", "cloud-monitor-v7", "cloud-monitor"}
	ObjectTypes = []string{"system", "system_io", "system_cpu", "system_mem", "system_disk", "system_net", "system_process", "system_process_cpu", "system_process_mem", "system_process_io", "system_process_thread", "system_process_thread_cpu", "system_process_thread_mem", "system_process_thread_io", "system_process_thread"}
	AgentIds    = []string{"100004604167", "100004604168", "100004604169", "100004604170", "100004604171", "100004604172", "100004604173", "100004604174", "100004604175", "100004604176", "100004604177", "100004604178", "100004604179", "100004604180", "100004604181", "100004604182", "100004604683", "100004604184", "100004604185", "100004604186", "100004604187", "100004604188", "100004604189", "100004604190", "100004604191", "100004604192", "100004604193", "100004604194", "100004604195", "100004604196", "100004604197", "100004604198"}
	SeriesIds   = []int64{2761070129987504083, 2761070129987504084, 2761070129987504085, 2761070129987504086, 2761070129987504087, 2761070129987504088, 2761070129987504089, 2761070129987504090, 2761070129987504091, 2761070129987504092, 2761070129987504093, 2761070129987504094, 2761070129987504095, 2761070129987504096, 2761070129987504097, 2761070129987504098, 2761070129987504099, 2761070129987504100, 2761070129987504101, 2761070129987504102, 2761070129987504103, 2761070129987504104, 2761070129987504105, 2761070129987504106, 2761070129987504107, 2761070129987504108, 2761070129987504109}
	ObjectIds   = []string{"1108491895784519", "1108491895784519", "1108491895784519", "1108491895784519", "1108491895784519", "1108491895784519", "1108491895784519", "1108491895784519", "1108491895784519", "1108491895784519", "1108491895784519", "110849189578451"}
	Names       = []string{"system_io_util", "system_cpu_util", "system_mem_util", "system_disk_util", "system_net_util", "system_process_util", "system_process_cpu_util", "system_process_mem_util", "system_process_io_util", "system_process_thread_util", "system_process_thread_cpu_util", "system_process"}
)

type Metric struct {
	MetricType string  `json:"__metric_type__"`
	FiledKey   string  `json:"__field_key__"`
	SeriesKey  string  `json:"__series_key__"`
	MgmtId     int64   `json:"__mgmt_id__"`
	ObjectName string  `json:"__object_name__"`
	MetricFlag string  `json:"__metric_flag__"`
	Ttl        string  `json:"__ttl__"`
	ObjectType string  `json:"__object_type__"`
	AgentId    string  `json:"__agent_id__"`
	SeriesId   int64   `json:"__series_id__"`
	ObjectId   string  `json:"__object_id__"`
	Name       string  `json:"__name__"`
	Timestamp  int64   `json:"timestamp"`
	Value      float64 `json:"value"`
	Uin        string  `json:"uin"`
	SecretKey  string  `json:"secret_key"`
}

func newMetric(i int) []byte {
	uin := Regions[randInt(len(Regions))]
	metric := Metric{
		MetricType: MetricTypes[randInt(len(MetricTypes))],
		FiledKey:   FiledKeys[randInt(len(FiledKeys))],
		SeriesKey:  SeriesKeys[randInt(len(SeriesKeys))],
		MgmtId:     MgmtIds[randInt(len(MgmtIds))],
		ObjectName: ObjectNames[randInt(len(ObjectNames))],
		MetricFlag: MetricFlags[randInt(len(MetricFlags))],
		Ttl:        time.Now().Add(time.Hour * 24).Format("2006-01-02"),
		ObjectType: ObjectTypes[randInt(len(ObjectTypes))],
		AgentId:    AgentIds[randInt(len(AgentIds))],
		SeriesId:   SeriesIds[randInt(len(SeriesIds))],
		ObjectId:   ObjectIds[randInt(len(ObjectIds))],
		Name:       Names[randInt(len(Names))],
		Timestamp:  time.Now().UnixMilli(),
		Value:      randValue(0, 100),
		Uin:        uin,
		SecretKey:  md5sum(uin),
	}
	m := make(map[string]interface{})
	json.Unmarshal([]byte(metric.String()), &m)
	n := randInt(500)
	m[selectKey(n)] = fmt.Sprintf("value%d", n)
	raw, _ := json.Marshal(&m)
	return raw
}

func (l Metric) Byte() []byte {
	raw, _ := json.Marshal(l)
	return raw
}

func (l Metric) String() string {
	return string(l.Byte())
}
