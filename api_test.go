package main

import (
	"fmt"
	"bytes"
	"log"
	"net/http"
	"testing"
	"io/ioutil"
	"net/http/httptest"
	"encoding/json"

	"github.com/mitchellh/mapstructure"

	"shortly/api"

	"shortly/app/urls"
)

type MockUrlsRepository struct {

}

func (repo *MockUrlsRepository) GetAllUrls() ([]urls.UrlPair, error) {
	return []urls.UrlPair{
		{Short: "12345", Long: "www.google.com"},
		{Short: "ABCDE", Long: "www.twitter.com"},
	}, nil
}

func (repo *MockUrlsRepository) CreateUrl(short, long string) error {
	return nil
}

func (repo *MockUrlsRepository) GetUserUrls(_, _ int64) ([]urls.UrlPair, error) {
	return []urls.UrlPair{
		{Short: "12345", Long: "www.facebook.com"},
		{Short: "ABCDE", Long: "www.netflix.com"},
	}, nil
}

func (repo *MockUrlsRepository) GetUserUrlsCount(userID int64) (int, error) {
	return 2, nil
}

func TestGetUrls(t *testing.T) {

	logger := log.New(ioutil.Discard, "", log.Lshortfile)

	mockRepo := &MockUrlsRepository{}
	urlhandler := api.GetURLList(mockRepo, logger)

	req := httptest.NewRequest("GET", "http://example.com/api/v1/urls", nil)
	w := httptest.NewRecorder()
	urlhandler.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status code != 200")
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type != application/json")
	}

	var response api.ApiResponse

	buff := bytes.NewBuffer(body)
	err := json.NewDecoder(buff).Decode(&response)
	if err != nil {
		t.Errorf("response is not api response")
	}

	urls, ok := response.Result.([]interface{})
	if !ok {
		t.Errorf("response result is not slice")
	}

	fmt.Println(string(body))

	if len(urls) != 2 {
		t.Errorf("response length != 2, %v", len(urls))
	}	

	for i, u := range urls {

		result, ok := u.(map[string]interface{})
		if !ok {
			t.Errorf("result is not map[string]interface{}, but %T", u)
		}

		var e api.UrlResponse
		if err := mapstructure.Decode(result, &e); err != nil {
			t.Errorf("result is not UrlResponse")
		}

		if i == 0 {
			if e.Short != "12345" && e.Long != "www.google.com" {
				t.Errorf("test response #1 is not valid")
			}
		} else if i == 1 {
			if e.Short != "ABCDE" && e.Long != "www.twitter.com" {
				t.Errorf("test response #2 is not valid")
			}
		}
	}
}