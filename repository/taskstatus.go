package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"CodeAutoGo/database"
	"CodeAutoGo/models"

	"go.mongodb.org/mongo-driver/mongo"
)

// 获取集合对象
func getCollection() *mongo.Collection {
	return database.GetCollection("ScanTaskStatus")
}

// 插入用户
func SaveTaskStatus(user models.Task) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := getCollection().InsertOne(ctx, user)
	if err != nil {
		log.Println("插入数据失败:", err)
		return nil, err
	}
	fmt.Println("插入成功,ID:", result.InsertedID)
	return result, nil
}
