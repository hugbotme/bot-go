package main

import (
	"fmt"
	//"github.com/hugbotme/go-aspell"
	"os"
	"os/signal"
)

func main() {
	fmt.Println("here we go #hahaha")

	// capture ctrl+c and stop execution
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	for {
		// jvt: run as long as no interrupt is sent
		go func() {
			for sig := range c {
				fmt.Printf("captured %v, notifying everyone...\n", sig)

				fmt.Println("exiting...")
				os.Exit(1)
			}
		}()

		// jvt: do things...
	}
}
