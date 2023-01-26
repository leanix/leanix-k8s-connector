package iris

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/leanix/leanix-k8s-connector/pkg/iris/models"
	"github.com/leanix/leanix-k8s-connector/pkg/kubernetes"
	"github.com/leanix/leanix-k8s-connector/pkg/logger"
	"github.com/leanix/leanix-k8s-connector/pkg/storage"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"net/http"
	"strconv"
	time2 "time"
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

	ecstEvents, oldEvents, err := s.ScanNamespace(kubernetesAPI, mapper, namespaces.Items, clusterDTO, workspaceId, kubernetesConfig)
	if err != nil {
		return s.LogAndShareError("Scan failed while retrieving k8s deployments. RunId: [%s], with reason: '%v'", ERROR, err, kubernetesConfig.ID, workspaceId, accessToken)
	}

	scannedObjectsByte, err := storage.Marshal(oldEvents)
	if err != nil {
		return s.LogAndShareError("Marshall scanned services", ERROR, err, kubernetesConfig.ID, workspaceId, accessToken)
	}

	scannedEcstObjectsByte, err := storage.Marshal(ecstEvents)
	if err != nil {
		return s.LogAndShareError("Marshall scanned ECST services", ERROR, err, kubernetesConfig.ID, workspaceId, accessToken)
	}

	err = s.irisApi.PostResults(scannedObjectsByte, accessToken)
	if err != nil {
		return s.LogAndShareError("Scan failed while posting results. RunId: [%s], with reason: '%v'", ERROR, err, kubernetesConfig.ID, workspaceId, accessToken)
	}

	err = s.irisApi.PostEcstResults(scannedEcstObjectsByte, accessToken)
	if err != nil {
		return s.LogAndShareError("Scan failed while posting ECST results. RunId: [%s], with reason: '%v'", ERROR, err, kubernetesConfig.ID, workspaceId, accessToken)
	}

	logger.Infof("Scan Finished for RunId: [%s]", s.runId)
	err = s.ShareStatus(kubernetesConfig.ID, workspaceId, accessToken, SUCCESSFUL, "Successfully Scanned")
	if err != nil {
		logger.Errorf(StatusErrorFormat, s.runId, err)
		return err
	}
	return err
}

func (s scanner) ScanNamespace(k8sApi *kubernetes.API, mapper Mapper, namespaces []corev1.Namespace, cluster ClusterDTO, workspaceId string, config kubernetesConfig) (interface{}, []models.DiscoveryItem, error) {
	//Metadata for OLD Discovery Item
	scope := fmt.Sprintf("workspace/%s", workspaceId)
	source := fmt.Sprintf("kubernetes/%s#%s", cluster.name, s.runId)
	var oldEvents []models.DiscoveryItem

	// Metadata for ECST Discovery Item
	ecstScope := fmt.Sprintf("workspace/%s/configuration/%s", workspaceId, config.ID)
	ecstSource := fmt.Sprintf("kubernetes/%s#%s", cluster.name, s.runId)
	var ecstEvents []interface{}

	startReplay := s.CreateStartReplay(workspaceId, config)
	for _, namespace := range namespaces {
		// collect all deployments
		deployments, err := k8sApi.Deployments(namespace.Name)
		if err != nil {
			return nil, nil, err
		}

		services, err := k8sApi.Services(namespace.Name)
		if err != nil {
			return nil, nil, err
		}

		mappedDeployments, err := mapper.MapDeployments(deployments, services)
		if err != nil {
			return nil, nil, err
		}

		// create OLD disovery item
		oldDiscoveryEvent := s.CreateDiscoveryItem(namespace, mappedDeployments, &cluster, source, scope)
		oldEvents = append(oldEvents, oldDiscoveryEvent)

		// create ECST discovery item for namespace
		ecstDiscoveryEvent := s.CreateEcstDiscoveryEvent(namespace, mappedDeployments, &cluster, ecstSource, ecstScope)
		ecstEvents = append(ecstEvents, ecstDiscoveryEvent)
	}
	endReplay := s.CreateEndReplay(workspaceId, config)

	ecstEvents = append(ecstEvents, startReplay, endReplay)

	return ecstEvents, oldEvents, nil
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

// OLD Discovery Items
func (s *scanner) CreateDiscoveryItem(namespace corev1.Namespace, deployments []models.Deployment, clusterDTO *ClusterDTO, source string, scope string) models.DiscoveryItem {
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
	id := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s-%s", clusterDTO.name, namespace.Name))))
	subject := fmt.Sprintf("namespace/%s", namespace.Name)
	time := time2.Now().Format(time2.RFC3339)

	// Build service/softwareArtifact event
	discoveryEvent := New().
		Id(id).
		Source(source).
		Subject(subject).
		Type(typeAsK8sNamespace).
		Scope(scope).
		Time(time).
		Cluster(result).
		Build()
	return discoveryEvent
}

// ECST Discovery Items
func (s *scanner) CreateEcstDiscoveryEvent(namespace corev1.Namespace, deployments []models.Deployment, clusterDTO *ClusterDTO, source string, scope string) models.DiscoveryEvent {
	result := models.ClusterEcst{
		Namespace:   namespace.Name,
		Deployments: deployments,
		Name:        clusterDTO.name,
		Os:          clusterDTO.osImage,
		K8sVersion:  clusterDTO.k8sVersion,
		NoOfNodes:   strconv.Itoa(clusterDTO.nodesCount),
	}

	// Metadata for the event
	class := fmt.Sprintf("discoveryItem/service/kubernetes")

	idString := fmt.Sprintf("%s/%s", class, scope)
	sum := sha256.Sum256([]byte(idString))

	id := hex.EncodeToString(sum[:])
	time := time2.Now().Format(time2.RFC3339)
	header := models.HeaderProperties{
		Id:    id,
		Scope: scope,
		Class: class,
		Type:  fmt.Sprintf("state"),
	}
	body := models.DiscoveryBody{
		State: models.State{
			Name:   namespace.Name,
			Source: source,
			Time:   time,
			Data: models.Data{
				Cluster: result,
			},
		},
	}

	// Build ECST event
	ecstDiscoveryEvent := NewEcstBuilder().
		Header(header).
		Body(body).
		Build()
	return ecstDiscoveryEvent
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
