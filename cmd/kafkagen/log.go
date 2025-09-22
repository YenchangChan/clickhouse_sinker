package main

import (
	"encoding/json"
	"fmt"
	"time"

	lorem "github.com/drhodes/golorem"
)

/**
CREATE TABLE sinker.log_test ON CLUSTER abc (
	level String,
	timestamp DateTime64(3),
	message String,
	path String,
	row_number Int64,
	ip String,
	hostname String,
	region String
	) ENGINE = ReplicatedMergeTree()
	 PARTITION BY toYYYYMMDD(timestamp)
	 ORDER BY (level, timestamp)；

CREATE TABLE sinker.dist_log_test ON CLUSTER abc AS sinker.log_test
ENGINE = Distributed(abc, sinker, log_test, rand());
)


*/

type Log struct {
	Level     string `json:"level"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	Path      string `json:"path"`
	RowNumber int    `json:"row_number"`
	Ip        string `json:"ip"`
	Hostname  string `json:"hostname"`
	Region    string `json:"region"`
}

var (
	LogLevels    = []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
	LogPaths     = []string{"/var/log/nginx/access.log", "/var/log/nginx/error.log"}
	LogHosts     = []string{"192.168.1.1", "192.168.1.2", "192.168.1.3", "192.168.1.4", "192.168.1.5", "192.168.1.6", "192.168.1.7", "192.168.1.8", "192.168.1.9", "192.168.1.10", "192.168.1.11", "192.168.1.12"}
	LogHostNames = []string{"host1", "host2", "host3", "host4", "host5", "host6", "host7", "host8", "host9", "host10", "host11", "host12"}
	//Regions   = []string{"CN", "US", "EU", "Asia", "Africa", "Australia"}

)

func newLog(n int) Log {
	ts := time.Now().Format("2006-01-02T15:04:05Z07:00")
	lvl := LogLevels[randInt(len(LogLevels))]
	idx := randInt(len(LogHosts))
	log := Log{
		Level:     lvl,
		Timestamp: ts,
		Message:   fmt.Sprintf("%s %s %s", ts, lvl, lorem.Sentence(5, 30)),
		Path:      LogPaths[randInt(len(LogPaths))],
		RowNumber: randInt(1000000),
		Ip:        LogHosts[idx],
		Hostname:  LogHostNames[idx],
		Region:    Regions[randInt(len(Regions))],
	}
	//fmt.Println(log.String())
	return log
}

func (l Log) Byte() []byte {
	raw, _ := json.Marshal(l)
	return raw
}

func (l Log) String() string {
	return string(l.Byte())
}

// func genLog(kgo *KafkaFranz, lines, keys int) error {
// 	var wg sync.WaitGroup
// 	wg.Add(keys)
// 	for k := 0; k < keys; k++ {
// 		go func(k int) {
// 			defer wg.Done()
// 			for i := 0; i < lines; i++ {
// 				record := newLog(i).Byte()
// 				if err := kgo.Producer(record); err != nil {
// 					util.Logger.Error("produce error", zap.Error(err))
// 					continue
// 				}
// 				atomic.AddInt64(&statics.TotalCnt, 1)
// 				atomic.AddInt64(&statics.TotalBytes, int64(len(record)))
// 			}
// 		}(k)
// 	}
// 	wg.Wait()
// 	return nil
// }
