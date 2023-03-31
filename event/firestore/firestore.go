package firestore

import (
	"context"
	"errors"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/bondhan/golib/client"
	"github.com/bondhan/golib/event"
	"github.com/bondhan/golib/util"
	"github.com/mitchellh/mapstructure"
)

type FireSender struct {
	store      *firestore.CollectionRef
	Collection string `json:"collection" mapstructure:"collection"`
	Credential string `json:"credential" mapstructure:"credential"`
	ProjectID  string `json:"project_id" mapstructure:"project_id"`
}

func init() {
	event.RegisterSender("firestore", NewFireSender)
	event.RegisterWriter("firestore", NewFireWriter)
}

func NewFireSender(ctx context.Context, config interface{}) (event.Sender, error) {
	return NewFireOutbox(ctx, config)
}

func NewFireWriter(ctx context.Context, config interface{}) (event.Writer, error) {
	return NewFireOutbox(ctx, config)
}

func NewFireOutbox(ctx context.Context, config interface{}) (*FireSender, error) {

	switch cfg := config.(type) {
	case *FireSender:
		return cfg, nil
	}

	var fs FireSender
	if err := mapstructure.Decode(config, &fs); err != nil {
		return nil, err
	}

	if fs.Credential == "" {
		return nil, errors.New("[event/firestore] missing credential param")
	}

	if fs.Collection == "" {
		return nil, errors.New("[event/firestore] missing collection param")
	}

	if fs.ProjectID == "" {
		return nil, errors.New("[event/firestore] missing project param")
	}

	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", fs.Credential)
	cl := client.FirestoreClient(ctx, fs.ProjectID)
	fs.store = cl.Collection(fs.Collection)
	return &fs, nil
}

func (f *FireSender) Send(ctx context.Context, message *event.EventMessage) error {
	outbox, err := event.OutboxFromMessage(message)
	if err != nil {
		return err
	}

	out := make(map[string]interface{})
	if err := util.DecodeJSON(outbox, out); err != nil {
		return err
	}

	ref := f.store.Doc(outbox.ID)
	_, err = ref.Set(ctx, out)
	return err
}

func (f *FireSender) Delete(ctx context.Context, message *event.EventMessage) error {
	outbox, err := event.OutboxFromMessage(message)
	if err != nil {
		return err
	}
	ref := f.store.Doc(outbox.ID)
	_, err = ref.Delete(ctx)
	return err
}

func (f *FireSender) As(i interface{}) bool {
	p, ok := i.(**firestore.CollectionRef)
	if !ok {
		return false
	}
	*p = f.store
	return true
}
