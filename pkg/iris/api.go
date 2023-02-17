package iris

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/leanix/leanix-k8s-connector/pkg/iris/models"
	"github.com/leanix/leanix-k8s-connector/pkg/logger"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type API interface {
	GetConfiguration(configurationName string) ([]byte, error)
	GetScanResults(configurationId string) ([]models.IrisResultItem, error)
	PostResults(results []byte) error
	PostEcstResults(ecstResults []byte) error
	PostStatus(status []byte) error
}

type api struct {
	client *http.Client
	kind   string
	uri    string
	token  string
}

func NewApi(client *http.Client, kind string, uri string, token string) API {
	protocol := ""
	if !strings.Contains(uri, "http") {
		protocol = "https://"
	}
	return &api{
		client: client,
		kind:   kind,
		uri:    fmt.Sprintf("%s%s", protocol, uri),
		token:  token,
	}
}

func (a *api) GetConfiguration(configurationName string) ([]byte, error) {
	if configurationName == "" {
		return nil, errors.New("configuration name should not be null or empty")
	}
	configUrl := fmt.Sprintf("%s/services/vsm-iris/v1/configurations/kubernetes/%s", a.uri, configurationName)
	req, err := http.NewRequest("GET", configUrl, nil)
	if err != nil {
		logger.Errorf("Error while creating request to retrieve configuration from %s: %v", configurationName, err)
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+a.token)

	// Execute request
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		logger.Errorf("Could not find configuration in Iris with error: '%s'", responseData)
		return nil, fmt.Errorf("failed to retrieve configuration with name '%s' from Iris", configurationName)

	}
	return responseData, nil
}

func (a *api) GetScanResults(configurationId string) ([]models.IrisResultItem, error) {
	if configurationId == "" {
		return nil, errors.New("configuration name should not be null or empty")
	}
	configUrl := fmt.Sprintf("%s/services/vsm-iris/v1/configuration/%s/results", a.uri, configurationId)
	req, err := http.NewRequest("GET", configUrl, nil)
	if err != nil {
		logger.Errorf("Error while creating request to retrieve latestResults from config '%s': %v", configurationId, err)
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+a.token)

	// Execute request
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		logger.Errorf("Could not find configuration in Iris with error: '%s'", responseData)
		return nil, fmt.Errorf("failed to retrieve configuration with name '%s' from Iris", configurationId)

	}
	previousResults := make([]models.IrisResultItem, 0)
	err = json.Unmarshal(responseData, &previousResults)
	if err != nil {
		return nil, err
	}

	return previousResults, nil
}

// OLD Results endpoint
func (a *api) PostResults(results []byte) error {
	resultUrl := fmt.Sprintf("%s/services/vsm-iris/v1/results", a.uri)
	postReq, err := http.NewRequest("POST", resultUrl, nil)
	if err != nil {
		logger.Errorf("Error creating request to post results results: %v", err)
		return err
	}
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("Authorization", "Bearer "+a.token)
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
		err := fmt.Errorf("posting results status [%s] could not be processed: '%s'", resp.Status, responseData)
		return err
	}
	logger.Infof("Discovery Event posted successfully [%s]", resp.Status)
	return nil
}

// Send request to ECST Endpoint
func (a *api) PostEcstResults(ecstResults []byte) error {
	resultUrl := fmt.Sprintf("%s/services/vsm-iris/v1/results/ecst", a.uri)
	postReq, err := http.NewRequest("POST", resultUrl, nil)
	if err != nil {
		logger.Errorf("Error creating request to post ECST results: %v", err)
		return err
	}
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("Authorization", "Bearer "+a.token)
	postReq.Body = ioutil.NopCloser(bytes.NewBuffer(ecstResults))

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
		err := fmt.Errorf("posting ECST results status [%s] could not be processed: '%s'", resp.Status, responseData)
		return err
	}
	logger.Infof("ECST Discovery Event posted successfully [%s]", resp.Status)
	return nil
}

func (a *api) PostStatus(status []byte) error {
	resultUrl := fmt.Sprintf("%s/services/vsm-iris/v1/status", a.uri)
	postReq, err := http.NewRequest("POST", resultUrl, nil)
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("Authorization", "Bearer "+a.token)
	postReq.Body = ioutil.NopCloser(bytes.NewBuffer(status))
	if err != nil {
		logger.Errorf("Error while posting status: %v", err)
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
	logger.Infof("Status Event posted successfully [%s]", resp.Status)
	return nil
}
