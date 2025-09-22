package main

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/housepower/clickhouse_sinker/util"
	"go.uber.org/zap"
)

const (
	_        = iota
	KB int64 = 1 << (10 * iota)
	MB
	GB
	TB
	PB
)

func ReadableSize(size int64) string {
	if size < KB {
		return fmt.Sprintf("%.2f B", float64(size)/float64(1))
	} else if size < MB {
		return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
	} else if size < GB {
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	} else if size < TB {
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	} else if size < PB {
		return fmt.Sprintf("%.2f TB", float64(size)/float64(TB))
	} else {
		return fmt.Sprintf("%.2f PB", float64(size)/float64(PB))
	}
}

func randInt(n int) int {
	return int(time.Now().UnixNano() % int64(n))
}

func randValue(min, max float64) float64 {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	return min + rand.Float64()*(max-min)
}

func setKeys(keys int) {
	Regions = strings.Split(regionList, "\n")
	if keys > len(Regions) {
		util.Logger.Warn("keys is larger than region list, truncate it")
		keys = len(Regions)
	}
	Regions = Regions[:keys]
	util.Logger.Info("set keys", zap.Any("keys", Regions))
}

func selectKey(n int) string {
	return fmt.Sprintf("key%d", n)
}

func md5sum(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return base64.RawStdEncoding.EncodeToString(h.Sum(nil))
}
