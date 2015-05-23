package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"
	//"flag"
)

func main() {
	// jvt: check for test string
	/*var testString string
	var testRepo string
	flag.StringVar(&testString, "t", "", "string to run a test translation on")
	flag.StringVar(&testRepo, "r", "", "repo URL to crawl")
	flag.Parse()

	if len(testString) > 0 {
		fmt.Println("got test string...")
		fmt.Println(spellCheck(testString))
		os.Exit(1)
	}

	if len(testRepo) > 0 {
		fmt.Println("got test repo...")
		hug(testRepo)
		os.Exit(1)
	}*/

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
