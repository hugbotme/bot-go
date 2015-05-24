package twitter

import (
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"github.com/hugbotme/bot-go/config"
	"strconv"
)

type Twitter struct {
	API *anaconda.TwitterApi
}

type Hug struct {
	TweetID string
	URL     string
}

func NewClient(config *config.Configuration) *Twitter {
	anaconda.SetConsumerKey(config.Twitter.ConsumerKey)
	anaconda.SetConsumerSecret(config.Twitter.ConsumerSecret)
	api := anaconda.NewTwitterApi(config.Twitter.AccessToken, config.Twitter.AccessTokenSecret)

	client := Twitter{
		API: api,
	}

	return &client
}

func (client *Twitter) GetScreennameAndLink(id_str string) (string, string) {
	id, _ := strconv.ParseInt(id_str, 10, 64)
	tweet, err := client.API.GetTweet(id, nil)
	if err != nil {
		fmt.Println("Twitter failed", err)
		return "", ""
	}

	screenname := tweet.User.ScreenName
	url := "https://twitter.com/" + screenname + "/status/" + id_str

	return screenname, url
}
