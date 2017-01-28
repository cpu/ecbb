# ecbb

:robot: :book: Electronic Code Book Bot :book: :robot:

Bzzzt. I eat images and spit them out AES-128-ECB encrypted.

:warning: :zap: **DANGER DANGER** Make sure to [never use ECB mode
yourself](http://crypto.stackexchange.com/questions/20941/why-shouldnt-i-use-ecb-encryption)! **DANGER DANGER** :zap: :warning:

![ECB Garfield](https://github.com/cpu/ecbb/blob/master/data/cc-garf.ecb.png)

## Setup

1. Setup Go
2. `go get github.com/cpu/ecbb`
3. `go get github.com/dghubble/oauth1 github.com/dghubble/go-twitter/twitter`
   (For the twitter bot)
4. `go install github.com/cpu/ecbb/..`

### Convert an image

1. `ecbb -listen localhost:6969`
2. `ecbb-convert -input data/cc-garf.png -output data/cc-garf.ecb.png -key lasagna`
3. Open `data/cc-garf.ecb.png`

### Run a twitter bot

1. Get a Twitter API consumer key and consumer secret.
2. Get an acccess token & access token secret for a specific user (e.g.
   `@ecb_penguin`)
3. Run:
```
ecbb-twitter -botUsername $BOT_USER_NAME_HERE  -consumerKey $CONSUMER_KEY 
   -consumerSecret $CONSUMER_SECRET -accessToken $ACCESS_TOKEN 
   -accessSecret $ACCESS_SECRET
```

## Credit

* `data/cc-garf.png` is licensed [CC-BY](https://creativecommons.org/licenses/by/4.0/) by [`_unicorn_`](https://www.sketchport.com/drawing/5744389380898816/garfield)
* ascii art ["Little robot head"](http://www.asciiworld.com/-Robots,24-.html) provided by [www.asciiworld.com](http://www.asciiworld.com)
