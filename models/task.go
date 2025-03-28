package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Task struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"` // MongoDB 会自动生成 ObjectID
	ProjectName string             `bson:"project"`
	Status      string             `bson:"status"`
	Content     string             `bson:"content"`
}
