package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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
	bytes, err := redis.Bytes(client.Do("BLPOP", "hug:queue"))
	if err != nil {
		return nil, err
	}
	if len(bytes) == 0 {
		return nil, errors.New("No job")
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

func main() {
	testFile, _ := ioutil.ReadFile("./README.md.1")
	// jvt: @todo error handling?
	processor, _ := newSpellCheckFileProcessor()
	fmt.Println(processor.processContent(string(testFile)))
	os.Exit(1)

	// capture ctrl+c and stop execution
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	jobs := make(chan *Hug)

	go func() {
		time.Sleep(time.Second * 1)
		redisClient, err := redis.Dial("tcp", ":6379")
		if err != nil {
			log.Fatal("Redis client init failed:", err)
		}
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
