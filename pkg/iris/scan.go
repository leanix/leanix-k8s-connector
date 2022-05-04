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

func ScanKubernetes(config *rest.Config, workspaceId string) ([]mapper.KubernetesObject, error) {
	resp, err := http.Get("http://127.0.0.1:8080/k8sConnector/getConfiguration")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
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
	response, err := http.Post("http://127.0.0.1:8080/k8sConnector/postResults", "application/json", bytes.NewBuffer(scannedObjectsByte))
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected response code %d", response.StatusCode)
	}
	return nil, nil
}
