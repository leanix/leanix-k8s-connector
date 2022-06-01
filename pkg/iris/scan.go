package iris

import (
	"encoding/json"

	"github.com/leanix/leanix-k8s-connector/pkg/kubernetes"
	"github.com/leanix/leanix-k8s-connector/pkg/storage"
	"github.com/op/go-logging"
	"k8s.io/client-go/rest"
)

type Scanner interface {
	Scan(config *rest.Config, workspaceId string, configurationName string, accessToken string) error
}

type scanner struct {
	api   API
	runId string
}

func NewScanner(kind string, uri string, runId string) Scanner {
	api := NewApi(kind, uri)
	return &scanner{
		api:   api,
		runId: runId,
	}
}

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

var log = logging.MustGetLogger("leanix-k8s-connector")

func (s *scanner) Scan(config *rest.Config, workspaceId string, configurationName string, accessToken string) error {
	configuration, err := s.api.GetConfiguration(configurationName, accessToken)
	if err != nil {
		return err
	}
	log.Infof("Scan started for RunId: [%s]", s.runId)
	log.Infof("Configuration used: %s", configuration)
	kubernetesConfig := kubernetesConfig{}
	err = json.Unmarshal(configuration, &kubernetesConfig)
	if err != nil {
		return err
	}

	kubernetesAPI, err := kubernetes.NewAPI(config)
	if err != nil {
		return err
	}
	log.Info("Retrieved kubernetes config Successfully")
	mapper := NewMapper(kubernetesAPI, kubernetesConfig.Cluster, workspaceId, kubernetesConfig.BlackListedNamespaces, s.runId)

	var scannedObjects []DiscoveryItem
	deployments, err := mapper.GetDeployments()
	if err != nil {
		log.Errorf("Scan failed while fetching deployments. RunId: [%s], with reason %v", s.runId, err)
		return err
	}

	scannedObjects = append(scannedObjects, deployments...)
	scannedObjectsByte, err := storage.Marshal(scannedObjects)
	if err != nil {
		log.Errorf("Scan failed while Unmarshalling results. RunId: [%s], with reason %v", s.runId, err)
		return err
	}
	err = s.api.PostResults(scannedObjectsByte, accessToken)
	if err != nil {
		log.Errorf("Scan failed while posting results. RunId: [%s], with reason %v", s.runId, err)
		return err
	}
	log.Infof("Scan Finished for RunId: [%s]", s.runId)
	return nil
}
