package config

import (
	"encoding/json"
	"io/ioutil"
)

type Configuration struct {
	Github GithubConfiguration `json:"github"`
}

type GithubConfiguration struct {
	Username   string                    `json:"username"`
	APIToken   string                    `json:"api-token"`
	PRTemplate githubPullRequestTemplate `json:"pull-request"`
}

type githubPullRequestTemplate struct {
	Title string   `json:"title"`
	Body  []string `json:"body"`
}

func NewConfiguration(configFile *string) (*Configuration, error) {
	fileContent, err := ioutil.ReadFile(*configFile)
	if err != nil {
		return nil, err
	}

	var config Configuration
	err = json.Unmarshal(fileContent, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
