package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"go.uber.org/zap"
)

var (
	CONTENT_JSON = "json"
	CONTENT_FORM = "form"
	CONTENT_TEXT = "text"
)

// Converts interface to []byte
func GetBytes(content interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(content)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Post without automatic failover
func HttpJsonPost(url string, skipTlsVerify bool, content []byte) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: skipTlsVerify},
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(content))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode > 299 {
		return respBody, fmt.Errorf("Unexpected status code received: %d", resp.StatusCode)
	}
	return respBody, nil
}

func NewHTTPPoster(skipTLSVerify bool, replyTimeout time.Duration) *HTTPPoster {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTLSVerify},
	}
	return &HTTPPoster{httpClient: &http.Client{Transport: tr, Timeout: replyTimeout}}
}

type HTTPPoster struct {
	httpClient *http.Client
}

// Post with built-in failover
// Returns also reference towards client so we can close it's connections when done
func (poster *HTTPPoster) Post(addr string, contentType string, content interface{}, attempts int, fallbackFilePath string) ([]byte, error) {
	if !IsSliceMember([]string{CONTENT_JSON, CONTENT_FORM, CONTENT_TEXT}, contentType) {
		return nil, fmt.Errorf("unsupported ContentType: %s", contentType)
	}
	var body []byte        // Used to write in file and send over http
	var urlVals url.Values // Used when posting form
	if IsSliceMember([]string{CONTENT_JSON, CONTENT_TEXT}, contentType) {
		body = content.([]byte)
	} else if contentType == CONTENT_FORM {
		urlVals = content.(url.Values)
		body = []byte(urlVals.Encode())
	}
	delay := Fib()
	bodyType := "application/x-www-form-urlencoded"
	if contentType == CONTENT_JSON {
		bodyType = "application/json"
	}
	var err error
	for i := 0; i < attempts; i++ {
		var resp *http.Response
		if IsSliceMember([]string{CONTENT_JSON, CONTENT_TEXT}, contentType) {
			resp, err = poster.httpClient.Post(addr, bodyType, bytes.NewBuffer(body))
		} else if contentType == CONTENT_FORM {
			resp, err = poster.httpClient.PostForm(addr, urlVals)
		}
		if err != nil {
			Logger.Warn("<HTTPPoster> Posting to : ", zap.String("addr", addr), zap.Error(err))
			time.Sleep(delay())
			continue
		}
		defer resp.Body.Close()
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			Logger.Warn("<HTTPPoster> Posting to : ", zap.String("addr", addr), zap.Error(err))
			time.Sleep(delay())
			continue
		}
		if resp.StatusCode > 299 {
			Logger.Warn("<HTTPPoster> Posting to : ", zap.String("addr", addr), zap.Int("unexpected status code", resp.StatusCode))
			time.Sleep(delay())
			continue
		}
		return respBody, nil
	}
	// If we got that far, post was not possible, write it on disk
	fileOut, err := os.Create(fallbackFilePath)
	if err != nil {
		return nil, err
	}
	defer fileOut.Close()
	if _, err := fileOut.Write(body); err != nil {
		return nil, err
	}
	return nil, nil
}
