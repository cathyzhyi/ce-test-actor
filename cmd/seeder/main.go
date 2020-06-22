package main

import (
	"context"
	"math/rand"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cecontext "github.com/cloudevents/sdk-go/v2/context"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
)

type config struct {
	Target      string        `envconfig:"TARGET"`
	Interval    time.Duration `envconfig:"INTERVAL"`
	Concurrency int           `envconfig:"CONCURRENCY" default:"1"`
	Extensions  string        `envconfig:"EXTENSIONS"`
	Size        int64         `envconfig:"SIZE" default:"0"`
}

func main() {
	var env config
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("Failed to process env config: %v", err)
	}

	client, err := cloudevents.NewDefaultClient()
	if err != nil {
		log.Fatalf("Failed to create cloudevents client: %v", err)
	}

	ext := map[string]string{}
	kvs := strings.Split(env.Extensions, ";")
	for _, kv := range kvs {
		p := strings.Split(kv, ":")
		if len(p) == 2 {
			ext[p[0]] = p[1]
		}
	}

	data := make([]byte, env.Size)
	rand.Read(data)

	for {
		e := event.New()
		e.SetID(uuid.New().String())
		e.SetSource("yolocs.ce-test-actor.seeder")
		e.SetType("seed")
		e.SetSubject("tick")
		e.SetTime(time.Now())
		if env.Size > 0 {
			e.SetData("application/octet-stream", data)
		}

		for k, v := range ext {
			e.SetExtension(k, v)
		}

		for i := 0; i < env.Concurrency; i++ {
			go func() {
				resp, ret := client.Request(cecontext.WithTarget(context.Background(), env.Target), e)
				if protocol.IsACK(ret) {
					log.Infof("Successfully seeded event (id=%s) to target %q", e.ID(), env.Target)
					if resp != nil {
						log.Infof("Event replied: %v", *resp)
					}
				} else {
					log.Errorf("Failed to seed event (id=%s) to target %q: %v", e.ID(), env.Target, ret.Error())
				}
			}()
		}

		log.Infof("Sleeping...")
		time.Sleep(env.Interval)
	}
}
