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

const (
	CONNECTOR_VERSION string = "6.4.1"
)

type SelfStartResponse struct {
	RunId                  string                 `json:"runId"`
	BindingKey             BindingKey             `json:"bindingKey"`
	ConnectorConfiguration ConnectorConfiguration `json:"connectorConfiguration"`
	LdifResultUrl          string                 `json:"ldifResultUrl"`
	ProgressCallbackUrl    string                 `json:"progressCallbackUrl"`
	ConnectorLoggingUrl    string                 `json:"connectorLoggingUrl"`
}

type BindingKey struct {
	ConnectorType       string `json:"connectorType"`
	ConnectorId         string `json:"connectorId"`
	ProcessingDirection string `json:"processingDirection"`
	ProcessingMode      string `json:"processingMode"`
	LXVersion           string `json:"lxVersion"`
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
	datasourceRunUrl := fmt.Sprintf("https://%s/services/integration-hub/v1/datasources/%s/selfStart", fqdn, datasource)
	req, err := http.NewRequest("POST", datasourceRunUrl, nil)
	req.Header.Set("connector-version", CONNECTOR_VERSION)
	req.Header.Set("connector-name", "k8s-connector")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	if err != nil {
		log.Infof("SelfStartRun: Error while starting run for datasource: %s", datasourceRunUrl)
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
	errorBinding := validateBindingKey(startResponse.BindingKey)
	if errorBinding != nil {
		_, err = UpdateFailedProgressStatus(startResponse.ProgressCallbackUrl, "INVALID BINDING KEY: ABORTING IHUB RUN.")
		if err != nil {
			log.Errorf("Failed to update progress[%s] to Integration Hub", FAILED)
		}
		return &startResponse, errorBinding
	}
	errorConfig := validateConnectorConfiguration(startResponse.ConnectorConfiguration)
	if errorConfig != nil {
		_, err = UpdateFailedProgressStatus(startResponse.ProgressCallbackUrl, "INVALID CONNECTOR CONFIGURATION: ABORTING IHUB RUN.")
		if err != nil {
			log.Errorf("SelfStartRun: Failed to update progress[%s] to Integration Hub", FAILED)
		}
		return &startResponse, errorConfig
	}
	log.Info("Connector Configuration is validated")
	log.Info("Binding Key is validated")
	return &startResponse, nil
}

func validateBindingKey(bindingKey BindingKey) error {
	if bindingKey.ConnectorType == "" {
		return fmt.Errorf("INVALID BINDING KEY: CONNECTOR TYPE CANNOT BE EMPTY")
	}

	if bindingKey.ConnectorId == "" {
		return fmt.Errorf("INVALID BINDING KEY: RESOLVE LABEL CANNOT BE EMPTY IF THE RESOLVE STRATEGY IS 'LABEL'")
	}

	if bindingKey.ProcessingDirection == "" {
		return fmt.Errorf("INVALID BINDING KEY: PROCESSING DIRECTION CANNOT BE EMPTY")
	}

	if bindingKey.ProcessingMode == "" {
		return fmt.Errorf("INVALID BINDING KEY: PROCESSING MODE CANNOT BE EMPTY")
	}

	if bindingKey.LXVersion == "" {
		return fmt.Errorf("INVALID BINDING KEY: LXVERSION CANNOT BE EMPTY")
	}

	return nil
}

func validateConnectorConfiguration(configuration ConnectorConfiguration) error {
	if configuration.ResolveStrategy == "" {
		return fmt.Errorf("INVALID CONNECTOR CONFIGURATION: RESOLVE STRATEGY CANNOT BE EMPTY")
	}

	if configuration.ResolveStrategy == "label" && configuration.ResolveLabel == "" {
		return fmt.Errorf("INVALID CONNECTOR CONFIGURATION: RESOLVE LABEL CANNOT BE EMPTY IF THE RESOLVE STRATEGY IS 'LABEL'")
	}

	if configuration.ClusterName == "" {
		return fmt.Errorf("INVALID CONNECTOR CONFIGURATION: CLUSTER NAME CANNOT BE EMPTY")
	}
	return nil
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
		log.Errorf("UpdateProgress: Failed to marshal the request body: '%s'", err)
		return status, err
	}
	log.Debugf("UpdateProgress: callback url: '%s', body: '%s'", progressCallbackUrl, body)
	resp, err := http.Post(progressCallbackUrl, "application/json", bytes.NewReader(marshal))
	if err != nil {
		return status, err
	}
	if resp.StatusCode != 200 {
		responseData, readErr := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if readErr != nil {
			return status, readErr
		}
		err := fmt.Errorf("Integration Hub connector status[%s] could not be updated %s: '%s'\n", status, resp.Status, responseData)
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
