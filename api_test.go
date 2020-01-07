package main

import (
	"bytes"
	"log"
	"net/http"
	"testing"
	"io/ioutil"
	"net/http/httptest"
	"encoding/json"

	"github.com/mitchellh/mapstructure"

	"shortly/api"

	"shortly/app/links"
)

type MockLinksRepository struct {

}

func (repo *MockLinksRepository) GetAllLinks() ([]links.Link, error) {
	return []links.Link{
		{Short: "12345", Long: "www.google.com"},
		{Short: "ABCDE", Long: "www.twitter.com"},
	}, nil
}

func (repo *MockLinksRepository) GenerateLink() string {
	return "ABCDE"
}

func (repo *MockLinksRepository) CreateLink(*links.Link) error {
	return nil
}

func (repo *MockLinksRepository) GetUserLinks(_, _ int64, filters ...links.LinkFilter) ([]links.Link, error) {
	return []links.Link{
		{Short: "12345", Long: "www.facebook.com"},
		{Short: "ABCDE", Long: "www.netflix.com"},
	}, nil
}

func (repo *MockLinksRepository) GetUserLinksCount(accountID int64) (int, error) {
	return 2, nil
}

func (repo *MockLinksRepository) CreateUserLink(accountID int64, _ *links.Link) (int64, error) {
	return 0, nil
}

func (repo *MockLinksRepository) DeleteUserLink(accountID int64, shortURL string) (int64, error) {
	return 0, nil
}


func TestGetLinks(t *testing.T) {

	logger := log.New(ioutil.Discard, "", log.Lshortfile)

	mockRepo := &MockLinksRepository{}
	urlhandler := api.GetURLList(mockRepo, logger)

	req := httptest.NewRequest("GET", "http://example.com/api/v1/links", nil)
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

	rows, ok := response.Result.([]interface{})
	if !ok {
		t.Errorf("response result is not slice")
	}

	if len(rows) != 2 {
		t.Errorf("response length != 2, %v", len(rows))
	}	

	for i, u := range rows {

		result, ok := u.(map[string]interface{})
		if !ok {
			t.Errorf("result is not map[string]interface{}, but %T", u)
		}

		var e api.LinkResponse
		if err := mapstructure.Decode(result, &e); err != nil {
			t.Errorf("result is not LinkResponse")
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