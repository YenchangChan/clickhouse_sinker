package main

import (
	"context"
	"strings"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/pkg/errors"
)

type KafkaConfig struct {
	Brokers string
	Topic   string
}

type KafkaFranz struct {
	cfg     *KafkaConfig
	cl      *kgo.Client
	ctx     context.Context
	cancel  context.CancelFunc
	batches []*kgo.Record
}

// NewKafkaFranz get instance of kafka reader
func NewKafkaFranz() *KafkaFranz {
	return &KafkaFranz{}
}

const (
	Krb5KeytabAuth            = 2
	CommitRetries             = 6
	RetryBackoff              = 5 * time.Second
	defaultKerberosConfigPath = "/etc/krb5.conf"
)

// normallize and validate configuration
func (k *KafkaFranz) Normallize() (err error) {
	if k.cfg.Brokers == "" {
		err = errors.Errorf("invalid configuration")
		return
	}
	return
}

// Init Initialise the kafka instance with configuration
func (k *KafkaFranz) Init(cfg *KafkaConfig) (err error) {
	k.cfg = cfg
	k.ctx, k.cancel = context.WithCancel(context.Background())
	kfkCfg := cfg

	if err := k.Normallize(); err != nil {
		return err
	}

	opts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(kfkCfg.Brokers, ",")...),
		kgo.DefaultProduceTopic(kfkCfg.Topic),
	}

	opts = append(opts,
		kgo.AllowAutoTopicCreation(),
	)

	if k.cl, err = kgo.NewClient(opts...); err != nil {
		return err
	}
	return nil
}

func (k *KafkaFranz) Producer(message []byte) error {
	record := &kgo.Record{
		Value: message,
	}
	// Alternatively, ProduceSync exists to synchronously produce a batch of records.
	if k.cl == nil {
		return nil
	}
	var err error
	k.cl.Produce(k.ctx, record, func(r *kgo.Record, e error) {
		if e != nil {
			err = e
		}
	})
	return err
}

func (k *KafkaFranz) Stop() {
	k.cl.Close()
}
