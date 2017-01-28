package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/cpu/ecbb/util"
	"github.com/dghubble/go-twitter/twitter"
)

const (
	// TODO(@cpu): Switch to the "modern" media API.
	twitterUploadAPI = "https://upload.twitter.com/1.1/media/upload.json"

	// The filename used for the multi-part form file field (Does Twitter even use
	// this?)
	imageFilename = "ecbb-result-img.png"
)

var (
	mediaURLRemoveRegexp = regexp.MustCompile(`https:\/\/t\.co\/[a-zA-Z0-9]+`)
)

// replyToTweet replies to a specified tweet with the given message, embedding
// the supplied mediaID in the update parameters. The posted tweet is returned,
// or an error
func (b bot) replyToTweet(tweet *twitter.Tweet, msg string, mediaID int64) (*twitter.Tweet, error) {
	status := fmt.Sprintf(".@%s %s", tweet.User.ScreenName, msg)
	updateParams := &twitter.StatusUpdateParams{
		InReplyToStatusID: tweet.ID,
		MediaIds:          []int64{mediaID},
	}
	t, resp, err := b.client.Statuses.Update(status, updateParams)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("expected 200 response, got %d", resp.StatusCode)
	}
	return t, nil
}

// twitterUploadImage uploads an image to the twitter API returning a Media ID or an
// error if unsuccessful.
func (b bot) twitterUploadImage(image []byte) (int64, error) {
	// Upload the image to twitter. The API expects the image in a form field
	// called "media"
	respBuf, err := util.PostImage(image, "media", imageFilename, nil, twitterUploadAPI, b.httpClient)
	if err != nil {
		return 0, err
	}
	// Attempt to unmarshal the Media ID from the response body
	var respOb struct {
		MediaID int64 `json:"media_id"`
	}
	err = json.Unmarshal(respBuf, &respOb)
	if err != nil {
		return 0, err
	}
	return respOb.MediaID, nil
}

// firstPictureEntity searches a *twitter.Entities's Media for the first
// MediaEntity with type "photo". If none are found an error is returned.
func firstPictureEntity(entities *twitter.Entities) (*twitter.MediaEntity, error) {
	if entities == nil {
		return nil, fmt.Errorf("nil entities")
	}
	// Find the first picture in the mention
	for _, m := range entities.Media {
		if m.Type == "photo" {
			return &m, nil
		}
	}
	return nil, fmt.Errorf("no photo media in tweet")
}

// getMentionPictureBytes downloads a photo twitter.MediaEntity by sending
// a HTTP GET to its media URL
func (b bot) getMentionPictureBytes(photo *twitter.MediaEntity) ([]byte, error) {
	// Prefer the MediaURLHttps if possible for fetching the attached media
	mediaURL := photo.MediaURL
	if photo.MediaURLHttps != "" {
		mediaURL = photo.MediaURLHttps
	}
	fmt.Printf("[*] - Fetching image %q\n", mediaURL)
	// Fetch the image's bytes from the media URL
	imgBytes, err := util.GetImage(mediaURL)
	if err != nil {
		return nil, err
	}
	return imgBytes, nil
}

// tweetTextToKey cleans up a tweet's Text for use as an encryption key input.
// Mainly we want to remove things that aren't part of the true message.
func (b bot) tweetTextToKey(text string) string {
	// Remove the first occurrence of "@username" that will always be in
	// the tweet text by virtue of it being a mention of the bot
	key := strings.Replace(text, fmt.Sprintf("@%s ", b.username), "", 1)
	// Replace all of the t.co media links (for the attached picture(s)) so they
	// aren't included in the key, just message text
	key = mediaURLRemoveRegexp.ReplaceAllString(key, "")
	return key
}

// handleMention processes a tweet that mentions the bot. If the mention has
// a image type media attached then the bot will encrypt the image and attach it
// to a reply tweet
func (b bot) handleMention(tweet *twitter.Tweet) {
	start := time.Now()

	// This shouldn't happen, but return early if it does
	if tweet == nil || tweet.Entities == nil || tweet.User == nil {
		return
	}

	from := tweet.User.ScreenName
	// Ignore our own tweets!
	if from == b.username {
		return
	}

	prefix := fmt.Sprintf("@%s ", b.username)
	if !strings.HasPrefix(tweet.Text, prefix) {
		fmt.Printf("[!] Skipping event. Didn't begin with %q\n", prefix)
		return
	}
	fmt.Printf("[*] Handling a mention from %s: %q\n", from, tweet.Text)

	// Find the first picture in the mention
	firstPicture, err := firstPictureEntity(tweet.Entities)
	if err != nil {
		fmt.Printf("[!] Couldn't find first picture in mention: %s\n", tweet.Text)
		return
	}

	// Create a passphrase key string from the tweet text
	key := b.tweetTextToKey(tweet.Text)

	job := replyJob{
		tweet: tweet,
		from:  from,
		photo: firstPicture,
		key:   key,
	}

	// Send the job off
	b.jobs <- job

	end := time.Now()
	duration := end.Sub(start)
	fmt.Printf("[*] Finished processing mention in %s\n", duration)
}

// processReplies blocks on reading a job from the jobs channel. When it has
// one, it calls `b.ProcessReply()` to take care of it. This is all done in the
// same goroutine to allow `processReplies` to be a rate limiter. It only wakes
// up a fixed number of times per hour, limiting the number of tweets we'll
// reply to in a fixed period.
func (b bot) processReplies() {
	for {
		fmt.Printf("[*] Awake and waiting on a job to reply to\n")
		// Block on getting a job from the jobs chan
		job := <-b.jobs
		fmt.Printf("[*] Starting to process reply for %q\n", job.from)
		// Complete the job
		b.processReply(job)
		// Sleep until we're ready to reply again
		fmt.Printf("[*] Sleeping for a bit. See you in %s\n", b.sleepDuration)
		time.Sleep(b.sleepDuration)
	}
}

// processReply takes a replyJob and does the grunt work to complete it. This
// involves downloading the photo bytes, POSTing them to the ECCBot API,
// uploading the returned bytes to twitter, and replying to the tweet with
// a photo attachment. What a hard working function!
func (b bot) processReply(job replyJob) {
	start := time.Now()

	// Download the picture bytes
	imgBytes, err := b.getMentionPictureBytes(job.photo)
	if err != nil {
		fmt.Printf("[!] failed to get mention tweet attached media: %s\n",
			err.Error())
		return
	}

	// Create the ECB encrypted version of the image with the ECBB API
	fmt.Printf("[*] - Sending image to ECBB API\n")
	ecbImgBytes, err := util.ECBPostImage(imgBytes, "twitter-image.png", job.key, b.ecbbServer)
	if err != nil {
		fmt.Printf("[!] - failed to POST to %q : %s\n", b.ecbbServer, err.Error())
		return
	}

	// Upload the ECB encrypted version of the image to twitter to get a media ID
	// for our reply tweet
	mediaID, err := b.twitterUploadImage(ecbImgBytes)
	if err != nil {
		fmt.Printf("[!] - failed to upload reply image to twitter: %s\n", err.Error())
		return
	}
	fmt.Printf("[*] Replying to user %q with media ID %d\n", job.from, mediaID)
	_, err = b.replyToTweet(job.tweet, "OK!", mediaID)
	if err != nil {
		fmt.Printf("[!] - Couldn't reply to tweet: %s", err.Error())
	}

	end := time.Now()
	duration := end.Sub(start)
	fmt.Printf("[*] Finished replying to job in %s\n", duration)
}

// drinkFromFirehose() starts a go routine monitoring the twitter mention stream
// and a go routine for processing replies
func (b bot) drinkFromFirehose() {
	params := &twitter.StreamUserParams{
		StallWarnings: twitter.Bool(true),
		// Only stream events about the user, not events about their followings
		With: "user",
	}
	stream, err := b.client.Streams.User(params)
	if err != nil {
		util.ErrorQuit(
			fmt.Sprintf("Error calling client.Streams.User(params): %s",
				err.Error()))
	}
	// Save the stream reference for later cleanup
	b.stream = stream
	demux := twitter.NewSwitchDemux()
	// When we demux a Tweet, give it to `handleMention`
	demux.Tweet = b.handleMention
	// Start handling the stream with our demuxer in a goroutine
	go demux.HandleChan(stream.Messages)
	// Start processing reply jobs
	go b.processReplies()
}
