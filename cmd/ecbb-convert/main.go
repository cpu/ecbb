package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/cpu/ecbb/util"
)

// sendImage reads an imageFile and sends it to the ECBB API at the given server
// to be encrypted with the given key. It returns the encrypted image bytes or
// an error
func sendImage(imageFile string, key string, server string) ([]byte, error) {
	imageBytes, err := ioutil.ReadFile(imageFile)
	if err != nil {
		return nil, err
	}

	return util.ECBPostImage(imageBytes, imageFile, key, server)
}

func main() {
	key := flag.String("key", "", "AES-ECB encryption key")
	inputFile := flag.String("input", "data/cc-garf.png", "input file to convert")
	server := flag.String("server", "http://localhost:6969", "ecbb server address")
	outputFile := flag.String("output", "data/cc-garf.ecb.png", "file to save output to")

	flag.Parse()

	if *key == "" {
		util.ErrorQuit("You must specify a non-empty -key for encryption")
	}

	result, err := sendImage(*inputFile, *key, *server)
	if err != nil {
		util.ErrorQuit(err.Error())
	}

	err = ioutil.WriteFile(*outputFile, result, 0644)
	if err != nil {
		util.ErrorQuit(err.Error())
	}
	fmt.Printf("Wrote output to %q\n", *outputFile)
}
