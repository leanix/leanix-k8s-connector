package iris

import (
	"encoding/json"
	"fmt"
	"github.com/leanix/leanix-k8s-connector/pkg/iris/models"
	"github.com/leanix/leanix-k8s-connector/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	"strconv"
	time2 "time"

	"github.com/leanix/leanix-k8s-connector/pkg/kubernetes"
	"github.com/leanix/leanix-k8s-connector/pkg/storage"
	"k8s.io/client-go/rest"
)

type Scanner interface {
	Scan(getKubernetesApiFunc kubernetes.GetKubernetesAPI, config *rest.Config, workspaceId string, configurationName string, accessToken string) error
}

type scanner struct {
	irisApi API
	runId   string
}

func NewScanner(kind string, uri string, runId string) Scanner {
	api := NewApi(http.DefaultClient, kind, uri)
	return &scanner{
		irisApi: api,
		runId:   runId,
	}
}

type kubernetesConfig struct {
	ID                    string   `json:"id"`
	Cluster               string   `json:"cluster"`
	BlackListedNamespaces []string `json:"blacklistedNamespaces"`
}

const (
	IN_PROGRESS        string = "IN_PROGRESS"
	FAILED             string = "FAILED"
	SUCCESSFUL         string = "SUCCESSFUL"
	SUCCESSFUL_WARNING string = "SUCCESSFUL_WARNING"
	ERROR              string = "ERROR"
	WARNING            string = "WARNING"
	INFO               string = "INFO"
)

const StatusErrorFormat = "Scan failed while posting status. RunId: [%s], with reason: '%v'"

func (s *scanner) Scan(getKubernetesAPI kubernetes.GetKubernetesAPI, config *rest.Config, workspaceId string, configurationName string, accessToken string) error {
	configuration, err := s.irisApi.GetConfiguration(configurationName, accessToken)
	if err != nil {
		return err
	}
	logger.Infof("Scan started for RunId: [%s]", s.runId)
	logger.Infof("Configuration used: %s", configuration)
	kubernetesConfig := kubernetesConfig{}
	err = json.Unmarshal(configuration, &kubernetesConfig)
	if err != nil {
		return err
	}
	err = s.ShareStatus(kubernetesConfig.ID, workspaceId, accessToken, IN_PROGRESS, "Started Kubernetes Scan")
	if err != nil {
		logger.Errorf(StatusErrorFormat, s.runId, err)
		return err
	}

	kubernetesAPI, err := getKubernetesAPI(config)
	if err != nil {
		return s.LogAndShareError("Scan failed while getting Kubernetes API. RunId: [%s], with reason: '%v'", ERROR, err, kubernetesConfig.ID, workspaceId, accessToken)
	}

	logger.Info("Retrieved kubernetes config successfully")
	err = s.ShareAdminLogs(kubernetesConfig.ID, workspaceId, accessToken, INFO, "Retrieved kubernetes config successfully")
	if err != nil {
		logger.Errorf(StatusErrorFormat, s.runId, err)
		return err
	}
	mapper := NewMapper(kubernetesAPI, kubernetesConfig.Cluster, workspaceId, kubernetesConfig.BlackListedNamespaces, s.runId)

	nodes, err := kubernetesAPI.Nodes()
	if err != nil {
		return s.LogAndShareError("Scan failed while retrieving k8s cluster nodes. RunId: [%s], with reason: '%v'", ERROR, err, kubernetesConfig.ID, workspaceId, accessToken)
	}

	clusterDTO, err := mapper.MapCluster(kubernetesConfig.Cluster, nodes)
	if err != nil {
		return s.LogAndShareError("Scan failed while aggregating cluster information. RunId: [%s], with reason: '%v'", ERROR, err, kubernetesConfig.ID, workspaceId, accessToken)
	}

	// Aggregate cluster information for the event
	namespaces, err := kubernetesAPI.Namespaces(kubernetesConfig.BlackListedNamespaces)
	if err != nil {
		return s.LogAndShareError("Scan failed while retrieving k8s namespaces. RunId: [%s], with reason: '%v'", ERROR, err, kubernetesConfig.ID, workspaceId, accessToken)
	}

	events, err := s.ScanNamespace(kubernetesAPI, mapper, namespaces.Items, clusterDTO, workspaceId, kubernetesConfig)
	if err != nil {
		return s.LogAndShareError("Scan failed while retrieving k8s deployments. RunId: [%s], with reason: '%v'", ERROR, err, kubernetesConfig.ID, workspaceId, accessToken)
	}

	scannedObjectsByte, err := storage.Marshal(events)
	if err != nil {
		return s.LogAndShareError("Marshall scanned services", ERROR, err, kubernetesConfig.ID, workspaceId, accessToken)
	}

	err = s.irisApi.PostResults(scannedObjectsByte, accessToken)
	if err != nil {
		return s.LogAndShareError("Scan failed while posting results. RunId: [%s], with reason: '%v'", ERROR, err, kubernetesConfig.ID, workspaceId, accessToken)
	}

	logger.Infof("Scan Finished for RunId: [%s]", s.runId)
	err = s.ShareStatus(kubernetesConfig.ID, workspaceId, accessToken, SUCCESSFUL, "Successfully Scanned")
	if err != nil {
		logger.Errorf(StatusErrorFormat, s.runId, err)
		return err
	}
	return err
}

func (s scanner) ScanNamespace(k8sApi *kubernetes.API, mapper Mapper, namespaces []corev1.Namespace, cluster ClusterDTO, workspaceId string, config kubernetesConfig) (interface{}, error) {
	// Metadata for the event
	scope := fmt.Sprintf("workspace/%s/configuration/%s", workspaceId, config.ID)
	source := fmt.Sprintf("kubernetes/%s#%s", cluster.name, s.runId)
	var events []interface{}
	startReplay := s.CreateStartReplay(workspaceId, config)
	for _, namespace := range namespaces {
		// collect all deployments
		deployments, err := k8sApi.Deployments(namespace.Name)
		if err != nil {
			return nil, err
		}

		services, err := k8sApi.Services(namespace.Name)
		if err != nil {
			return nil, err
		}

		mappedDeployments, err := mapper.MapDeployments(deployments, services)
		if err != nil {
			return nil, err
		}

		// create kubernetes event for namespace
		discoveryEvent := s.CreateDiscoveryEvent(namespace, mappedDeployments, &cluster, source, scope)

		events = append(events, discoveryEvent)
	}

	endReplay := s.CreateEndReplay(workspaceId, config)

	events = append(events, startReplay, endReplay)

	return events, nil
}

func (s scanner) LogAndShareError(message string, loglevel string, err error, id string, workspaceId string, accessToken string) error {
	logger.Errorf(message, s.runId, err)
	statusErr := s.ShareStatus(id, workspaceId, accessToken, FAILED, "Kubernetes scan failed")
	if statusErr != nil {
		logger.Errorf(StatusErrorFormat, s.runId, statusErr)
	}
	logErr := s.ShareAdminLogs(id, workspaceId, accessToken, loglevel, fmt.Sprintf(message, s.runId, err))
	if logErr != nil {
		logger.Errorf(StatusErrorFormat, s.runId, logErr)
	}
	return err
}

func (s *scanner) CreateDiscoveryEvent(namespace corev1.Namespace, deployments []models.Deployment, clusterDTO *ClusterDTO, source string, scope string) models.DiscoveryEvent {
	result := models.Cluster{
		Namespace: models.Namespace{
			Name: namespace.Name,
		},
		Deployments: deployments,
		Name:        clusterDTO.name,
		Os:          clusterDTO.osImage,
		K8sVersion:  clusterDTO.k8sVersion,
		NoOfNodes:   strconv.Itoa(clusterDTO.nodesCount),
	}

	// Metadata for the event
	class := fmt.Sprintf("discoveryItem/service/kubernetes")
	id := fmt.Sprintf("%s/%s", class, scope)
	time := time2.Now().Format(time2.RFC3339)
	header := models.HeaderProperties{
		HeaderId:    id,
		HeaderScope: scope,
		HeaderClass: class,
		HeaderType:  fmt.Sprintf("state"),
	}
	body := models.DiscoveryItem{
		State: models.State{
			Name:   namespace.Name,
			Source: source,
			Time:   time,
			Data: models.Data{
				Cluster: result,
			},
		},
	}

	// Build service/softwareArtifact event
	discoveryEvent := New().
		Header(header).
		Body(body).
		Build()
	return discoveryEvent
}

// Command Events
func (s *scanner) CreateStartReplay(workspaceId string, config kubernetesConfig) models.CommandEvent {
	// Metadata for the command event
	eventType := fmt.Sprintf("command")
	action := fmt.Sprintf("startReplay")
	scope := fmt.Sprintf("workspace/%s/configuration/%s", workspaceId, config.ID)
	header := models.CommandProperties{
		Type:   eventType,
		Action: action,
		Scope:  scope,
	}

	startReplayEvent := NewCommand().Header(header).Build()
	return startReplayEvent
}

func (s *scanner) CreateEndReplay(workspaceId string, config kubernetesConfig) models.CommandEvent {
	// Metadata for the command event
	eventType := fmt.Sprintf("command")
	action := fmt.Sprintf("endReplay")
	scope := fmt.Sprintf("workspace/%s/configuration/%s", workspaceId, config.ID)
	header := models.CommandProperties{
		Type:   eventType,
		Action: action,
		Scope:  scope,
	}

	endReplayEvent := NewCommand().Header(header).Build()
	return endReplayEvent
}

// Status
func (s *scanner) ShareStatus(configid string, workspaceId string, accessToken string, status string, message string) error {
	var statusArray []StatusItem
	statusObject := NewStatusEvent(configid, s.runId, workspaceId, status, message)
	statusArray = append(statusArray, *statusObject)
	statusByte, err := storage.Marshal(statusArray)
	err = s.irisApi.PostStatus(statusByte, accessToken)
	if err != nil {
		logger.Debugf("Failed sharing status for RunId: [%s], with reason %v", s.runId, err)
		return err
	}
	return nil
}

func (s *scanner) ShareAdminLogs(configid string, workspaceId string, accessToken string, loglevel string, message string) error {
	var statusArray []StatusItem
	statusObject := NewAdminLogEvent(configid, s.runId, workspaceId, loglevel, message)
	statusArray = append(statusArray, *statusObject)
	statusByte, err := storage.Marshal(statusArray)
	err = s.irisApi.PostStatus(statusByte, accessToken)
	if err != nil {
		logger.Debugf("Failed sharing admin logs for RunId: [%s], with reason %v", s.runId, err)
		return err
	}
	return nil
}
