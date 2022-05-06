package iris

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/leanix/leanix-k8s-connector/pkg/kubernetes"
	"github.com/leanix/leanix-k8s-connector/pkg/mapper"
	"github.com/leanix/leanix-k8s-connector/pkg/storage"
	"k8s.io/client-go/rest"
)

type kubernetesConfig struct {
	ID       string `json:"id"`
	Schedule string `json:"schedule"`
	Cluster  string `json:"cluster"`
}

/* {
	"id": "9aeb0fdf-c01e-0131-0922-9eb54906e209",
	"scope": "workspace/123e4567-e89b-12d3-a456-426614174000",
	"type": "leanix.vsm.item-discovered.deployment",
	"source": "kubernetes/some-cluster",
	"time": "2019-11-18T15:13:39.4589254Z",
	"subject": "deployment/deployment-name",
	"data": {
		"key": "value",
	}
  } */

type DiscoveryItem struct {
	ID      string      `json:"id"`
	Scope   string      `json:"scope"`
	Type    string      `json:"type"`
	Source  string      `json:"source"`
	Time    string      `json:"time"`
	Subject string      `json:"subject"`
	Data    interface{} `json:"data"`
}

func ScanKubernetes(config *rest.Config, workspaceId string, accessToken string) ([]mapper.KubernetesObject, error) {
	configUrl := "http://localhost:8080/configurations/" + workspaceId + "/kubernetesConnector/1234"
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
	log.Infof("Configuration used: %s", responseData)
	kubernetesConfig := kubernetesConfig{}
	json.Unmarshal(responseData, &kubernetesConfig)
	fmt.Println(kubernetesConfig)
	kubernetesAPI, err := kubernetes.NewAPI(config)
	if err != nil {
		return nil, err
	}
	var blacklistedNamespacesList []string
	var scannedObjects []DiscoveryItem
	namespaces, err := kubernetesAPI.Namespaces(blacklistedNamespacesList)
	if err != nil {
		return nil, err
	}
	deployments, err := GetDeployments(kubernetesConfig.Cluster, workspaceId, namespaces, kubernetesAPI)
	if err != nil {
		return nil, err
	}

	scannedObjects = append(scannedObjects, deployments...)
	scannedObjectsByte, err := storage.Marshal(scannedObjects)
	if err != nil {
		return nil, err
	}
	resultUrl := "http://127.0.0.1:8080/results"
	postReq, err := http.NewRequest("POST", resultUrl, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Body = ioutil.NopCloser(bytes.NewBuffer(scannedObjectsByte))
	if err != nil {
		log.Infof("Post results request: Error while posting results: %s", resultUrl)
		return nil, err
	}
	response, err := http.DefaultClient.Do(postReq)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {

		return nil, err
	}
	if response.StatusCode == 200 {
		log.Info(response)
		return nil, nil
	}
	return nil, nil
}
