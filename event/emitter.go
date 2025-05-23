package event

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/bondhan/golib/config"
	"github.com/bondhan/golib/log"
	"github.com/bondhan/golib/util"
	"github.com/sirupsen/logrus"
)

const (
	MetaHash    = "hash"
	MetaTime    = "timestamp"
	MetaEvent   = "event"
	MetaVersion = "version"
	MetaDefault = "default"
)

var (
	senders = map[string]SenderFactory{
		"logger": EventLoggerSender,
	}
	writers = map[string]WriterFactory{
		"logger": EventLoggerWriter,
	}
)

type (
	Publisher interface {
		Publish(ctx context.Context, event, key string, message interface{}, metadata map[string]interface{}) error
	}

	Sender interface {
		Send(ctx context.Context, message *EventMessage) error
		As(i interface{}) bool
	}

	Writer interface {
		Sender
		Delete(ctx context.Context, message *EventMessage) error
	}

	EventMessage struct {
		Topic    string                 `json:"-"`
		Key      string                 `json:"-"`
		Data     interface{}            `json:"data,omitempty" mapstructure:"data"`
		Metadata map[string]interface{} `json:"metadata,omitempty" mapstructure:"metadata"`
		RawData  []byte                 `json:"-"` // To provide raw data to consumer
	}

	EventConfig struct {
		TopicPrefix string                            `json:"topic_prefix,omitempty" mapstructure:"topic_prefix"`
		Metadata    map[string]map[string]interface{} `json:"metadata,omitempty" mapstructure:"metadata"`
		EventMap    map[string]string                 `json:"event_map,omitempty" mapstructure:"event_map"`
		GroupMap    map[string]string                 `json:"group_map,omitempty" mapstructure:"group_map"`
	}

	DriverConfig struct {
		Type   string      `json:"type" mapstructure:"type"`
		Config interface{} `json:"config" mapstructure:"config"`
	}

	Emitter struct {
		sender      Sender
		writer      Writer
		channel     chan *EventMessage
		eventConfig *EventConfig
	}

	EmitterConfig struct {
		Sender      *DriverConfig `json:"sender" mapstructure:"sender"`
		Writer      *DriverConfig `json:"writer" mapstructure:"writer"`
		EventConfig *EventConfig  `json:"event_config" mapstructure:"event_config"`
	}

	SenderFactory func(ctx context.Context, config interface{}) (Sender, error)
	WriterFactory func(ctx context.Context, config interface{}) (Writer, error)
)

func RegisterSender(name string, fn SenderFactory) {
	senders[name] = fn
}

func RegisterWriter(name string, fn WriterFactory) {
	writers[name] = fn
}

func (m *EventMessage) ToBytes() ([]byte, error) {
	return json.Marshal(m)
}

func (m *EventMessage) Hash() (string, error) {
	return hash(m)
}

func (m *EventMessage) GetMeta(key string) interface{} {
	if m.Metadata == nil {
		return nil
	}
	return m.Metadata[key]
}

func (c *EventConfig) getTopic(event string) string {
	if t, ok := c.EventMap[event]; ok {
		return c.TopicPrefix + t
	}
	return event
}

func NewEventConfig() *EventConfig {
	return &EventConfig{
		Metadata: make(map[string]map[string]interface{}),
		EventMap: make(map[string]string),
	}
}

func (c *EventConfig) getMetadata(event string) map[string]interface{} {
	if m, ok := c.getMetadataCopy(event); ok {
		m[MetaEvent] = event
		return m
	}
	return c.getDefaultMetadata(event)
}

func (c *EventConfig) getDefaultMetadata(event string) map[string]interface{} {
	if m, ok := c.getMetadataCopy(MetaDefault); ok {
		m[MetaEvent] = event
		return m
	}

	return map[string]interface{}{
		MetaVersion: 1,
		MetaEvent:   event,
	}
}

// getMetadataCopy return copy of metadata map if available
func (c EventConfig) getMetadataCopy(name string) (map[string]interface{}, bool) {
	if m, ok := c.Metadata[name]; ok {
		copyMap := map[string]interface{}{}
		for k, v := range m {
			copyMap[k] = v
		}
		return copyMap, true
	}
	return nil, false
}

func hash(m interface{}) (string, error) {
	mb, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	k := sha256.Sum256(mb)

	return string(base64.StdEncoding.EncodeToString(k[:])), nil
}

func New(ctx context.Context, conf interface{}) (*Emitter, error) {

	var econfig *EmitterConfig

	switch c := conf.(type) {
	case *EmitterConfig:
		econfig = c
	case EmitterConfig:
		econfig = &c
	case config.Getter:
		var cfg EmitterConfig
		if err := c.Unmarshal(&cfg); err != nil {
			return nil, err
		}
		econfig = &cfg
	default:
		var cfg EmitterConfig
		if err := util.DecodeJSON(c, &cfg); err != nil {
			return nil, err
		}
		econfig = &cfg
	}

	if econfig == nil {
		return nil, errors.New("[event/emitter] missing config")
	}

	em := &Emitter{
		channel:     make(chan *EventMessage),
		eventConfig: econfig.EventConfig,
	}

	if econfig.Sender == nil {
		econfig.Sender = &DriverConfig{Type: "logger"}
		log.GetLogger(ctx, "event/emitter", "New").Info("empty sender, using logger by default")
		// return nil, errors.New("[event/emitter] missing sender driver config")
	}

	if em.eventConfig == nil {
		em.eventConfig = NewEventConfig()
	}

	sf, ok := senders[econfig.Sender.Type]
	if !ok {
		return nil, errors.New("[event/emitter] unsupported sender driver")
	}

	sd, err := sf(ctx, econfig.Sender.Config)
	if err != nil {
		return nil, err
	}
	em.sender = sd

	if econfig.Writer != nil {

		wf, ok := writers[econfig.Writer.Type]
		if !ok {
			return nil, errors.New("[event/emitter] unsupported writer driver")
		}

		wr, err := wf(ctx, econfig.Writer.Config)
		if err != nil {
			return nil, err
		}

		em.writer = wr
		log.GetLogger(ctx, "event/emitter", "New").Info("enable hybrid mode")
		// running hybrid mode
		// don't use parent context on routine
		// because it might be canceled from parent routine when they finish
		// causing whatever logic inside the routine to be canceled right away
		// when they checking if the context is done
		go em.worker(context.Background())
	}

	return em, nil
}

func (e *Emitter) Publish(ctx context.Context, event, key string, message interface{}, metadata map[string]interface{}) error {
	logger := log.WithContext(ctx).WithFields(logrus.Fields{
		"pkg":      "event",
		"function": "Publish",
	})

	if e.sender == nil {
		logger.Error("driver is not set")
		return errors.New("[event/emitter] driver is not set")
	}

	topic := e.eventConfig.getTopic(event)

	md := e.eventConfig.getMetadata(event)

	if metadata == nil {
		metadata = map[string]interface{}{}
	}

	for k, v := range md {
		metadata[k] = v
	}

	mhash, err := hash(message)
	if err != nil {
		return err
	}

	metadata[MetaHash] = mhash
	metadata[MetaTime] = time.Now()

	msg := &EventMessage{
		Topic:    topic,
		Key:      key,
		Data:     message,
		Metadata: metadata,
	}

	if e.writer != nil {
		// Using hybrid mode
		if err := e.writer.Send(ctx, msg); err != nil {
			return err
		}
		e.channel <- msg
		return nil
	}

	return e.sender.Send(ctx, msg)
}

func (e *Emitter) GetTopicName(event string) string {
	return e.eventConfig.getTopic(event)
}

func (e *Emitter) As(i interface{}) bool {
	return e.sender.As(i)
}

func (e *Emitter) worker(ctx context.Context) {
	logger := log.WithContext(ctx).WithFields(logrus.Fields{
		"pkg":      "event",
		"function": "worker",
	})
	logger.Info("Running event emitter in hybrid mode")
	for msg := range e.channel {

		if err := e.sender.Send(ctx, msg); err != nil {
			logger.WithError(err).Error("Error sending message through sender")
			continue
		}

		if err := e.writer.Delete(ctx, msg); err != nil {
			logger.WithError(err).Error("Error deleting message through writer")
			continue
		}
	}
}
