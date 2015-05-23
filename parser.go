package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"github.com/hugbotme/bot-go/repository"
	"github.com/hugbotme/bot-go/config"
)

var (
	flagConfigFile *string
	flagPidFile    *string
	flagVersion    *bool
)

const (
	majorVersion = 1
	minorVersion = 0
	patchVersion = 0
)

// Init function to define arguments
func init() {
	flagConfigFile = flag.String("config", "", "Configuration file")
	flagPidFile = flag.String("pidfile", "", "Write the process id into a given file")
	flagVersion = flag.Bool("version", false, "Outputs the version number and exits")
}

func getGithubClient() *repository.GithubClient {

	config, err := getConfig()

	if err != nil {
		log.Fatal("Configuration initialisation failed:", err)
	}

	log.Println(config.Github)

	return repository.NewGithubClient(&config.Github)
}

func getConfig() (*config.Configuration, error)  {

	flag.Parse()

	// Check for configuration file
	if len(*flagConfigFile) <= 0 {
		log.Println("No configuration file found. Please add the --config parameter")
		return nil, errors.New("No configuration file found")
	}

	// PID-File
	if len(*flagPidFile) > 0 {
		ioutil.WriteFile(*flagPidFile, []byte(strconv.Itoa(os.Getpid())), 0644)
	}

	// Bootstrap configuration file
	return config.NewConfiguration(flagConfigFile)
}
