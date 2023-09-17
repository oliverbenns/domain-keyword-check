package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

type Word struct {
	Text     string `json:"text"`
	Location int    `json:"location"`
}

type Data struct {
	Words []Word `json:"words"`
}

func main() {
	ctx := context.Background()
	redisClient, err := createRedisClient(ctx)
	if err != nil {
		log.Panic(err.Error())
	}
	svc := &Service{
		redisClient: redisClient,
	}
	err = svc.run(ctx)
	if err != nil {
		log.Panic(err.Error())
	}
}

func createRedisClient(ctx context.Context) (*redis.Client, error) {
	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		return nil, fmt.Errorf("missing redis url")
	}

	opt, err := redis.ParseURL(redisUrl)
	if err != nil {
		return nil, err
	}

	redisClient := redis.NewClient(opt)

	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return redisClient, nil
}
