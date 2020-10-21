package models

import "time"

type License struct {
	RawData   []byte
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}
