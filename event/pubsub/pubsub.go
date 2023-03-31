package pubsub

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/bondhan/golib/event"
	"github.com/bondhan/golib/util"
	"github.com/mitchellh/mapstructure"
	"gocloud.dev/pubsub"
	"gocloud.dev/pubsub/kafkapubsub"
)

const (
	kafkaAuthSASL        = "SASL"
	kafkaAuthTLS         = "TLS"
	KafkaKeyName         = "key"
	SchedulerTimeKey     = "scheduler-epoch"
	SchedulerTargetTopic = "scheduler-target-topic"
	SchedulerTargetKey   = "scheduler-target-key"
	DefaultSchedulerKey  = "schedule"
)

type PubsubSender struct {
	topics              map[string]*pubsub.Topic
	schema              string
	Schema              string `json:"schema" mapstructure:"schema"`
	KafkaBrokers        string `json:"kafka_brokers" mapstructure:"kafka_brokers"`
	KafkaAuth           string `json:"kafka_auth" mapstructure:"kafka_auth"`
	KafkaCert           string `json:"kafka_cert" mapstructure:"kafka_cert"`
	KafkaKey            string `json:"kafka_key" mapstructure:"kafka_key"`
	KafkaPem            string `json:"kafka_pem" mapstructure:"kafka_pem"`
	KafkaUser           string `json:"kafka_user" mapstructure:"kafka_user"`
	KafkaPassword       string `json:"kafka_password" mapstructure:"kafka_password"`
	KafkaSchedulerTopic string `json:"kafka_scheduler_topic" mapstructure:"kafka_scheduler_topic"`
	KafkaScheduleKey    string `json:"kafka_schedule_key" mapstructure:"kafka_schedule_key"`
	Credential          string `json:"credential" mapstructure:"credential"`
}

type PubsubMessageCarrier struct {
	*pubsub.Message
}

func init() {
	event.RegisterSender("pubsub", NewPubsubSender)
}

func NewPubsubSender(ctx context.Context, config interface{}) (event.Sender, error) {
	var pub PubsubSender
	if err := mapstructure.Decode(config, &pub); err != nil {
		return nil, err
	}

	parseUri, err := url.Parse(pub.Schema)
	if err != nil {
		return nil, err
	}

	pub.schema = parseUri.Scheme

	switch parseUri.Scheme {
	case "kafka":
		os.Setenv("KAFKA_BROKERS", pub.KafkaBrokers)
		if pub.KafkaSchedulerTopic != "" && pub.KafkaScheduleKey == "" {
			pub.KafkaScheduleKey = DefaultSchedulerKey
		}
	case "gcppubsub":
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", pub.Credential)
	}

	pub.topics = make(map[string]*pubsub.Topic)

	return &pub, nil
}

func (p *PubsubSender) getTopic(ctx context.Context, topicName string) (*pubsub.Topic, error) {
	if top, ok := p.topics[topicName]; ok {
		return top, nil
	}

	var top *pubsub.Topic
	var err error

	parseUri, err := url.Parse(p.Schema)
	if err != nil {
		return nil, err
	}
	switch parseUri.Scheme {
	case "kafka":
		config, er := p.configureKafka()
		if er != nil {
			return nil, er
		}
		top, err = kafkapubsub.OpenTopic(strings.Split(p.KafkaBrokers, ","), config, topicName, &kafkapubsub.TopicOptions{KeyName: KafkaKeyName})
	case "gcppubsub":
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", p.Credential)
		top, err = pubsub.OpenTopic(ctx, p.Schema+topicName)
	default:
		top, err = pubsub.OpenTopic(ctx, p.Schema+topicName)
	}

	if err != nil {
		return nil, err
	}
	p.topics[topicName] = top
	return top, nil
}

func (p *PubsubSender) Send(ctx context.Context, message *event.EventMessage) error {

	topic, err := p.getTopic(ctx, message.Topic)
	if err != nil {
		return err
	}
	mb, err := message.ToBytes()

	if err != nil {
		return err
	}
	msg := &pubsub.Message{Body: mb}
	if message.Key != "" {
		msg.Metadata = map[string]string{KafkaKeyName: message.Key}
	}

	if ts := message.GetMeta(p.KafkaScheduleKey); ts != nil && p.schema == "kafka" && message.Key != "" && p.KafkaSchedulerTopic != "" {
		st, err := p.getTopic(ctx, p.KafkaSchedulerTopic)
		if err != nil {
			return err
		}

		if msg.Metadata == nil {
			msg.Metadata = make(map[string]string)
		}

		var schedule time.Time

		switch v := ts.(type) {
		case time.Time:
			schedule = v
		default:
			if t, err := util.DateStringToTime(fmt.Sprintf("%v", v)); err == nil {
				schedule = t
			}
		}

		if !schedule.IsZero() && schedule.After(time.Now()) {
			msg.Metadata[SchedulerTimeKey] = fmt.Sprintf("%d", schedule.Unix())
			msg.Metadata[SchedulerTargetTopic] = message.Topic
			msg.Metadata[SchedulerTargetKey] = message.Key
			msg.Metadata[KafkaKeyName] = message.Key
			topic = st
		}

	}

	if err := topic.Send(ctx, msg); err != nil {
		return err
	}

	return nil

}

func (p *PubsubSender) As(i interface{}) bool {
	for _, p := range p.topics {
		if p.As(i) {
			return true
		}
	}
	return false
}

func (p *PubsubSender) configureKafka() (*sarama.Config, error) {
	config := kafkapubsub.MinimalConfig()
	switch p.KafkaAuth {
	case kafkaAuthTLS:
		keypair, err := tls.LoadX509KeyPair(p.KafkaCert, p.KafkaKey)
		if err != nil {
			return nil, err
		}

		caCert, err := ioutil.ReadFile(p.KafkaPem)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{keypair},
			RootCAs:      caCertPool,
		}

		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
		config.Version = sarama.V0_11_0_0

	case kafkaAuthSASL:
		caCert, err := ioutil.ReadFile(p.KafkaPem)
		if err != nil {
			panic(err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		tlsConfig := &tls.Config{
			RootCAs: caCertPool,
		}

		// parse Kafka cluster version
		version, err := sarama.ParseKafkaVersion("2.4.0")
		if err != nil {
			return nil, err
		}

		// init config, enable errors and notifications
		config.Version = version
		config.Metadata.Full = true

		// Kafka SASL configuration
		config.Net.SASL.Enable = true
		config.Net.SASL.User = p.KafkaUser
		config.Net.SASL.Password = p.KafkaPassword
		config.Net.SASL.Handshake = true
		config.Net.SASL.Mechanism = sarama.SASLTypePlaintext

		// TLS configuration
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	}
	return config, nil
}

func NewPubsubMessageCarrier(msg *pubsub.Message) *PubsubMessageCarrier {
	if msg.Metadata == nil {
		msg.Metadata = make(map[string]string)
	}
	return &PubsubMessageCarrier{
		Message: msg,
	}
}

// Get retrieves a single value for a given key.
func (k PubsubMessageCarrier) Get(key string) string {
	return k.Metadata[key]
}

// Set sets a header.
func (k PubsubMessageCarrier) Set(key, val string) {
	// Ensure uniqueness of keys
	k.Metadata[key] = val
}

// Keys returns a slice of all key identifiers in the carrier.
func (k PubsubMessageCarrier) Keys() []string {
	out := make([]string, len(k.Metadata))
	i := 0
	for _, h := range k.Metadata {
		out[i] = h
	}
	return out
}
