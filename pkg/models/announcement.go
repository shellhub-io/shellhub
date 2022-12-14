package models

import (
	"time"
)

type AnnouncementShort struct {
	UUID  string    `json:"uuid" bson:"uuid"`
	Title string    `json:"title" bson:"title"`
	Date  time.Time `json:"date" bson:"date"`
}

type Announcement struct {
	UUID    string    `json:"uuid" bson:"uuid"`
	Title   string    `json:"title" bson:"title"`
	Content string    `json:"content" bson:"content"`
	Date    time.Time `json:"date" bson:"date"`
}
