package iris

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type API interface {
	GetConfiguration(configurationName string, accessToken string) ([]byte, error)
	PostResults(results []byte, accessToken string) error
	PostStatus(status []byte, accessToken string) error
}

type api struct {
	client *http.Client
	kind   string
	uri    string
}

func NewApi(client *http.Client, kind string, uri string) API {
	protocol := ""
	if !strings.Contains(uri, "http") {
		protocol = "https://"
	}
	return &api{
		client: client,
		kind:   kind,
		uri:    fmt.Sprintf("%s%s", protocol, uri),
	}
}

func (a *api) GetConfiguration(configurationName string, accessToken string) ([]byte, error) {
	configUrl := fmt.Sprintf("%s/services/vsm-iris/v1/configurations/kubernetes/%s", a.uri, configurationName)
	req, err := http.NewRequest("GET", configUrl, nil)
	if err != nil {
		log.Errorf("Error while creating request to retrieve configuration from %s: %v", configurationName, err)
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Execute request
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response
	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return responseData, nil
}

func (a *api) PostResults(results []byte, accessToken string) error {
	resultUrl := fmt.Sprintf("%s/services/vsm-iris/v1/results", a.uri)
	postReq, err := http.NewRequest("POST", resultUrl, nil)
	if err != nil {
		log.Errorf("Error creating request to post results results: %v", err)
		return err
	}
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("Authorization", "Bearer "+accessToken)
	postReq.Body = ioutil.NopCloser(bytes.NewBuffer(results))

	// Execute request
	resp, err := a.client.Do(postReq)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		responseData, readErr := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if readErr != nil {
			return readErr
		}
		err := fmt.Errorf("posting results status[%s]could not be processed: '%s'", resp.Status, responseData)
		return err
	}
	log.Infof("Event posted successfully [%s]", resp.Status)
	return nil
}

func (a *api) PostStatus(status []byte, accessToken string) error {
	resultUrl := fmt.Sprintf("https://%s/services/vsm-iris/v1/status", a.uri)
	postReq, err := http.NewRequest("POST", resultUrl, nil)
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("Authorization", "Bearer "+accessToken)
	postReq.Body = ioutil.NopCloser(bytes.NewBuffer(status))
	if err != nil {
		log.Errorf("Error while posting status: %v", err)
		return err
	}
	resp, err := http.DefaultClient.Do(postReq)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		responseData, readErr := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if readErr != nil {
			return readErr
		}
		err := fmt.Errorf("posting results status[%s]could not be processed: '%s'", resp.Status, responseData)
		return err
	}
	log.Infof("Status Event posted successfully [%s]", resp.Status)
	return nil
}
