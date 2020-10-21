package leanix

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/op/go-logging"
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

var log = logging.MustGetLogger("leanix-k8s-connector")

// Authenticate uses token to authenticate against MTM and response with access_token
func Authenticate(fqdn string, token string) (string, error) {
	body := strings.NewReader("grant_type=client_credentials")
	req, err := http.NewRequest("POST", "https://"+fqdn+"/services/mtm/v1/oauth2/token", body)
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

		log.Fatalf("Failed to authenticate. Check if the provided LeanIX API token is still valid and not expired. status: %d", resp.StatusCode)
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

// Upload uploads the generated LDIF to the Integration API and response with id
func Upload(fqdn string, accessToken string, ldif []byte) (SyncRunResponse, error) {
	body := bytes.NewReader(ldif)
	req, err := http.NewRequest("POST", "https://"+fqdn+"/services/integration-api/v1/synchronizationRuns", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	if err != nil {
		return SyncRunResponse{"", "", ""}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return SyncRunResponse{"", "", ""}, err
	}
	defer resp.Body.Close()
	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return SyncRunResponse{"", "", ""}, err
	}
	syncRunResponse := SyncRunResponse{}
	json.Unmarshal(responseData, &syncRunResponse)
	return syncRunResponse, nil
}

// StartRun starts the Integration API run and response with id
func StartRun(fqdn string, accessToken string, id string) (int, error) {
	req, err := http.NewRequest("POST", "https://"+fqdn+"/services/integration-api/v1/synchronizationRuns/"+id+"/start", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	if err != nil {
		return 0, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return resp.StatusCode, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}
