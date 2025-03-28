package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var db *mongo.Database

func ConnectDB(mongoURI, dbName, username, password string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 设置认证信息
	credential := options.Credential{
		Username: username,
		Password: password,
	}

	// 连接 MongoDB
	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI).SetAuth(credential))
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// 测试连接
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("无法连接到 MongoDB:", err)
	}

	// 选择数据库
	db = client.Database(dbName)

	fmt.Println("MongoDB 连接成功")
}

// 获取数据库对象
func GetCollection(collectionName string) *mongo.Collection {
	return db.Collection(collectionName)
}

// 断开连接
func DisconnectDB() {
	if client != nil {
		client.Disconnect(context.TODO())
	}
}
