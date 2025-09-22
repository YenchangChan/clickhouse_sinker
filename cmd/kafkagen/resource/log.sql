CREATE TABLE default.log_test ON CLUSTER test1 (
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
	 ORDER BY (level, timestamp);

CREATE TABLE default.dist_log_test ON CLUSTER test1 AS default.log_test
ENGINE = Distributed(test1, default, log_test, rand());

CREATE TABLE IF Not EXISTS default.dist_logic_log_test ON cluster test1 as default.log_test
    ENGINE = Distributed('test', 'default', 'log_test', rand());
