package leanix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/leanix/leanix-k8s-connector/pkg/storage"
	"io/ioutil"
	"net/http"
	"net/url"
)

type SelfStartResponse struct {
	RunId                  string                 `json:"runId"`
	ConnectorConfiguration ConnectorConfiguration `json:"connectorConfiguration"`
	LdifResultUrl          string                 `json:"ldifResultUrl"`
	ProgressCallbackUrl    string                 `json:"progressCallbackUrl"`
}

type ConnectorConfiguration struct {
	ResolveStrategy string `json:"resolveStrategy"`
}

const (
	INPROGRESS string = "INPROGRESS"
	FAILED     string = "FAILED"
	FINISHED   string = "FINISHED"
)

// SelfStartRun initiates the Integration Hub run and response with id
func SelfStartRun(fqdn string, accessToken string, datasource string) (*SelfStartResponse, error) {
	datasourceRunUrl := "https://" + fqdn + "/services/integration-hub/v1/datasourceRuns/name/" + datasource + "/selfStart"
	req, err := http.NewRequest("GET", datasourceRunUrl, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		err := fmt.Errorf("Integration Hub run could not be started: %s\n", resp.Status)
		return nil, err
	}
	defer resp.Body.Close()
	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	startResponse := SelfStartResponse{}
	json.Unmarshal(responseData, &startResponse)
	return &startResponse, nil
}

// UpdateProgress Updates progress to Integration Hub
func UpdateProgress(progressCallbackUrl string, status string, message string) (string, error) {
	body := struct {
		Message        string `json:"message"`
		Status         string `json:"status"`
		ProgressOrigin string `json:"progressOrigin"`
	}{Message: message, Status: status, ProgressOrigin: "CONNECTOR"}
	marshal, err := json.Marshal(body)
	if err != nil {
		fmt.Printf("Failed to marshal the request body")
		return status, err
	}
	resp, err := http.Post(progressCallbackUrl, "application/json", bytes.NewReader(marshal))
	if err != nil {
		return status, err
	}
	if resp.StatusCode != 200 {
		err := fmt.Errorf("Integration Hub connector status[%s] could not be updated: %s\n", status, resp.Status)
		return status, err
	}
	return status, nil
}

func UpdateInProgressStatus(progressCallbackUrl string, message string) (string, error) {
	return UpdateProgress(progressCallbackUrl, INPROGRESS, message)
}

func UpdateFailedProgressStatus(progressCallbackUrl string, message string) (string, error) {
	return UpdateProgress(progressCallbackUrl, FAILED, message)
}

// UploadLdif Uploads ldif for processing to provided ldif result url
func UploadLdif(ldifResultUrl string, ldif []byte) (string, error) {
	parse, err := url.Parse(ldifResultUrl)
	if err != nil {
		return "", err
	}
	err = storage.UploadFileToSasUrl(*parse, ldif)
	if err != nil {
		return "", err
	}
	return "", nil
}
