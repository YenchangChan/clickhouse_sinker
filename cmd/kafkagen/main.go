package main

import (
	_ "embed"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/housepower/clickhouse_sinker/util"
	"go.uber.org/zap"
)

var (
	Mode     = kingpin.Flag("mode", "generate mode: log|monitor").Default("log").Short('m').String()
	Topic    = kingpin.Flag("topic", "kafka topic").Default("sinker_test_log").Short('t').String()
	Keys     = kingpin.Flag("keys", "db keys").Default("1").Short('c').Int()
	Brokers  = kingpin.Flag("brokers", "kafka brokers").Default("localhost:9092").Short('k').String()
	Lines    = kingpin.Flag("lines", "lines per key").Default("100000").Short('l').Int()
	LongTime = kingpin.Flag("longtime", "long time").Default("false").Short('d').Bool()

	pool    *util.WorkerPool
	statics Stastics
	done    chan struct{}
	Regions []string
)

//go:embed resource/regions.txt
var regionList string

type Stastics struct {
	Timestamp  time.Time
	LastTime   time.Time
	LastCnt    int64
	TotalCnt   int64
	TotalBytes int64
	LastBytes  int64
}

func main() {
	kingpin.Parse()
	pool = util.NewWorkerPool(10, 20)
	done = make(chan struct{})

	util.InitLogger([]string{"stdout"})
	util.Logger.Info("starting kafkagen", zap.String("mode", *Mode),
		zap.String("topic", *Topic),
		zap.Int("keys", *Keys),
		zap.String("brokers", *Brokers),
		zap.Int("lines", *Lines))

	kconf := &KafkaConfig{
		Brokers: *Brokers,
		Topic:   *Topic,
	}
	k := NewKafkaFranz()
	if err := k.Init(kconf); err != nil {
		util.Logger.Error("failed to init kafka", zap.Error(err))
		return
	}
	defer k.Stop()
	statics.Timestamp = time.Now()
	statics.LastTime = time.Now()
	statics.LastCnt = 0
	statics.TotalCnt = 0
	setKeys(*Keys)
	go kafkagen(k, *Lines, *Keys)

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			util.Logger.Info("quit due to context been canceled")
			return
		case <-ticker.C:
			windowCnt := statics.TotalCnt - statics.LastCnt
			windowBytes := statics.TotalBytes - statics.LastBytes
			window := time.Now().Sub(statics.LastTime).Seconds()
			statics.LastCnt = statics.TotalCnt
			statics.LastBytes = statics.TotalBytes
			statics.LastTime = time.Now()

			util.Logger.Info("metrics",
				zap.Int64("lines", statics.TotalCnt),
				zap.String("bytes", ReadableSize(statics.TotalBytes)),
				zap.Int64("speed(lines/s)", windowCnt/int64(window)),
				zap.String("speed(bytes/s)", ReadableSize(windowBytes/int64(window))))

		}
	}

}

func newRecord(i int) []byte {
	switch *Mode {
	case "log":
		return newLog(i).Byte()
	case "monitor":
		return newMetric(i).Byte()
	default:
		util.Logger.Error("unknown mode", zap.String("mode", *Mode))
	}
	return nil
}

func once(kgo *KafkaFranz, lines, keys int) error {
	var wg sync.WaitGroup
	wg.Add(keys)
	for k := 0; k < keys; k++ {
		go func(k int) {
			defer wg.Done()
			for i := 0; i < lines; i++ {
				record := newRecord(i)
				if err := kgo.Producer(record); err != nil {
					util.Logger.Error("produce error", zap.Error(err))
					continue
				}
				atomic.AddInt64(&statics.TotalCnt, 1)
				atomic.AddInt64(&statics.TotalBytes, int64(len(record)))
			}
		}(k)
	}
	wg.Wait()
	//util.Logger.Info("finished generating", zap.Int("lines", *Lines), zap.Int("keys", *Keys))
	return nil
}

func kafkagen(k *KafkaFranz, lines, keys int) error {
	var err error
	if *LongTime {
		for {
			once(k, lines, keys)
		}
	} else {
		once(k, lines, keys)
	}

	done <- struct{}{}
	return err
}
