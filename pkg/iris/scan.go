package iris

import (
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
)

type Scanner interface {
	Scan(getKubernetesApiFunc kubernetes.GetKubernetesAPI, config *rest.Config, configurationName string) error
}

type scanner struct {
	configService ConfigService
	eventProducer EventProducer
	runId         string
	workspaceId   string
}

func NewScanner(kind string, uri string, runId string, token string, workspaceId string) Scanner {
	api := NewApi(http.DefaultClient, kind, uri, token)
	configService := NewConfigService(api)
	eventProducer := NewEventProducer(api, runId, workspaceId)
	return &scanner{
		configService: configService,
		eventProducer: eventProducer,
		runId:         runId,
		workspaceId:   workspaceId,
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

func (s *scanner) Scan(getKubernetesAPI kubernetes.GetKubernetesAPI, config *rest.Config, configurationName string) error {
	configuration, err := s.configService.GetConfiguration(configurationName)
	if err != nil {
		return err
	}
	kubernetesConfig := kubernetesConfig{}
	err = json.Unmarshal(configuration, &kubernetesConfig)
	if err != nil {
		return err
	}
	oldResults, err := s.configService.GetScanResults(kubernetesConfig.ID)
	if err != nil {
		return err
	}
	logger.Infof("Scan started for RunId: [%s]", s.runId)
	logger.Infof("Configuration used: %s", configuration)

	err = s.ShareStatus(kubernetesConfig.ID, IN_PROGRESS, "Started Kubernetes Scan")
	if err != nil {
		logger.Errorf(StatusErrorFormat, s.runId, err)
		return err
	}

	kubernetesAPI, err := getKubernetesAPI(config)
	if err != nil {
		return s.LogAndShareError("Scan failed while getting Kubernetes API. RunId: [%s], with reason: '%v'", ERROR, err, kubernetesConfig.ID)
	}

	logger.Info("Retrieved kubernetes config successfully")
	err = s.ShareAdminLogs(kubernetesConfig.ID, INFO, "Retrieved kubernetes config successfully")
	if err != nil {
		logger.Errorf(StatusErrorFormat, s.runId, err)
		return err
	}
	mapper := NewMapper(kubernetesAPI, kubernetesConfig.Cluster, s.workspaceId, kubernetesConfig.BlackListedNamespaces, s.runId)

	nodes, err := kubernetesAPI.Nodes()
	if err != nil {
		return s.LogAndShareError("Scan failed while retrieving k8s cluster nodes. RunId: [%s], with reason: '%v'", ERROR, err, kubernetesConfig.ID)
	}

	clusterDTO, err := mapper.MapCluster(kubernetesConfig.Cluster, nodes)
	if err != nil {
		return s.LogAndShareError("Scan failed while aggregating cluster information. RunId: [%s], with reason: '%v'", ERROR, err, kubernetesConfig.ID)
	}

	// Aggregate cluster information for the event
	namespaces, err := kubernetesAPI.Namespaces(kubernetesConfig.BlackListedNamespaces)
	if err != nil {
		return s.LogAndShareError("Scan failed while retrieving k8s namespaces. RunId: [%s], with reason: '%v'", ERROR, err, kubernetesConfig.ID)
	}
	//Fetch old scan results
	ecstDiscoveredData, legacyData, err := s.ScanNamespace(kubernetesAPI, mapper, namespaces.Items, clusterDTO)
	if err != nil {
		return s.LogAndShareError("Scan failed while retrieving k8s deployments. RunId: [%s], with reason: '%v'", ERROR, err, kubernetesConfig.ID)
	}

	err = s.eventProducer.ProcessResults(ecstDiscoveredData, oldResults, kubernetesConfig.ID)
	if err != nil {
		return s.LogAndShareError("Scan failed while posting ECST results. RunId: [%s], with reason: '%v'", ERROR, err, kubernetesConfig.ID)
	}

	err = s.eventProducer.PostLegacyResults(legacyData)
	if err != nil {
		return s.LogAndShareError("Scan failed while posting legacy results. RunId: [%s], with reason: '%v'", ERROR, err, kubernetesConfig.ID)
	}

	logger.Infof("Scan Finished for RunId: [%s]", s.runId)
	err = s.ShareStatus(kubernetesConfig.ID, SUCCESSFUL, "Successfully Scanned")
	if err != nil {
		logger.Errorf(StatusErrorFormat, s.runId, err)
		return err
	}
	return err
}

func (s scanner) ScanNamespace(k8sApi *kubernetes.API, mapper Mapper, namespaces []corev1.Namespace, cluster ClusterDTO) ([]models.Data, []models.DiscoveryData, error) {
	var legacyData = make([]models.DiscoveryData, 0)
	var ecstData = make([]models.Data, 0)

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

		mappedDeploymentsEcst, err := mapper.MapDeploymentsEcst(deployments, services)
		if err != nil {
			return nil, nil, err
		}

		// create OLD disovery item
		oldDataDiscoveryData := s.CreateLegacyDiscoveryData(namespace, mappedDeployments, &cluster)
		legacyData = append(legacyData, oldDataDiscoveryData)

		// create ECST discovery item for namespace
		ecstDiscoveryData := s.CreateEcstDiscoveryData(namespace, mappedDeploymentsEcst, &cluster)
		ecstData = append(ecstData, ecstDiscoveryData)
	}

	return ecstData, legacyData, nil
}

func (s scanner) LogAndShareError(message string, loglevel string, err error, id string) error {
	logger.Errorf(message, s.runId, err)
	statusErr := s.ShareStatus(id, FAILED, "Kubernetes scan failed")
	if statusErr != nil {
		logger.Errorf(StatusErrorFormat, s.runId, statusErr)
	}
	logErr := s.ShareAdminLogs(id, loglevel, fmt.Sprintf(message, s.runId, err))
	if logErr != nil {
		logger.Errorf(StatusErrorFormat, s.runId, logErr)
	}
	return err
}

func (s *scanner) CreateLegacyDiscoveryData(namespace corev1.Namespace, deployments []models.Deployment, clusterDTO *ClusterDTO) models.DiscoveryData {
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

	return models.DiscoveryData{
		Cluster: result,
	}

}

func (s *scanner) CreateEcstDiscoveryData(namespace corev1.Namespace, deployments []models.DeploymentEcst, clusterDTO *ClusterDTO) models.Data {
	result := models.ClusterEcst{
		Namespace:   namespace.Name,
		Deployments: deployments,
		Name:        clusterDTO.name,
		Os:          clusterDTO.osImage,
		K8sVersion:  clusterDTO.k8sVersion,
		NoOfNodes:   strconv.Itoa(clusterDTO.nodesCount),
	}

	return models.Data{
		Cluster: result,
	}
}

func (s *scanner) ShareStatus(configid string, status string, message string) error {
	var statusArray []StatusItem
	statusObject := NewStatusEvent(configid, s.runId, s.workspaceId, status, message)
	statusArray = append(statusArray, *statusObject)
	statusByte, err := storage.Marshal(statusArray)
	err = s.eventProducer.PostStatus(statusByte)
	if err != nil {
		logger.Debugf("Failed sharing status for RunId: [%s], with reason %v", s.runId, err)
		return err
	}
	return nil
}

func (s *scanner) ShareAdminLogs(configid string, loglevel string, message string) error {
	var statusArray []StatusItem
	statusObject := NewAdminLogEvent(configid, s.runId, s.workspaceId, loglevel, message)
	statusArray = append(statusArray, *statusObject)
	statusByte, err := storage.Marshal(statusArray)
	err = s.eventProducer.PostStatus(statusByte)
	if err != nil {
		logger.Debugf("Failed sharing admin logs for RunId: [%s], with reason %v", s.runId, err)
		return err
	}
	return nil
}
