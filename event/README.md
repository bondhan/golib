# event
# Event Emitter

Supported driver
- Logger (sender & writer)
- Pubsub (sender)
- MongoDB outbox (sender & writer)

Support for hybrid mode, combination of sender and writer

Hybrid mode can be used to combine outbox pattern with direct publish more efficiently. For exmaple, with the combination of SQL writer and Kafka sender, when `Publish` is called, it will try to write the event in SQL database first, and then asynchronously trying to send the event to kafka topic after that. If the publishing is success, then the SQL writer will delete the record. In an exception case, when the sender is failed to send the event, another service scould be used to query the database and send the event to kafka.

## Usage

API

```
Publish(ctx context.Context, event, key string, message interface{}, metadata map[string]interface{}) error
```


### Event Config

Event config can be used to map internal event name into actual topic name and to add or replace default metadata

```go
package main

import (
	"github.com/bondhan/golib/event"
)

func main() {
	conf := &EmitterConfig{
		Sender: &DriverConfig{
			Type: "logger",
		},
		EventConfig : &EventConfig{
			EventMap : map[string]string{
				"test" : "kafka-test-topic", // map event test into topic kakfa-test-topic
			},
			Metadata : map[string]map[string]interface{} {
				"test" : { // add metdata on event test
					"schema": "test-schema",	
				},
			}
		}
	}

    ctx := context.Background()

	em, err := event.New(ctx, conf)
    em.Publish(ctx, "test", "t123", "testdata", nil)
}
```


### Example (Logger)

```go
package main

import (
	"github.com/bondhan/golib/event"
)

func main() {
	conf := &EmitterConfig{
		Sender: &DriverConfig{
			Type: "logger",
		},
	}

    ctx := context.Background()

	em, err := event.New(ctx, conf)
    em.Publish(ctx, "test", "t123", "testdata", nil)
}
```

### Example (Kafka Sender)

```go
package main

import (
	"github.com/bondhan/golib/event"
    _ "github.com/bondhan/golib/event/pubsub"
)

func main() {
	conf := &event.EmitterConfig{
		Sender: &event.DriverConfig{
			Type: "kafka",
			Config: map[string]interface{}{
				"schema" : "kafka://",
				"kafka_brokers": []string{"localhost:9092"},
			},
		},
	}

    ctx := context.Background()

	em, err := event.New(ctx, conf)
    em.Publish(ctx, "test", "t123", "testdata", nil)
}
```