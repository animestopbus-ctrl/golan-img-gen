package models

import "time"

type User struct {
	UserID                  int64     `bson:"user_id"`
	Username                string    `bson:"username,omitempty"`
	FirstName               string    `bson:"first_name,omitempty"`
	GenerationCount         int       `bson:"generation_count"`
	LastGenerationTimestamp time.Time `bson:"last_generation_timestamp"`
	RateLimitWindow         time.Time `bson:"rate_limit_window"`
}