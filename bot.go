package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"
	"io/ioutil"
)

func main() {
	/*testFile, _ := ioutil.ReadFile("./README.md.1")
	// jvt: @todo error handling?
	processor, _ := newSpellCheckFileProcessor()
	correctedContent := processor.processContent(testFile);
	fmt.Println(correctedContent)
	os.Exit(1)*/

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

	// jvt: check for new job
	for job := range jobs {
		fmt.Println("got new job: " + job)
		go hug(job)
	}
}
