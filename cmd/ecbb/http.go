package main

import (
	"fmt"
	"image/png"
	"net/http"
	"os"
	"time"
)

// logError spits out a message to STDERR
func logError(msg string, code int) {
	fmt.Fprintf(os.Stderr, "[!] - %d - %s\n", code, msg)
}

// logSuccess spits out a message to STDOUT
func logSuccess(msg string) {
	fmt.Printf("[*] - 200 - %s\n", msg)
}

// newECB is an HTTP handler that processes a multi-part form submission and
// returns an ECB encrypted image
func newECB(w http.ResponseWriter, r *http.Request) {
	reqStart := time.Now()

	if r.Method != "POST" {
		logError(fmt.Sprintf("Unsupported HTTP method %q", r.Method), http.StatusMethodNotAllowed)
		http.Error(w, "Unsupported HTTP method - use POST", http.StatusMethodNotAllowed)
		return
	}

	// TODO(@cpu): Set a sane & configurable limit to the form size
	r.ParseMultipartForm(32 << 20)

	key := r.FormValue("key")
	if key == "" {
		// TODO(@cpu): read default key from param/config
		key = "<3 - @ecb_penguin"
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		logError(
			fmt.Sprintf("Error calling FormFile: %s", err.Error()),
			http.StatusInternalServerError)
		http.Error(w, "bad \"image\"", http.StatusBadRequest)
		return
	}
	defer file.Close()

	img, err := parseReaderToImage(file)
	if err != nil {
		logError(
			fmt.Sprintf("Error calling parseReaderToImage: %s", err.Error()),
			http.StatusInternalServerError)
		http.Error(w, "bad \"image\"", http.StatusInternalServerError)
		return
	}

	rgba := toRGBA(*img)
	ecbImage, err := ecbEncrypt(*rgba, key)
	if err != nil {
		logError(
			fmt.Sprintf("Error calling ecbEncrypt: %s", err.Error()),
			http.StatusInternalServerError)
		http.Error(w, "An internal server error has occurred",
			http.StatusInternalServerError)
		return
	}

	err = png.Encode(w, ecbImage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	duration := time.Since(reqStart)
	logSuccess(fmt.Sprintf("Processed ECB image with key %q in %s", key, duration))
}
