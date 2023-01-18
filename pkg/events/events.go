package events

import "time"

// Event is the type of event.
// It is used to identify the event type.
type Event string

type EventPayload struct {
	// Request is the data sent to whom generated the event.
	Request interface{} `json:"request"`
	// Response is the data returned by whom generated the event.
	Response interface{} `json:"response"`
	// Date is the date of the event.
	Date time.Time `json:"date"`
}

// NewEventPayload creates a new EventPayload struct with the current date.
func NewEventPayload(request interface{}, response interface{}) EventPayload {
	return EventPayload{
		Request:  request,
		Response: response,
		Date:     time.Now(),
	}
}

// A new event should be added here, and follow the pattern:
// resource:action
const (
	// EventSessionCreated is the event type for session creation.
	EventSessionCreated Event = "event:session:created"
	// EventSessionAuthenticated is the event type for session authentication.
	EventSessionAuthenticated Event = "event:session:authenticated"
	// EventSessionFinished is the event type for session deletion.
	EventSessionFinished Event = "event:session:finished"
)

// Events is the interface that wraps the methods required by an event queue.
type Events interface {
	// Push pushes an event to the queue.
	Push(event Event, payload EventPayload) error
	// Close closes the connection to the queue.
	Close() error
}

var events Events

// Push pushes an event to the queue.
//
// It uses the global events variable to push the event.
// The global events variable is set by the NewEvents function.
// If the global events variable is not set, nothing happen.
func Push(event Event, payload EventPayload) error {
	if events != nil {
		return events.Push(event, payload)
	}

	return nil
}
