package client

import (
	"context"
	"os"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	otelmongo "go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
)

var mongoClients = make(map[string]*mongo.Client)

type MongoClient struct {
	URI              string `json:"uri" mapstructure:"uri"`
	AppName          string `json:"name" mapstructure:"name"`
	MinPool          int    `json:"min_pool" mapstructure:"min_pool"`
	MaxPool          int    `json:"max_pool" mapstructure:"max_pool"`
	ReadSecondary    bool   `json:"read_secondary" mapstructure:"read_secondary"`
	ReadMaxStaleness string `json:"read_max_staleness" mapstructure:"read_max_staleness"`
	// on second
	ConnectTimeout time.Duration
	// on second
	PingTimeout time.Duration
}

var (
	defaultConnectTimeout = 10 * time.Second
	defaultPingTimeout    = 2 * time.Second
)

func NewMongoClient(uri, name string) (*mongo.Client, error) {
	return (&MongoClient{
		URI:     uri,
		AppName: name,
	}).MongoConnect()
}

func (c *MongoClient) MongoConnect() (mc *mongo.Client, err error) {

	if cl, ok := mongoClients[c.URI]; ok {
		return cl, nil
	}

	if c.ConnectTimeout == 0 {
		c.ConnectTimeout = defaultConnectTimeout
	}

	if c.PingTimeout == 0 {
		c.PingTimeout = defaultPingTimeout
	}

	if c.AppName == "" {
		c.AppName = "Default"
	}

	connectCtx, cancelConnectCtx := context.WithTimeout(context.Background(), c.ConnectTimeout)
	defer cancelConnectCtx()

	otelMon := otelmongo.NewMonitor()
	opts := []*options.ClientOptions{
		options.Client().SetConnectTimeout(c.ConnectTimeout).ApplyURI(c.URI).SetAppName(c.AppName),
		options.Client().SetMonitor(otelMon),
	}

	if c.MinPool > 0 {
		opts = append(opts, options.Client().SetMinPoolSize(uint64(c.MinPool)))
	}

	if c.MaxPool > 0 {
		opts = append(opts, options.Client().SetMaxPoolSize(uint64(c.MaxPool)))
	}

	if c.ReadSecondary {
		ropt := make([]readpref.Option, 0)
		if c.ReadMaxStaleness != "" {
			d, err := time.ParseDuration(c.ReadMaxStaleness)
			if err == nil {
				ropt = append(ropt, readpref.WithMaxStaleness(d))
			}
		}

		opts = append(opts, options.Client().SetReadPreference(readpref.SecondaryPreferred(ropt...)))
	}

	mc, err = mongo.Connect(connectCtx, opts...)
	if err != nil {
		err = errors.Wrap(err, "failed to create mongodb client")
		return
	}

	pingCtx, cancelPingCtx := context.WithTimeout(context.Background(), c.PingTimeout)
	defer cancelPingCtx()

	if err = mc.Ping(pingCtx, readpref.Primary()); err != nil {
		err = errors.Wrap(err, "failed to establish connection to mongodb server")
	}

	mongoClients[c.URI] = mc
	return
}

func GetMongoClient(url string) *mongo.Client {
	if url == "" {
		url = os.Getenv("MONGO_SERVER_URL")
	}

	if url == "" {
		return nil
	}

	c, ok := mongoClients[url]
	if ok {
		return c
	}

	name := os.Getenv("NAME")
	if name == "" {
		name = "Default"
	}

	cfg := &MongoClient{URI: url, AppName: name}
	client, err := cfg.MongoConnect()
	if err != nil {
		return nil
	}

	return client
}
