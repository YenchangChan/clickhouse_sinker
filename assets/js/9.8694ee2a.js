(window.webpackJsonp=window.webpackJsonp||[]).push([[9],{197:function(e,n,t){"use strict";t.r(n);var a=t(3),i=Object(a.a)({},(function(){var e=this.$createElement,n=this._self._c||e;return n("ContentSlotsDistributor",{attrs:{"slot-key":this.$parent.slotKey}},[n("h1",{attrs:{id:"config-items"}},[n("a",{staticClass:"header-anchor",attrs:{href:"#config-items"}},[this._v("#")]),this._v(" Config Items")]),this._v(" "),n("blockquote",[n("p",[this._v("Here we use json with comments for documentation")])]),this._v(" "),n("div",{staticClass:"language- extra-class"},[n("pre",{pre:!0,attrs:{class:"language-text"}},[n("code",[this._v('{\n  // clickhouse configs, it\'s map[string]ClickHouse for multiple clickhouse\n  "clickhouse": {\n    // hosts for connection, it\'s Array(Array(String))\n    // we can put hosts with same shard into the inner array\n    // it helps data deduplication for ReplicateMergeTree when driver error occurs\n    "hosts": [\n      [\n        "127.0.0.1"\n      ]\n    ],\n    "port": 9000,\n    "username": "default"\n    "password": "",\n    "db": "default",  // database name\n    // retryTimes when error occurs in inserting datas\n    "retryTimes": 0,\n  },\n\n  // kafka configs\n  "kafka": {\n    "brokers": "127.0.0.1:9093",\n\n    // SSL\n    "tls": {\n      "enable": false,\n      // Required. It\'s the CA certificate with which Kafka brokers certs be signed.\n      "caCertFiles": "/etc/security/ca-cert",\n      // Required if Kafka brokers require client authentication.\n      clientCertFile: "",\n      // Required if and only if ClientCertFile is present.\n      clientKeyFile: "",\n    }\n\n    // SASL\n    "sasl": {\n      "enable": false,\n      // Mechanism is the name of the enabled SASL mechanism.\n      // Possible values: PLAIN, SCRAM-SHA-256, SCRAM-SHA-512, GSSAPI (defaults to PLAIN)\n      "mechanism": "PLAIN",\n      // Username is the authentication identity (authcid) to present for\n      // SASL/PLAIN or SASL/SCRAM authentication\n      "username": "",\n      // Password for SASL/PLAIN or SASL/SCRAM authentication\n      "password": "",\n      "gssapi": {\n        // authtype - 1. KRB5_USER_AUTH, 2. KRB5_KEYTAB_AUTH\n        "authtype": 0,\n        "keytabpath": "",\n        "kerberosconfigpath": "",\n        "servicename": "",\n        "username": "",\n        "password": "",\n        "realm": "",\n        "disablepafxfast": false\n      }\n    },\n\n    // kafka version, if you use sarama, the version must be specified\n    "version": "2.2.1"\n  },\n\n  "task": {\n    "name": "daily_request",\n    // kafka topic\n    "topic": "topic",\n    // kafka consume from earliest or latest\n    "earliest": true,\n    // kafka consumer group\n    "consumerGroup": "group",\n\n    // message parser\n    "parser": "json",\n\n    // clickhouse table name\n    "tableName": "daily",\n\n    // columns of the table\n    "dims": [\n      {\n        "name": "day",\n        "type": "Date",\n        "sourceName": "day"\n      },\n      ...\n    ],\n\n    // if it\'s specified, the schema will be auto mapped from clickhouse,\n    "autoSchema" : true,\n    // "this columns will be excluded by insert SQL "\n    "excludeColumns": []\n\n    // (experiment feature) detect new fields and their type, and add columns to the ClickHouse table accordingly. This feature requires parser be "fastjson", and support following ClickHouse data types: Int64, Float64, String.\n    "dynamicSchema": {\n      // whether enable this feature, default to false\n      "enable": true,\n      // cluster the ClickHouse node belongs\n      "cluster": "test",\n      // distributed table name prefix, default to "dist_"\n      "distTblPrefix": ""\n    },\n\n    // shardingKey is the column name to which sharding against\n    "shardingKey": "",\n    // shardingPolicy is `stripe,<interval>`(requires ShardingKey be numerical) or `hash`(requires ShardingKey be string)\n    "shardingPolicy": "",\n\n    // interval of flushing the batch\n    "flushInterval": 5,\n    // batch size to insert into clickhouse. sinker will round upward it to the the nearest 2^n.\n    "bufferSize": 90000,\n    // min batch size to insert into clickhouse. sinker will round upward it to the the nearest 2^n.\n    "minBufferSize": 1,\n    // estimated avg message size. kafka-go needs this to determize receive buffer size. default to 1000.\n    "msgSizeHint": 1000,\n\n    // Date format in message, default to "2006-01-02".\n    "layoutDate": "",\n    // DateTime format in message, default to "2006-01-02T15:04:05Z07:00" (aka time.RFC3339).\n    "layoutDateTime": "",\n    // DateTime64 format in message, default to "2006-01-02T15:04:05.999999999Z07:00" (aka time.RFC3339Nano).\n    "layoutDateTime64": "",\n    // In the absence of time zone information, interprets the time as in the given location. Default to "Local" (aka /etc/localtime of the machine on which sinker runs)\n    "timezone": ""\n  },\n\n  // log level, possible value: panic, fatal, error, warn, warning, info, debug, trace\n  "logLevel": "debug"\n}\n')])])])])}),[],!1,null,null,null);n.default=i.exports}}]);