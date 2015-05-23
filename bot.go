package main

import (
	"encoding/json"
	"errors"
	"fmt"
	//"io/ioutil"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/garyburd/redigo/redis"
)

type Hug struct {
	TweetID       string
	URL           string
	PullRequestId int
}

func FetchFromQueue(client redis.Conn) (*Hug, error) {
	values, err := redis.Values(client.Do("BLPOP", "hug:queue", 0))
	if err != nil {
		return nil, err
	}

	bytes, err := redis.Bytes(values[1], nil)
	if err != nil {
		log.Print("Something broke in bytes", err)
		return nil, errors.New("Something broke in bytes")
	}

	var hug Hug
	err = json.Unmarshal(bytes, &hug)
	if err != nil {
		return nil, err
	}

	return &hug, nil
}

func AddFinished(client redis.Conn, hug *Hug) error {
	client.Do("RPUSH", "hug:finished")
	jsonHug, err := json.Marshal(hug)
	if err != nil {
		return err
	}

	_, err = client.Do("RPUSH", "hug:finished", string(jsonHug))
	return nil
}

func ConnectRedis(url string, auth string) redis.Conn {
	redisClient, err := redis.Dial("tcp", url)
	if err != nil {
		log.Fatal("Redis client init (connect) failed:", err)
	}

	if len(auth) == 0 {
		return redisClient
	}

	if _, err := redisClient.Do("AUTH", auth); err != nil {
		redisClient.Close()
		log.Fatal("Redis client init (auth) failed:", err)
	}

	return redisClient
}

func main() {
	// jvt: @todo error handling?

	// capture ctrl+c and stop execution
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	jobs := make(chan *Hug)

	redisUrl := os.Getenv("REDIS_URL")
	redisAuth := os.Getenv("REDIS_AUTH")

	go func() {
		time.Sleep(time.Second * 1)
		redisClient := ConnectRedis(redisUrl, redisAuth)
		defer redisClient.Close()

		for {
			hug, err := FetchFromQueue(redisClient)
			fmt.Println(hug, err)
			if err == nil {
				jobs <- hug
			}
		}
	}()

	// jvt: run as long as no interrupt is sent
	go func() {
		for sig := range c {
			fmt.Printf("captured %v, notifying everyone...\n", sig)

			fmt.Println("exiting...")
			os.Exit(1)
		}
	}()

	// jvt: check for new job
	for job := range jobs {
		fmt.Println("got new job: ", job)
		go processHug(job)
	}
}
