package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/mitchellh/mapstructure"

	"shortly/api/response"
	"shortly/config"

	"shortly/app/rbac"
)

type MockRbacRepository struct {
}

func (repo *MockRbacRepository) CreateRole(accountID int64, role rbac.Role) (int64, error) {
	return 1, nil
}

func (repo *MockRbacRepository) UpdateRole(accountID int64, role rbac.Role) error {
	return nil
}

func (repo *MockRbacRepository) GetRole(roleID int64) (rbac.Role, error) {
	return rbac.Role{}, nil
}

func createFakeToken() (string, error) {

	accountID := int64(1)

	claims := &JWTClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: int64(time.Duration(time.Millisecond * 1000 * 3600)),
			Issuer:    fmt.Sprintf("%v", accountID),
		},
		UserID:    accountID,
		Name:      "admin",
		Email:     "test@mail.com",
		Phone:     "123",
		IsStaff:   false,
		AccountID: accountID,
		RoleID:    0,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenSigned, err := token.SignedString([]byte("secret"))
	if err != nil {
		return "", err
	}

	return tokenSigned, nil
}

func TestCreateRole(t *testing.T) {

	logger := log.New(ioutil.Discard, "", log.Lshortfile)

	mockRepo := &MockRbacRepository{}

	permissionRegistry := make(map[string]rbac.Permission)

	authConfig := config.JWTConfig{
		Secret: "secret",
	}

	auth := AuthMiddleware(nil, authConfig, permissionRegistry)

	urlhandler := auth(
		rbac.NewPermission("/api/v1/roles/create", "create_role", "POST"),
		CreateUserRole(mockRepo, logger),
	)

	form := CreateRoleForm{
		Name:        "account_manager",
		Description: "account_manager role",
	}

	buff := bytes.NewBuffer([]byte{})
	if err := json.NewEncoder(buff).Encode(&form); err != nil {
		t.Error("encode error")
	}

	req := httptest.NewRequest("POST", "http://example.com/api/v1/roles/create", buff)

	token, err := createFakeToken()
	if err != nil {
		t.Error("token error")
	}

	req.Header.Set("x-access-token", token)

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

	buff = bytes.NewBuffer(body)
	err = json.NewDecoder(buff).Decode(&response)
	if err != nil {
		t.Errorf("response is not api response")
	}

	var e RoleResponse
	if err := mapstructure.Decode(response.Result, &e); err != nil {
		t.Errorf("result is not RoleResponse")
	}

	if e.ID != 1 {
		t.Errorf("role id != 1")
	}

	if e.Name != form.Name {
		t.Errorf("form name != response name")
	}

	if e.Description != form.Description {
		t.Errorf("form description != response description")
	}

}
