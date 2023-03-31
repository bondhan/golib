package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/bondhan/golib/client"
	"github.com/bondhan/golib/event"
	"github.com/bondhan/golib/util"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoSender struct {
	Collection string      `json:"collection" mapstructure:"collection"`
	Connection interface{} `json:"connection" mapstructure:"connection"`
	store      *mongo.Collection
}

type MongoOutbox struct {
	ID        string    `bson:"_id,omitempty"`
	Topic     string    `bson:"topic,omitempty"`
	Key       string    `bson:"key,omitempty"`
	Value     string    `bson:"value,omitempty"`
	CreatedAt time.Time `bson:"created_at,omitempty"`
}

func FromOutbox(out *event.OutboxRecord) *MongoOutbox {
	return &MongoOutbox{
		ID:        out.ID,
		Topic:     out.Topic,
		Key:       out.Key,
		Value:     out.Value,
		CreatedAt: out.CreatedAt,
	}
}

func init() {
	event.RegisterSender("mongo", NewMongoSender)
	event.RegisterWriter("mongo", NewMongoWriter)
}

func NewMongoSender(ctx context.Context, config interface{}) (event.Sender, error) {
	return NewMongoOutbox(ctx, config)
}

func NewMongoWriter(ctx context.Context, config interface{}) (event.Writer, error) {
	return NewMongoOutbox(ctx, config)
}

func NewMongoOutbox(ctx context.Context, config interface{}) (*MongoSender, error) {
	var ms MongoSender
	if err := mapstructure.Decode(config, &ms); err != nil {
		return nil, err
	}

	if ms.Connection == nil {
		return nil, errors.New("[event/mongo] missing connection param")
	}

	if ms.Collection == "" {
		return nil, errors.New("[event/mongo] missing collection param")
	}

	switch con := ms.Connection.(type) {
	case *mongo.Database:
		ms.store = con.Collection(ms.Collection)
		return &ms, nil
	case *client.MongoClient:
		cl, err := con.MongoConnect()
		if err != nil {
			return nil, err
		}
		db := cl.Database(con.AppName)
		ms.store = db.Collection(ms.Collection)
		return &ms, nil
	case map[string]interface{}:
		var conf client.MongoClient
		if err := util.DecodeJSON(con, &conf); err != nil {
			return nil, err
		}
		cl, err := conf.MongoConnect()
		if err != nil {
			return nil, err
		}
		db := cl.Database(conf.AppName)
		ms.store = db.Collection(ms.Collection)
		return &ms, nil
	default:
		return nil, errors.New("[event/mongo] unsupported connection type")
	}
}

func (m *MongoSender) Send(ctx context.Context, message *event.EventMessage) error {

	outbox, err := event.OutboxFromMessage(message)
	if err != nil {
		return err
	}

	_, err = m.store.InsertOne(ctx, FromOutbox(outbox))
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil
		}
		return err
	}

	return nil
}

func (m *MongoSender) Delete(ctx context.Context, message *event.EventMessage) error {
	outbox, err := event.OutboxFromMessage(message)
	if err != nil {
		return err
	}

	_, err = m.store.DeleteOne(ctx, bson.D{primitive.E{Key: "_id", Value: outbox.ID}})
	return err
}

func (m *MongoSender) As(i interface{}) bool {
	p, ok := i.(**mongo.Collection)
	if !ok {
		return false
	}
	*p = m.store
	return true
}
