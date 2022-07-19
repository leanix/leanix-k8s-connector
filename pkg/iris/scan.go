package iris

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/leanix/leanix-k8s-connector/pkg/iris/models"
	"net/http"
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
	api := NewApi(http.DefaultClient, kind, uri)
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

const (
	STARTED     string = "STARTED"
	IN_PROGRESS string = "IN_PROGRESS"
	FAILED      string = "FAILED"
	SUCCESSFUL  string = "SUCCESSFUL"
)

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
	err = s.ShareStatus(kubernetesConfig.ID, workspaceId, accessToken, STARTED, "Started Kubernetes Scan")
	if err != nil {
		log.Errorf("Scan failed while posting status. RunId: [%s], with reason %v", s.runId, err)
	}
	kubernetesAPI, err := kubernetes.NewAPI(config)
	if err != nil {
		return err
	}
	log.Info("Retrieved kubernetes config Successfully")
	err = s.ShareStatus(kubernetesConfig.ID, workspaceId, accessToken, IN_PROGRESS, "Retrieved Kubernetes configuration Successfully")
	if err != nil {
		log.Errorf("Scan failed while posting status. RunId: [%s], with reason %v", s.runId, err)
	}
	mapper := NewMapper(kubernetesAPI, kubernetesConfig.Cluster, workspaceId, kubernetesConfig.BlackListedNamespaces, s.runId)
	var scannedServices []models.DiscoveryItem

	nodes, err := kubernetesAPI.Nodes()
	if err != nil {
		return err
	}
	clusterDTO, err := mapper.GetCluster(kubernetesConfig.Cluster, nodes)
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
		// collect all deployments
		deployments, err := kubernetesAPI.Deployments(namespace.Name)
		if err != nil {
			err = s.ShareStatus(kubernetesConfig.ID, workspaceId, accessToken, FAILED, err.Error())
		if err != nil {
			log.Errorf("Scan failed while posting status. RunId: [%s], with reason %v", s.runId, err)
		}return err
		}

		services, err := kubernetesAPI.Services(namespace.Name)
		if err != nil {
			return err
		}
		mappedDeployments, err := mapper.GetDeployments(deployments, services)
		if err != nil {
			return err
		}

		// collect all cronjobs

		// create kubernetes event for every software artifact
		for _, deployment := range mappedDeployments {

			result := models.Cluster{
				Namespace: models.Namespace{
					Name: namespace.Name,
				},
				Deployment: deployment,
				Name:       clusterDTO.name,
				Os:         clusterDTO.osImage,
				K8sVersion: clusterDTO.k8sVersion,
				NoOfNodes:  strconv.Itoa(clusterDTO.nodesCount),
			}

			// Metadata for the event

			id := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s-%s", namespace.Name, deployment.Name))))
			subject := fmt.Sprintf("workload/%s", deployment.Name)
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
		err = s.ShareStatus(kubernetesConfig.ID, workspaceId, accessToken, FAILED, err.Error())
		if err != nil {
			log.Errorf("Scan failed while posting status. RunId: [%s], with reason %v", s.runId, err)
		}
		return err
	}
	err = s.api.PostResults(scannedObjectsByte, accessToken)
	if err != nil {
		log.Errorf("Scan failed while posting results. RunId: [%s], with reason %v", s.runId, err)
		err = s.ShareStatus(kubernetesConfig.ID, workspaceId, accessToken, FAILED, err.Error())
		if err != nil {
			log.Errorf("Scan failed while posting status. RunId: [%s], with reason %v", s.runId, err)
		}
		return err
	}
	log.Infof("Scan Finished for RunId: [%s]", s.runId)
	err = s.ShareStatus(kubernetesConfig.ID, workspaceId, accessToken, SUCCESSFUL, "Successfully Scanned")
	if err != nil {
		log.Errorf("Scan failed while posting status. RunId: [%s], with reason %v", s.runId, err)
		return err
	}
	return nil
}

func (s *scanner) ShareStatus(configid string, workspaceId string, accessToken string, status string, message string) error {
	var statusArray []StatusItem
	statusObject := NewStatusEvent(configid, s.runId, workspaceId, FAILED, message)
	statusArray = append(statusArray, *statusObject)
	statusByte, err := storage.Marshal(statusArray)
	err = s.api.PostStatus(statusByte, accessToken)
	if err != nil {
		log.Debugf("Failed sharing status for RunId: [%s], with reason %v", s.runId, err)
		return err
	}
	return nil
}
