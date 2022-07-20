package leanix

import (
	"encoding/json"
	"fmt"
	"github.com/leanix/leanix-k8s-connector/pkg/logger"
	"io/ioutil"
	"net/http"
	"strings"
)

// AuthResponse struct
type AuthResponse struct {
	Scope       string `json:"scope"`
	Expired     bool   `json:"expired"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// SyncRunResponse struct
type SyncRunResponse struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

// Authenticate uses token to authenticate against MTM and response with access_token
func Authenticate(fqdn string, token string) (string, error) {
	body := strings.NewReader("grant_type=client_credentials")
	req, err := http.NewRequest("POST", "https://"+fqdn+"/services/mtm/v1/oauth2/token", body)
	logger.Infof("LeanIX integration api fqdn : %s", fqdn)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		return "", err
	}
	req.SetBasicAuth("apitoken", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		err := fmt.Errorf("Integration API authentication failed: %s", resp.Status)
		return "", err
	}
	defer resp.Body.Close()
	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	authResponse := AuthResponse{}
	json.Unmarshal(responseData, &authResponse)
	return authResponse.AccessToken, nil
}
