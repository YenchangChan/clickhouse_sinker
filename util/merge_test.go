package util

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type ClickHouseConfig struct {
	Cluster  string     `merge:"clickhouse.cluster"`
	DB       string     `merge:"clickhouse.db"`
	Hosts    [][]string `merge:"clickhouse.shard"`
	Port     int        `merge:"clickhouse.port"`
	Username string     `merge:"clickhouse.username"`
	Password string     `merge:"clickhouse.password"`
	Protocol string     //native, http

	// Whether enable TLS encryption with clickhouse-server
	Secure bool
	// Whether skip verify clickhouse-server cert
	InsecureSkipVerify bool
	RetryTimes         int  `merge:"clickhouse.retryTimes"` // <=0 means retry infinitely
	MaxOpenConns       int  `merge:"clickhouse.maxOpenConns"`
	ReadTimeout        int  `merge:"clickhouse.readTimeout"`
	AsyncInsert        bool `merge:"clickhouse.asyncInsert"`
	AsyncSettings      struct {
		// refers to https://clickhouse.com/docs/en/operations/settings/settings#async-insert
		AsyncInsertMaxDataSize    int `json:"async_insert_max_data_size,omitempty"`
		AsyncInsertMaxQueryNumber int `json:"async_insert_max_query_number,omitempty"` // 450
		AsyncInsertBusyTimeoutMs  int `json:"async_insert_busy_timeout_ms,omitempty"`  // 200
		WaitforAsyncInsert        int `json:"wait_for_async_insert,omitempty"`
		WaitforAsyncInsertTimeout int `json:"wait_for_async_insert_timeout,omitempty"`
		AsyncInsertThreads        int `json:"async_insert_threads,omitempty"` // 16
		AsyncInsertDeduplicate    int `json:"async_insert_deduplicate,omitempty"`
	}
	Ctx context.Context `json:"-"`
}

func TestMergeCK(t *testing.T) {
	jsonData := []byte(`{
		"clickhouse": {
			"db": "test",
			"shard": [
				["127.0.0.1", "127.0.0.2"],
				["127.0.0.3", "127.0.0.4"]
			],
			"username": "test",
			"password": "test",
			"asyncSettings": {
				"async_insert_max_data_size": 1000000
			}
		}
}`)
	var cfg ClickHouseConfig
	cfg.DB = "default"
	// cfg.Hosts = [][]string{
	// 	{"10.0.12.2"},
	// 	{"10.0.12.3"},
	// }
	cfg.AsyncInsert = true
	cfg.Secure = true
	cfg.Protocol = "http"
	cfg.Password = "123456"
	raw, err := json.MarshalIndent(cfg, "  ", "  ")
	assert.Nil(t, err)
	//fmt.Println(string(raw))
	err = MergeConfig(&cfg, jsonData)
	assert.Nil(t, err)
	raw, err = json.MarshalIndent(cfg, "  ", "  ")
	assert.Nil(t, err)
	fmt.Println(string(raw))
}

type KafkaConfig struct {
	Brokers    string `merge:"kafka.brokers"`
	Properties struct {
		HeartbeatInterval      int `json:"heartbeat.interval.ms" merge:"kafka.properties.heartbeatInterval"`
		SessionTimeout         int `json:"session.timeout.ms" merge:"kafka.properties.sessionTimeout"`
		RebalanceTimeout       int `json:"rebalance.timeout.ms" merge:"kafka.properties.rebalanceTimeout"`
		RequestTimeoutOverhead int `json:"request.timeout.ms" merge:"kafka.properties.requestTimeoutOverhead"`
		MaxPollInterval        int `json:"max.poll.interval.ms" merge:"kafka.properties.maxPollInterval"`
	}
	ResetSaslRealm bool              `merge:"kafka.resetSaslRealm"`
	Security       map[string]string `merge:"kafka.security"`
	TLS            struct {
		Enable         bool
		CaCertFiles    string // CA cert.pem with which Kafka brokers certs be signed.  Leave empty for certificates trusted by the OS
		ClientCertFile string // Required for client authentication. It's client cert.pem.
		ClientKeyFile  string // Required if and only if ClientCertFile is present. It's client key.pem.

		TrustStoreLocation string // JKS format of CA certificate, used to extract CA cert.pem.
		TrustStorePassword string
		KeystoreLocation   string // JKS format of client certificate and key, used to extrace client cert.pem and key.pem.
		KeystorePassword   string
		EndpIdentAlgo      string
	}
	// simplified sarama.Config.Net.SASL to only support SASL/PLAIN and SASL/GSSAPI(Kerberos)
	Sasl struct {
		// Whether or not to use SASL authentication when connecting to the broker
		// (defaults to false).
		Enable bool
		// Mechanism is the name of the enabled SASL mechanism.
		// Possible values: PLAIN, SCRAM-SHA-256, SCRAM-SHA-512, GSSAPI (defaults to PLAIN)
		Mechanism string
		// Username is the authentication identity (authcid) to present for
		// SASL/PLAIN or SASL/SCRAM authentication
		Username string
		// Password for SASL/PLAIN or SASL/SCRAM authentication
		Password string
		GSSAPI   struct {
			AuthType           int // 1. KRB5_USER_AUTH, 2. KRB5_KEYTAB_AUTH
			KeyTabPath         string
			KerberosConfigPath string
			ServiceName        string
			Username           string
			Password           string
			Realm              string
			DisablePAFXFAST    bool
		}
	}
	AssignInterval  int
	CalcLagInterval int
	RebalanceByLags bool
}

func TestMergeKafka(t *testing.T) {
	jsonData := []byte(`{
		"kafka": {
			"brokers": "192.168.2.7:9092,192.168.2.8:9092",
			"group": "test",
			"auth": {
				"type": "sasl",
				"sasl": {
					"mechanism": "PLAIN",
					"username": "test",
					"password": "test"
				}
			},
			"properties":{
				"rebalanceTimeout": 8888,
				"maxPollInterval": 3600000
			},
			"security": {
				"key2": "newValue2",
				"key4": "newValue4"
			}
		}
	}`)

	var cfg KafkaConfig
	// cfg.Brokers = "192.168.3.70:9092,192.168.3.71:9092"
	cfg.ResetSaslRealm = true
	cfg.Security = map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	err := MergeConfig(&cfg, jsonData)
	assert.Nil(t, err)
	raw, err := json.MarshalIndent(cfg, "  ", "  ")
	assert.Nil(t, err)
	fmt.Println(string(raw))
}

func TestMergeKafkaEmpty(t *testing.T) {
	jsonData := []byte(``)
	var cfg KafkaConfig
	cfg.Brokers = "192.168.3.70:9092,192.168.3.71:9092"
	cfg.ResetSaslRealm = true
	cfg.Security = map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	err := MergeConfig(&cfg, jsonData)
	assert.Nil(t, err)
	raw, err := json.MarshalIndent(cfg, "  ", "  ")
	assert.Nil(t, err)
	fmt.Println(string(raw))
}

type Config struct {
	ClickHouse ClickHouseConfig
	KafKa      KafkaConfig
}

func TestNested(t *testing.T) {
	jsonData := []byte(`{
		"clickhouse": {
			"db": "test",
			"shard": [
				["192.168.1.17:9000", "192.168.1.18:9000"],
				["192.168.1.19:9000", "192.168.1.20:9000"]
			]
		},
		"kafka": {
			"brokers": "192.168.1.17:9092,192.168.1.18:9092"
		}
	}`)

	cfg := Config{}
	err := MergeConfig(&cfg, jsonData)
	assert.Nil(t, err)
	raw, err := json.MarshalIndent(cfg, "  ", "  ")
	assert.Nil(t, err)
	fmt.Println(string(raw))
}
