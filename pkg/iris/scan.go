package iris

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/leanix/leanix-k8s-connector/pkg/iris/models"
	"strconv"
	time2 "time"

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
	var scannedServices []models.DiscoveryItem

	clusterDTO, err := mapper.GetCluster(kubernetesConfig.Cluster, kubernetesAPI)
	if err != nil {
		return err
	}
	// Aggregate cluster information for the event
	namespaces, err := kubernetesAPI.Namespaces(kubernetesConfig.BlackListedNamespaces)
	if err != nil {
		return err
	}

	// Metadata for the event
	scope := fmt.Sprintf("workspace/%s", workspaceId)
	source := fmt.Sprintf("kubernetes/%s#%s", clusterDTO.name, s.runId)

	for _, namespace := range namespaces.Items {
		deploymentsPerSoftwareArtifact := map[string][]models.Deployment{}
		// collect all deployments
		deployments, err := mapper.GetDeployments(namespace.Name, kubernetesAPI)
		if err != nil {
			return err
		}
		// collect all cronjobs

		// Group deployments, cronjobs and jobs by software artifact and create an event
		for _, deployment := range deployments {
			if _, ok := deploymentsPerSoftwareArtifact[deployment.Name]; !ok {
				deploymentsPerSoftwareArtifact[deployment.Name] = make([]models.Deployment, 0)
			}
			// Key is service/softwareArtifact name, value is the list of deployments connected to it
			deploymentsPerSoftwareArtifact[deployment.Name] = append(deploymentsPerSoftwareArtifact[deployment.Name], deployment)
		}

		// create kubernetes event for every software artifact
		for service, deploymentList := range deploymentsPerSoftwareArtifact {

			result := models.Cluster{
				Namespace: models.Namespace{
					Name: namespace.Name,
				},
				Deployments: deploymentList,
				Name:        clusterDTO.name,
				Os:          clusterDTO.osImage,
				K8sVersion:  clusterDTO.k8sVersion,
				NoOfNodes:   strconv.Itoa(clusterDTO.nodesCount),
			}

			// Metadata for the event
			id := fmt.Sprintf("%s", uuid.New().String())
			subject := fmt.Sprintf("softwareArtifact/%s", service)
			time := time2.Now().Format(time2.RFC3339)

			// Build service/softwareArtifact event
			serviceEvent := New().
				Id(id).
				Source(source).
				Subject(subject).
				Type(typeAsK8s).
				Scope(scope).
				Time(time).
				Cluster(result).
				Build()

			scannedServices = append(scannedServices, serviceEvent)
		}

	}

	scannedObjectsByte, err := storage.Marshal(scannedServices)
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
