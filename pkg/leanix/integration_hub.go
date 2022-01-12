package leanix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/leanix/leanix-k8s-connector/pkg/storage"
	"github.com/op/go-logging"
)

type SelfStartResponse struct {
	RunId                  string                 `json:"runId"`
	ConnectorConfiguration ConnectorConfiguration `json:"connectorConfiguration"`
	LdifResultUrl          string                 `json:"ldifResultUrl"`
	ProgressCallbackUrl    string                 `json:"progressCallbackUrl"`
	ConnectorLoggingUrl    string                 `json:"connectorLoggingUrl" `
}

type ConnectorConfiguration struct {
	ResolveStrategy       string `json:"resolveStrategy"`
	ResolveLabel          string `json:"resolveLabel"`
	ClusterName           string `json:"clusterName"`
	EnabledLabelWhitelist bool   `json:"enabledLabelWhitelist"` // special case for Numberger customer (Until execution groups is supported by Integration Hub)
}

const (
	IN_PROGRESS string = "IN_PROGRESS"
	FAILED      string = "FAILED"
	FINISHED    string = "FINISHED"
)

var log = logging.MustGetLogger("leanix-k8s-connector")

// SelfStartRun initiates the Integration Hub run and response with id
func SelfStartRun(fqdn string, accessToken string, datasource string) (*SelfStartResponse, error) {
	datasourceRunUrl := "https://" + fqdn + "/services/integration-hub/v1/datasourceRuns/name/" + datasource + "/selfStart"
	req, err := http.NewRequest("GET", datasourceRunUrl, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	if err != nil {
		return nil, err
	}
	log.Infof("Initiating connection to Integration Hub API with dataSource name: %s with region value: %s\n", datasource, fqdn)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		err := fmt.Errorf("Integration Hub run could not be started: %s\n ", resp.Status)
		responseData, readErr := ioutil.ReadAll(resp.Body)
		if readErr != nil {
			return nil, readErr
		}
		log.Errorf("Integration Hub run failed with error message: %s\n ", responseData)
		return nil, err
	}
	defer resp.Body.Close()
	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	startResponse := SelfStartResponse{}
	json.Unmarshal(responseData, &startResponse)
	valid := validateConnectorConfiguration(startResponse.ConnectorConfiguration)
	if valid {
		log.Info("Connector Configuration is validated")
		return &startResponse, nil
	} else {
		_, err = UpdateFailedProgressStatus(startResponse.ProgressCallbackUrl, "INVALID CONNECTOR CONFIGURATION: ABORTING IHUB RUN.")
		if err != nil {
			log.Infof("Failed to update progress[%s] to Integration Hub", FAILED)
		}
		return &startResponse, fmt.Errorf("INVALID CONNECTOR CONFIGURATION: CHECK AND RE-RUN WITH VALID CONNECTOR DEFINITION")
	}
}

func validateConnectorConfiguration(configuration ConnectorConfiguration) bool {
	if configuration.ResolveStrategy == "" {
		log.Errorf("INVALID CONNECTOR CONFIGURATION: RESOLVE STRATEGY CANNOT BE EMPTY")
		return false
	}

	if configuration.ResolveStrategy == "label" && configuration.ResolveLabel == "" {
		log.Errorf("INVALID CONNECTOR CONFIGURATION: RESOLVE LABEL CANNOT BE EMPTY IF THE RESOLVE STRATEGY IS 'LABEL'")
		return false
	}

	if configuration.ResolveStrategy == "label/namespace" && configuration.ResolveLabel == "" {
		log.Errorf("INVALID CONNECTOR CONFIGURATION: RESOLVE STRATEGY CAN BE EITHER 'LABEL' OR 'NAMESPACE', ONLY ONE CAN BE SET")
		return false
	}

	if configuration.ClusterName == "" {
		log.Errorf("INVALID CONNECTOR CONFIGURATION: CLUSTER NAME CANNOT BE EMPTY")
		return false
	}
	return true
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
		err := fmt.Errorf("Integration Hub connector status[%s] could not be updated %s\n", status, resp.Status)
		return status, err
	}
	return status, nil
}

func UpdateInProgressStatus(progressCallbackUrl string, message string) (string, error) {
	return UpdateProgress(progressCallbackUrl, IN_PROGRESS, message)
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
