# Config Items
> Here we use json with comments for documentation

```
{
  // clickhouse configs, it's map[string]ClickHouse for multiple clickhouse
  "clickhouse": {
    // hosts for connection, it's Array(Array(String))
    // we can put hosts with same shard into the inner array
    // it helps data deduplication for ReplicateMergeTree when driver error occurs
    "hosts": [
      [
        "127.0.0.1"
      ]
    ],
    "port": 9000,
    "username": "default"
    "password": "",
    "db": "default",  // database name
    // retryTimes when error occurs in inserting datas
    "retryTimes": 0,
  },

  // kafka configs
  "kafka": {
    "brokers": "127.0.0.1:9093",

    // SSL
    "tls": {
      "enable": false,
      // Required. It's the CA certificate with which Kafka brokers certs be signed.
      "caCertFiles": "/etc/security/ca-cert",
      // Required if Kafka brokers require client authentication.
      clientCertFile: "",
      // Required if and only if ClientCertFile is present.
      clientKeyFile: "",
    }

    // SASL
    "sasl": {
      "enable": false,
      // Mechanism is the name of the enabled SASL mechanism.
      // Possible values: PLAIN, SCRAM-SHA-256, SCRAM-SHA-512, GSSAPI (defaults to PLAIN)
      "mechanism": "PLAIN",
      // Username is the authentication identity (authcid) to present for
      // SASL/PLAIN or SASL/SCRAM authentication
      "username": "",
      // Password for SASL/PLAIN or SASL/SCRAM authentication
      "password": "",
      "gssapi": {
        // authtype - 1. KRB5_USER_AUTH, 2. KRB5_KEYTAB_AUTH
        "authtype": 0,
        "keytabpath": "",
        "kerberosconfigpath": "",
        "servicename": "",
        "username": "",
        "password": "",
        "realm": "",
        "disablepafxfast": false
      }
    },

    // kafka version, if you use sarama, the version must be specified
    "version": "2.2.1"
  },

  "task": {
    "name": "daily_request",
    // kafka topic
    "topic": "topic",
    // kafka consume from earliest or latest
    "earliest": true,
    // kafka consumer group
    "consumerGroup": "group",

    // message parser
    "parser": "json",

    // clickhouse table name
    "tableName": "daily",

    // columns of the table
    "dims": [
      {
        "name": "day",
        "type": "Date",
        "sourceName": "day"
      },
      ...
    ],

    // if it's specified, the schema will be auto mapped from clickhouse,
    "autoSchema" : true,
    // "this columns will be excluded by insert SQL "
    "excludeColumns": []

    // (experiment feature) detect new fields and their type, and add columns to the ClickHouse table accordingly. This feature requires parser be "fastjson", and support following ClickHouse data types: Int64, Float64, String.
    "dynamicSchema": {
      // whether enable this feature, default to false
      "enable": true,
      // cluster the ClickHouse node belongs
      "cluster": "test",
      // distributed table name prefix, default to "dist_"
      "distTblPrefix": ""
    },

    // shardingKey is the column name to which sharding against
    "shardingKey": "",
    // shardingPolicy is `stripe,<interval>`(requires ShardingKey be numerical) or `hash`(requires ShardingKey be string)
    "shardingPolicy": "",

    // interval of flushing the batch
    "flushInterval": 5,
    // batch size to insert into clickhouse. sinker will round upward it to the the nearest 2^n.
    "bufferSize": 90000,
    // min batch size to insert into clickhouse. sinker will round upward it to the the nearest 2^n.
    "minBufferSize": 1,
    // estimated avg message size. kafka-go needs this to determize receive buffer size. default to 1000.
    "msgSizeHint": 1000,

    // Date format in message, default to "2006-01-02".
    "layoutDate": "",
    // DateTime format in message, default to "2006-01-02T15:04:05Z07:00" (aka time.RFC3339).
    "layoutDateTime": "",
    // DateTime64 format in message, default to "2006-01-02T15:04:05.999999999Z07:00" (aka time.RFC3339Nano).
    "layoutDateTime64": "",
    // In the absence of time zone information, interprets the time as in the given location. Default to "Local" (aka /etc/localtime of the machine on which sinker runs)
    "timezone": ""
  },

  // log level, possible value: panic, fatal, error, warn, warning, info, debug, trace
  "logLevel": "debug"
}
```
