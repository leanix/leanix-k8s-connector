package iris

import (
	"encoding/json"
	"fmt"
	"github.com/leanix/leanix-k8s-connector/pkg/iris/common/models"
	"github.com/leanix/leanix-k8s-connector/pkg/iris/common/services"
	namespaceModels "github.com/leanix/leanix-k8s-connector/pkg/iris/namespaces/models"
	"github.com/leanix/leanix-k8s-connector/pkg/iris/namespaces/services/events"
	namespaceMap "github.com/leanix/leanix-k8s-connector/pkg/iris/namespaces/services/mapper"
	workload "github.com/leanix/leanix-k8s-connector/pkg/iris/workloads/models"
	workloadService "github.com/leanix/leanix-k8s-connector/pkg/iris/workloads/services/events"
	workloadMap "github.com/leanix/leanix-k8s-connector/pkg/iris/workloads/services/mapper"
	"net/http"
	"strconv"

	"github.com/leanix/leanix-k8s-connector/pkg/kubernetes"
	"github.com/leanix/leanix-k8s-connector/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

type Scanner interface {
	Scan(getKubernetesApiFunc kubernetes.GetKubernetesAPI, config *rest.Config, configurationName string) error
}

type scanner struct {
	configService         services.ConfigService
	eventProducer         events.EventProducer
	workloadEventProducer workloadService.WorkloadEventProducer
	runId                 string
	workspaceId           string
}

func NewScanner(kind string, uri string, runId string, token string, workspaceId string) Scanner {
	api := services.NewIrisApi(http.DefaultClient, kind, uri, token)
	configService := services.NewConfigService(api)
	eventProducer := events.NewEventProducer(api, runId, workspaceId)
	workloadEventProducer := workloadService.NewEventWorkloadProducer(api, runId, workspaceId)
	return &scanner{
		configService:         configService,
		eventProducer:         eventProducer,
		workloadEventProducer: workloadEventProducer,
		runId:                 runId,
		workspaceId:           workspaceId,
	}
}

const (
	IN_PROGRESS        string = "IN_PROGRESS"
	FAILED             string = "FAILED"
	SUCCESSFUL         string = "SUCCESSFUL"
	SUCCESSFUL_WARNING string = "SUCCESSFUL_WARNING"
	ERROR              string = "ERROR"
	WARNING            string = "WARNING"
	INFO               string = "INFO"
	WORKLOAD           string = "WORKLOAD"
)

const StatusErrorFormat = "Scan failed while posting status. RunId: [%s], with reason: '%v'"

func (s *scanner) Scan(getKubernetesAPI kubernetes.GetKubernetesAPI, config *rest.Config, configurationName string) error {
	configuration, err := s.configService.GetConfiguration(configurationName)
	if err != nil {
		return err
	}
	kubernetesConfig := models.KubernetesConfig{}
	err = json.Unmarshal(configuration, &kubernetesConfig)
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

	if kubernetesConfig.DiscoveryMode == WORKLOAD {
		logger.Info("Scanning of Workloads is enabled")
		return s.ScanWorkloads(kubernetesAPI, kubernetesConfig)
	}

	return s.ScanNamespaces(kubernetesConfig, kubernetesAPI)
}

func (s *scanner) ScanNamespaces(kubernetesConfig models.KubernetesConfig, kubernetesAPI *kubernetes.API) error {
	oldResults, err := s.configService.GetScanResults(kubernetesConfig.ID)
	if err != nil {
		return err
	}

	mapper := namespaceMap.NewMapper(kubernetesAPI, kubernetesConfig.Cluster, s.workspaceId, kubernetesConfig.BlackListedNamespaces, s.runId)

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
	ecstDiscoveredData, err := s.ProcessNamespace(kubernetesAPI, mapper, namespaces.Items, clusterDTO)
	if err != nil {
		return s.LogAndShareError("Scan failed while retrieving k8s deployments. RunId: [%s], with reason: '%v'", ERROR, err, kubernetesConfig.ID)
	}

	err = s.eventProducer.ProcessResults(ecstDiscoveredData, oldResults, kubernetesConfig.ID)
	if err != nil {
		return s.LogAndShareError("Scan failed while posting ECST results. RunId: [%s], with reason: '%v'", ERROR, err, kubernetesConfig.ID)
	}

	logger.Infof("Scan Finished for RunId: [%s]", s.runId)
	err = s.ShareStatus(kubernetesConfig.ID, SUCCESSFUL, "Successfully Scanned")
	if err != nil {
		logger.Errorf(StatusErrorFormat, s.runId, err)
		return err
	}
	return err
}

func (s *scanner) ScanWorkloads(kubernetesAPI *kubernetes.API, kubernetesConfig models.KubernetesConfig) error {
	mapper := workloadMap.NewMapper(kubernetesAPI, kubernetesConfig.Cluster, s.workspaceId, s.runId)

	nodes, err := kubernetesAPI.Nodes()
	if err != nil {
		return s.LogAndShareError("Scan failed while retrieving k8s cluster nodes. RunId: [%s], with reason: '%v'", ERROR, err, kubernetesConfig.ID)
	}

	clusterInfo, err := mapper.MapCluster(kubernetesConfig.Cluster, nodes)
	if err != nil {
		return s.LogAndShareError("Scan failed while aggregating cluster information. RunId: [%s], with reason: '%v'", ERROR, err, kubernetesConfig.ID)
	}

	discoveredWorkloads, err := s.ProcessWorkloads(mapper, clusterInfo)
	if err != nil {
		return s.LogAndShareError("Scan failed while retrieving k8s workload. RunId: [%s], with reason: '%v'", ERROR, err, kubernetesConfig.ID)
	}
	oldResults, err := s.configService.GetScanResults(kubernetesConfig.ID)
	if err != nil {
		return err
	}
	err = s.workloadEventProducer.ProcessWorkloads(discoveredWorkloads, oldResults, kubernetesConfig.ID)
	if err != nil {
		return s.LogAndShareError("Scan failed while posting ECST results. RunId: [%s], with reason: '%v'", ERROR, err, kubernetesConfig.ID)
	}
	logger.Infof("Scan Finished for RunId: [%s]", s.runId)
	err = s.ShareStatus(kubernetesConfig.ID, SUCCESSFUL, "Successfully Scanned")
	if err != nil {
		logger.Errorf(StatusErrorFormat, s.runId, err)
		return err
	}
	return err
}

func (s *scanner) ProcessNamespace(k8sApi *kubernetes.API, mapper namespaceMap.Mapper, namespaces []corev1.Namespace, cluster namespaceMap.ClusterDTO) ([]namespaceModels.Data, error) {
	var ecstData = make([]namespaceModels.Data, 0)

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

		mappedDeploymentsEcst, err := mapper.MapDeploymentsEcst(deployments, services)
		if err != nil {
			return nil, err
		}

		// create ECST discovery item for namespaceModels
		ecstDiscoveryData := s.CreateEcstDiscoveryData(namespace, mappedDeploymentsEcst, &cluster)
		ecstData = append(ecstData, ecstDiscoveryData)
	}

	return ecstData, nil
}

func (s *scanner) ProcessWorkloads(mapper workloadMap.WorkloadMapper, clusterInfo workload.Cluster) ([]workload.Data, error) {
	return mapper.MapWorkloads(clusterInfo)
}

func (s *scanner) LogAndShareError(message string, loglevel string, err error, id string) error {
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

func (s *scanner) CreateEcstDiscoveryData(currentNamespace corev1.Namespace, deployments []namespaceModels.DeploymentEcst, clusterDTO *namespaceMap.ClusterDTO) namespaceModels.Data {
	result := namespaceModels.ClusterEcst{
		Namespace:   currentNamespace.Name,
		Deployments: deployments,
		Name:        clusterDTO.Name,
		Os:          clusterDTO.OsImage,
		K8sVersion:  clusterDTO.K8sVersion,
		NoOfNodes:   strconv.Itoa(clusterDTO.NodesCount),
	}

	return namespaceModels.Data{
		Cluster: result,
	}
}

func (s *scanner) ShareStatus(configid string, status string, message string) error {
	var statusArray []models.StatusItem
	statusObject := models.NewStatusEvent(configid, s.runId, s.workspaceId, status, message)
	statusArray = append(statusArray, *statusObject)
	statusByte, err := json.Marshal(statusArray)
	err = s.eventProducer.PostStatus(statusByte)
	if err != nil {
		logger.Debugf("Failed sharing status for RunId: [%s], with reason %v", s.runId, err)
		return err
	}
	return nil
}

func (s *scanner) ShareAdminLogs(configId string, loglevel string, message string) error {
	var statusArray []models.StatusItem
	statusObject := models.NewAdminLogEvent(configId, s.runId, s.workspaceId, loglevel, message)
	statusArray = append(statusArray, *statusObject)
	statusByte, err := json.Marshal(statusArray)
	err = s.eventProducer.PostStatus(statusByte)
	if err != nil {
		logger.Debugf("Failed sharing admin logs for RunId: [%s], with reason %v", s.runId, err)
		return err
	}
	return nil
}
