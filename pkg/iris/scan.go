package iris

import (
	"encoding/json"

	"github.com/leanix/leanix-k8s-connector/pkg/kubernetes"
	"github.com/leanix/leanix-k8s-connector/pkg/storage"
	"github.com/op/go-logging"
	"k8s.io/client-go/rest"
)

type kubernetesConfig struct {
	ID                    string   `json:"id"`
	Cluster               string   `json:"cluster"`
	BlackListedNamespaces []string `json:"blacklistedNamespaces"`
}

/* {
	"id": "9aeb0fdf-c01e-0131-0922-9eb54906e209",
	"scope": "workspace/123e4567-e89b-12d3-a456-426614174000",
	"type": "leanix.vsm.item-discovered.softwareArtifact",
	"source": "kubernetes/some-cluster",
	"time": "2019-11-18T15:13:39.4589254Z",
	"subject": "softwareArtifact/app1",
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

var log = logging.MustGetLogger("leanix-k8s-connector")

func ScanKubernetes(config *rest.Config, workspaceId string, configurationName string, accessToken string) error {
	configuration, err := GetConfiguration(configurationName, accessToken)
	if err != nil {
		return err
	}
	log.Infof("Configuration used: %s", configuration)
	kubernetesConfig := kubernetesConfig{}
	json.Unmarshal(configuration, &kubernetesConfig)
	kubernetesAPI, err := kubernetes.NewAPI(config)
	if err != nil {
		return err
	}
	var scannedObjects []DiscoveryItem
	namespaces, err := kubernetesAPI.Namespaces(kubernetesConfig.BlackListedNamespaces)
	if err != nil {
		return err
	}
	deployments, err := GetDeployments(kubernetesConfig.Cluster, workspaceId, namespaces, kubernetesAPI)
	if err != nil {
		return err
	}

	scannedObjects = append(scannedObjects, deployments...)
	scannedObjectsByte, err := storage.Marshal(scannedObjects)
	if err != nil {
		return err
	}
	result, err := PostResults(scannedObjectsByte, accessToken)
	if err != nil {
		return err
	}
	log.Info(result)
	return nil
}
