package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"
)

func main() {
	fmt.Println("here we go #hahaha")

	// capture ctrl+c and stop execution
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	jobs := make(chan string)

	go func() {
		time.Sleep(time.Second * 1)
		jobs <- "ping"
	}()

	// jvt: run as long as no interrupt is sent
	go func() {
		for sig := range c {
			fmt.Printf("captured %v, notifying everyone...\n", sig)

			fmt.Println("exiting...")
			os.Exit(1)
		}
	}()

	for job := range jobs {
		// jvt: check for new job
		fmt.Println("got new job: " + job)
		go hug(job)
	}
}
