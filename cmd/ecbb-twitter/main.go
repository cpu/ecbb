package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cpu/ecbb/util"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

const (
	// Sleeping 5 minutes between every reply means we can post a maximum of 12
	// tweets per hour.
	sleepDuration = time.Minute * 5
	// If we can tweet a maximum of 12 tweets per hour then we should probably not
	// accept more than 24 hours worth of tweets into the backlog without blocking.
	// If the number of unprocessed replies reaches this limit we will block and
	// likely lose work. Oh well, we're just a silly bot!
	maximumBacklog = 288
)

// bot structs package up all the things a bot needs to be happy
type bot struct {
	httpClient    *http.Client
	client        *twitter.Client
	username      string
	ecbbServer    string
	stream        *twitter.Stream
	jobs          chan replyJob
	sleepDuration time.Duration
}

// replyJob structs encapsulate an item of work for the bot to do in order to
// produce a reply to someone
type replyJob struct {
	tweet *twitter.Tweet
	from  string
	photo *twitter.MediaEntity
	key   string
}

func main() {
	// TODO(@cpu): Awful lot of config flags... Add support for a config file
	consumerPubKey := flag.String("consumerKey", "", "Twitter Consumer Public Key")
	consumerSecKey := flag.String("consumerSecret", "", "Twitter Consumer Secret Key")
	accessPubKey := flag.String("accessToken", "", "Twitter User Access Token")
	accessSecKey := flag.String("accessSecret", "", "Twitter User Access Secret Key")
	botName := flag.String("botUsername", "", "Twitter Username for Access Token/Bot Acct")
	ecbbServer := flag.String("ecbbServer", "http://localhost:6969", "ecbb server address")
	flag.Parse()

	if *consumerPubKey == "" || *consumerSecKey == "" {
		util.ErrorQuit(fmt.Sprintf("You must provide a -consumerKey and -consumerSecret"))
	}
	if *accessPubKey == "" || *accessSecKey == "" {
		util.ErrorQuit(fmt.Sprintf("You must provide a -accessToken and -accessSecret"))
	}
	// TODO(@cpu): There's probably a way to find this with the access token
	// instead of burdening the user with Yet Another Config Flag
	if *botName == "" {
		util.ErrorQuit(fmt.Sprintf("You must provide a -botUsername"))
	}

	// Construct an authenticating httpClient for the consumer & access token
	// pairing, then use it for a new twitter API client
	config := oauth1.NewConfig(*consumerPubKey, *consumerSecKey)
	token := oauth1.NewToken(*accessPubKey, *accessSecKey)
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	// Create a bot to wrap everything up into into one coherent object
	b := bot{
		httpClient:    httpClient,
		client:        client,
		username:      *botName,
		ecbbServer:    *ecbbServer,
		jobs:          make(chan replyJob, maximumBacklog),
		sleepDuration: sleepDuration,
	}

	fmt.Printf("[*] Drinking from the twitter firehose...\n")
	// Start monitoring twitter for mentions in a go routine
	b.drinkFromFirehose()

	// Create a channel for SIGINT and SIGTERM signals
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	// Block on reading from this channel to keep the main goroutine alive until
	// signaled. A bot's work is never done...
	fmt.Println(<-ch)
	fmt.Printf("[!] Quitting!\n")
}
