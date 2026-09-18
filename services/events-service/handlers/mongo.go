package handlers

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStore struct {
	events *mongo.Collection
	meta   *mongo.Collection
}

func NewMongoStore(db *mongo.Database) *MongoStore {
	s := &MongoStore{
		events: db.Collection("events"),
		meta:   db.Collection("meta"),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	_, _ = s.events.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "expire_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0),
		},
		{Keys: bson.D{{Key: "name", Value: 1}, {Key: "received_at", Value: -1}}},
	})
	_, _ = s.meta.UpdateOne(ctx, bson.M{"_id": "ttl"},
		bson.M{"$setOnInsert": bson.M{"seconds": DefaultTTLSeconds}},
		options.Update().SetUpsert(true))
	return s
}

func (s *MongoStore) Insert(ctx context.Context, ev Event) error {
	doc := bson.M{
		"name":        ev.Name,
		"session_id":  ev.SessionID,
		"user_id":     ev.UserID,
		"path":        ev.Path,
		"properties":  ev.Properties,
		"occurred_at": ev.OccurredAt,
		"expire_at":   ev.ExpireAt,
		"received_at": ev.ReceivedAt,
	}
	_, err := s.events.InsertOne(ctx, doc)
	return err
}

func (s *MongoStore) List(ctx context.Context, name string, limit int) ([]Event, error) {
	if limit < 1 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	filter := bson.M{}
	if name != "" {
		filter["name"] = name
	}
	cur, err := s.events.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "received_at", Value: -1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := []Event{}
	for cur.Next(ctx) {
		var raw bson.M
		if err := cur.Decode(&raw); err != nil {
			continue
		}
		out = append(out, eventFromBSON(raw))
	}
	return out, cur.Err()
}

func (s *MongoStore) Count(ctx context.Context) (int64, error) {
	return s.events.CountDocuments(ctx, bson.M{})
}

func (s *MongoStore) GetTTL(ctx context.Context) (int, error) {
	var row struct {
		Seconds int `bson:"seconds"`
	}
	err := s.meta.FindOne(ctx, bson.M{"_id": "ttl"}).Decode(&row)
	if err == mongo.ErrNoDocuments {
		return DefaultTTLSeconds, nil
	}
	if err != nil {
		return DefaultTTLSeconds, err
	}
	return ClampTTL(row.Seconds), nil
}

func (s *MongoStore) SetTTL(ctx context.Context, seconds int) error {
	seconds = ClampTTL(seconds)
	_, err := s.meta.UpdateOne(ctx, bson.M{"_id": "ttl"},
		bson.M{"$set": bson.M{"seconds": seconds}},
		options.Update().SetUpsert(true))
	return err
}

func eventFromBSON(raw bson.M) Event {
	ev := Event{Properties: map[string]interface{}{}}
	if id, ok := raw["_id"].(primitive.ObjectID); ok {
		ev.ID = id.Hex()
	}
	if v, ok := raw["name"].(string); ok {
		ev.Name = v
	}
	if v, ok := raw["session_id"].(string); ok {
		ev.SessionID = v
	}
	switch v := raw["user_id"].(type) {
	case int32:
		ev.UserID = int(v)
	case int64:
		ev.UserID = int(v)
	case int:
		ev.UserID = v
	}
	if v, ok := raw["path"].(string); ok {
		ev.Path = v
	}
	if v, ok := raw["properties"].(bson.M); ok {
		ev.Properties = v
	}
	if v, ok := raw["occurred_at"].(primitive.DateTime); ok {
		ev.OccurredAt = v.Time()
	}
	if v, ok := raw["expire_at"].(primitive.DateTime); ok {
		ev.ExpireAt = v.Time()
	}
	if v, ok := raw["received_at"].(primitive.DateTime); ok {
		ev.ReceivedAt = v.Time()
	}
	return ev
}
