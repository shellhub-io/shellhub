package events

import (
	"encoding/json"
	"sync"

	"github.com/hibiken/asynq"
)

type EventsAsynq struct { //nolint:revive
	client *asynq.Client
}

// NewEventsAsynq connects to the events queue and returns a Events struct to push events.
//
// It also set the Events struct as the global events variable to be used by the Push function.
func NewEventsAsynq(uri string) *EventsAsynq {
	mu := new(sync.Mutex)
	mu.Lock()
	events = &EventsAsynq{
		client: asynq.NewClient(asynq.RedisClientOpt{Addr: "redis:6379"}), //nolint:exhaustruct
	}
	mu.Unlock()

	return events.(*EventsAsynq)
}

func (e *EventsAsynq) Push(event Event, data EventPayload) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if _, err := e.client.Enqueue(asynq.NewTask(string(event), payload)); err != nil {
		return err
	}

	return nil
}

func (e *EventsAsynq) Close() error {
	return e.client.Close()
}
