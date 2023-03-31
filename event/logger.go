package event

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/bondhan/golib/log"
)

type EventLogger struct {
	log *logrus.Entry
}

func EventLoggerSender(ctx context.Context, config interface{}) (Sender, error) {
	return NewEventLogger(ctx, config)
}

func EventLoggerWriter(ctx context.Context, config interface{}) (Writer, error) {
	return NewEventLogger(ctx, config)
}

func NewEventLogger(ctx context.Context, config interface{}) (*EventLogger, error) {
	return &EventLogger{log: log.GetLogger(ctx, "event", "logger")}, nil
}

func (e *EventLogger) Send(ctx context.Context, message *EventMessage) error {
	h, _ := message.Hash()
	e.log.WithFields(logrus.Fields{
		"topic":    message.Topic,
		"key":      message.Key,
		"data":     message.Data,
		"metadata": message.Metadata,
		"hash":     h,
	}).Info()
	return nil
}

func (e *EventLogger) Delete(ctx context.Context, message *EventMessage) error {
	h, _ := message.Hash()
	e.log.WithFields(logrus.Fields{
		"hash": h,
	}).Info("message succesfully sent")
	return nil
}

func (e *EventLogger) As(i interface{}) bool {
	return false
}
