package iris

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

const IRIS_URI string = "https://eu.leanix.net/services/vsm-iris/v1/"

func GetConfiguration(configurationName string, accessToken string) ([]byte, error) {
	configUrl := fmt.Sprintf(IRIS_URI+"configurations/kubernetesConnector/%s", configurationName)
	req, err := http.NewRequest("GET", configUrl, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	if err != nil {
		log.Infof("SelfStartRun: Error while retrieving configuration from %s: %v", configUrl, err)
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return responseData, nil
}

func PostResults(results []byte, accessToken string) (string, error) {
	resultUrl := IRIS_URI + "results"
	postReq, err := http.NewRequest("POST", resultUrl, nil)
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("Authorization", "Bearer "+accessToken)
	postReq.Body = ioutil.NopCloser(bytes.NewBuffer(results))
	if err != nil {
		log.Infof("Post results request: Error while posting results: %s", resultUrl)
		return postReq.URL.String(), err
	}
	resp, err := http.DefaultClient.Do(postReq)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		responseData, readErr := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if readErr != nil {
			return "", readErr
		}
		err := fmt.Errorf("Posting results status[%s]could not be processed: '%s'\n", resp.Status, responseData)
		return "", err
	}
	return "Events posted successfully", nil
}
