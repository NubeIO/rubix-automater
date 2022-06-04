package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
)

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

var ctx = context.Background()

var redisClient = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})

func main() {
	subscriber := redisClient.Subscribe(ctx, "test")

	user := User{}

	for {
		msg, err := subscriber.ReceiveMessage(ctx)
		if err != nil {
			panic(err)
		}

		fmt.Println("Received message from "+msg.Channel+" channel.", "data", msg.String())
		if err := json.Unmarshal([]byte(msg.Payload), &user); err != nil {
			//panic(err)
		} else {
			fmt.Println("Received message from " + msg.Channel + " channel.")
			fmt.Printf("%+v\n", user)
		}

	}
}
