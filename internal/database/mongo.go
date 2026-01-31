package database

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/image-generator-bot/internal/models"
	"github.com/yourusername/image-generator-bot/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	dbName            = "imagebot"
	usersCollection   = "users"
	historyCollection = "history"
	rateLimitMax      = 5
	rateLimitWindow   = time.Hour
)

// NewMongoClient creates a new MongoDB client
func NewMongoClient(uri string) (*mongo.Client, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	return client, nil
}

// CheckRateLimit checks if user can generate
func CheckRateLimit(ctx context.Context, client *mongo.Client, userID int64) (bool, error) {
	coll := client.Database(dbName).Collection(usersCollection)

	var user models.User
	err := coll.FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		// New user
		user = models.User{
			UserID:                userID,
			GenerationCount:       0,
			LastGenerationTimestamp: time.Time{},
			RateLimitWindow:       time.Now().Add(rateLimitWindow),
		}
		_, err = coll.InsertOne(ctx, user)
		return true, err
	} else if err != nil {
		return false, err
	}

	now := time.Now()
	if now.After(user.RateLimitWindow) {
		// Reset window
		user.GenerationCount = 0
		user.RateLimitWindow = now.Add(rateLimitWindow)
	}

	if user.GenerationCount >= rateLimitMax {
		return false, nil
	}

	return true, nil
}

// UpdateUserAfterGeneration updates count and history
func UpdateUserAfterGeneration(ctx context.Context, client *mongo.Client, user *models.User, prompt string, ts time.Time) error {
	usersColl := client.Database(dbName).Collection(usersCollection)
	historyColl := client.Database(dbName).Collection(historyCollection)

	// Update user
	filter := bson.M{"user_id": user.UserID}
	update := bson.M{
		"$inc": bson.M{"generation_count": 1},
		"$set": bson.M{
			"last_generation_timestamp": ts,
			"rate_limit_window":         time.Now().Add(rateLimitWindow), // Refresh window on generation?
		},
	}
	_, err := usersColl.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	// Add to history
	history := models.History{
		UserID:    user.UserID,
		Prompt:    prompt,
		Timestamp: ts,
		// Seed: optional, not implemented here
	}
	_, err = historyColl.InsertOne(ctx, history)
	if err != nil {
		return err
	}

	// Keep last 10: Delete old ones
	cursor, err := historyColl.Find(ctx, bson.M{"user_id": user.UserID}, options.Find().SetSort(bson.M{"timestamp": -1}).SetSkip(10))
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var ids []interface{}
	for cursor.Next(ctx) {
		var h models.History
		if err := cursor.Decode(&h); err != nil {
			return err
		}
		ids = append(ids, h.ID)
	}
	if len(ids) > 0 {
		_, err = historyColl.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": ids}})
	}

	utils.LogInfo(fmt.Sprintf("Updated DB for user %d", user.UserID))
	return err
}

// GetUserHistory gets last N prompts
func GetUserHistory(ctx context.Context, client *mongo.Client, userID int64, limit int) ([]models.History, error) {
	coll := client.Database(dbName).Collection(historyCollection)
	opts := options.Find().SetSort(bson.M{"timestamp": -1}).SetLimit(int64(limit))
	cursor, err := coll.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var histories []models.History
	if err = cursor.All(ctx, &histories); err != nil {
		return nil, err
	}
	return histories, nil
}