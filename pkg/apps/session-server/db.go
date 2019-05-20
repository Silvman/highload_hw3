package session_server

import (
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

var db *redis.Client

func InitRedis(redisAddr string) error {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	status := client.Ping()
	if _, err := status.Result(); err != nil {
		return errors.Wrap(err, "can't init redis")
	}
	db = client
	return nil
}

func GetInstanse() *redis.Client {
	return db
}
