package util

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
)

// ErrorQuit prints msg to stderr and then os.Exit's non-zero
func ErrorQuit(msg string) {
	fmt.Fprintf(os.Stderr, "ERROR! ERROR! %s\n", msg)
	os.Exit(420)
}

// GetImage performs an HTTP Get of a target URL, returning the response body
// bytes or an error
func GetImage(targetUrl string) ([]byte, error) {
	resp, err := http.Get(targetUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBuf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.Status != "200 OK" {
		return nil, fmt.Errorf("Non-200 response code: %#v", resp.Status)
	}
	return respBuf, nil
}

/*
 *  PostImage performs an HTTP POST with a multi-part encoded POST body using
 *  the given http Client. It abstracts the process of performing an image
 *  upload by requiring the caller provide image bytes, the name of the image
 *  form field, the image "filename", and any extra form fields to be added. It
 *  returns the response body bytes or an error
 */
func PostImage(image []byte, imageField, imageName string, extra map[string]string, targetUrl string, client *http.Client) ([]byte, error) {
	// Create a buffer for the POST body and a multipart form writer to add
	// content to it
	body := &bytes.Buffer{}
	bufWriter := multipart.NewWriter(body)

	// Add the extra form fields
	for k, v := range extra {
		bufWriter.WriteField(k, v)
	}

	// Add the image form field and filename
	formWriter, err := bufWriter.CreateFormFile(imageField, imageName)
	if err != nil {
		return nil, err
	}

	// Copy the input image bytes to the multipart form field
	inputReader := bytes.NewReader(image)
	_, err = io.Copy(formWriter, inputReader)
	if err != nil {
		return nil, err
	}
	// Save the content type before closing the writer
	contentType := bufWriter.FormDataContentType()
	bufWriter.Close()

	// POST to the target URL with the form data as the POST body
	resp, err := client.Post(targetUrl, contentType, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// Read the response data
	respBuf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.Status != "200 OK" {
		return nil, fmt.Errorf("Non-200 response code: %#v", resp.Status)
	}
	return respBuf, nil
}

// ECBPostImage is a conveneince wrapper around PostImage that uses the
// `http.DefaultClient` to send an image to the ECCB HTTP api
func ECBPostImage(imageBytes []byte, filename, key, server string) ([]byte, error) {
	endpoint := fmt.Sprintf("%s/new", server)
	extraFields := map[string]string{
		"key": key,
	}
	return PostImage(imageBytes, "image", filename, extraFields, endpoint, http.DefaultClient)
}
