package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type History struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    int64              `bson:"user_id"`
	Prompt    string             `bson:"prompt"`
	Timestamp time.Time          `bson:"timestamp"`
	Seed      int64              `bson:"seed,omitempty"` // Optional
}