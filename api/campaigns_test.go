package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mitchellh/mapstructure"

	"shortly/api/response"

	"shortly/app/campaigns"
)

type MockCampaignRepository struct {
}

func (repo *MockCampaignRepository) GetUserCampaigns(accountID int64) ([]campaigns.Campaign, error) {
	return []campaigns.Campaign{
		{ID: 1, Name: "campaign_1", Description: "campaign_1 description"},
		{ID: 2, Name: "campaign_2", Description: "campaign_2 description"},
	}, nil
}

func TestGetAccountCampaigns(t *testing.T) {

	logger := log.New(ioutil.Discard, "", log.Lshortfile)

	mockRepo := &MockCampaignRepository{}
	urlhandler := GetUserCampaigns(mockRepo, logger)

	req := httptest.NewRequest("GET", "http://example.com/api/v1/campaigns", nil)
	ctx := context.Background()
	ctx = context.WithValue(ctx, "user", &JWTClaims{AccountID: 1})
	req = req.WithContext(ctx)
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

	var response response.ApiResponse

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

		var e CampaignResponse
		if err := mapstructure.Decode(result, &e); err != nil {
			t.Errorf("result is not CampaignResponse")
		}

		if i == 0 {
			if e.ID != 1 || e.Name != "campaign_1" || e.Description != "campaign_1 description" {
				t.Errorf("test response #1 is not valid")
			}
		} else if i == 1 {
			if e.ID != 2 || e.Name != "campaign_2" || e.Description != "campaign_2 description" {
				t.Errorf("test response #2 is not valid")
			}
		}
	}
}
