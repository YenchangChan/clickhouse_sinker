CREATE TABLE IF Not EXISTS default.metric_series ON CLUSTER test1 (
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
CREATE TABLE IF Not EXISTS default.dist_metric_series ON CLUSTER test1 AS default.metric_series
    ENGINE = Distributed(test1, default, metric_series);

CREATE TABLE IF Not EXISTS default.dist_logic_metric_series ON cluster test1 as default.metric_series
    ENGINE = Distributed('test', 'default', 'metric_series', rand());


CREATE TABLE IF Not EXISTS default.metric ON CLUSTER test1 (
    timestamp DateTime CODEC(DoubleDelta, LZ4),
    value Float64 CODEC(ZSTD(15)),
    __series_id__ Int64,
    __parse_start_time__ Nullable(DateTime) CODEC(DoubleDelta, LZ4),
    __parse_end_time__ Nullable(DateTime) CODEC(DoubleDelta, LZ4)
    )  ENGINE = ReplicatedReplacingMergeTree()
    PARTITION BY toYYYYMMDD(timestamp)
    ORDER BY (__series_id__, timestamp)
    TTL toDate(`timestamp`) + toIntervalDay(7) delete;

CREATE TABLE IF Not EXISTS default.dist_metric ON CLUSTER test1 AS default.metric
    ENGINE = Distributed('test1', 'default', 'metric')  ;

CREATE TABLE IF Not EXISTS default.dist_logic_metric ON cluster test1 as default.metric
    ENGINE = Distributed('test', 'default', 'metric', rand());

