package iris

import (
	"encoding/json"
	"fmt"
	"github.com/leanix/leanix-k8s-connector/pkg/iris/models/namespace"
	newmapper "github.com/leanix/leanix-k8s-connector/pkg/newmapper/workloadEcst"
	"github.com/leanix/leanix-k8s-connector/pkg/utils"
	"github.com/spf13/viper"
	"net/http"
	"strconv"

	workloads "github.com/leanix/leanix-k8s-connector/pkg/iris/models/workload"
	"github.com/leanix/leanix-k8s-connector/pkg/kubernetes"
	"github.com/leanix/leanix-k8s-connector/pkg/logger"
	"github.com/leanix/leanix-k8s-connector/pkg/storage"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

type Scanner interface {
	Scan(getKubernetesApiFunc kubernetes.GetKubernetesAPI, config *rest.Config, configurationName string) error
}

type scanner struct {
	configService         ConfigService
	eventProducer         EventProducer
	workloadEventProducer WorkloadEventProducer
	runId                 string
	workspaceId           string
}

func NewScanner(kind string, uri string, runId string, token string, workspaceId string) Scanner {
	api := NewApi(http.DefaultClient, kind, uri, token)
	configService := NewConfigService(api)
	eventProducer := NewEventProducer(api, runId, workspaceId)
	workloadEventProducer := NewEventWorkloadProducer(api, runId, workspaceId)
	return &scanner{
		configService:         configService,
		eventProducer:         eventProducer,
		workloadEventProducer: workloadEventProducer,
		runId:                 runId,
		workspaceId:           workspaceId,
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

	if viper.GetBool(utils.WorkloadFlag) {
		logger.Info("Scanning of Workloads is enabled")
		mapper := newmapper.MapWorkload(kubernetesAPI, kubernetesConfig.Cluster, s.workspaceId, s.runId)
		ecstWorkloadDiscoveredData, err := s.ScanWorkloads(kubernetesAPI, mapper, kubernetesConfig.Cluster)
		if err != nil {
			return s.LogAndShareError("Scan failed while retrieving k8s workload. RunId: [%s], with reason: '%v'", ERROR, err, kubernetesConfig.ID)
		}
		oldResults, err := s.configService.GetWorkloadScanResults(kubernetesConfig.ID)
		if err != nil {
			return err
		}
		err = s.workloadEventProducer.ProcessWorkloads(ecstWorkloadDiscoveredData, oldResults, kubernetesConfig.ID)
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
	oldResults, err := s.configService.GetNamespaceScanResults(kubernetesConfig.ID)
	if err != nil {
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
	ecstDiscoveredData, err := s.ScanNamespace(kubernetesAPI, mapper, namespaces.Items, clusterDTO)
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

func (s scanner) ScanNamespace(k8sApi *kubernetes.API, mapper Mapper, namespaces []corev1.Namespace, cluster ClusterDTO) ([]models.Data, error) {
	var ecstData = make([]models.Data, 0)

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

		// create ECST discovery item for namespace
		ecstDiscoveryData := s.CreateEcstDiscoveryData(namespace, mappedDeploymentsEcst, &cluster)
		ecstData = append(ecstData, ecstDiscoveryData)
	}

	return ecstData, nil
}

// ScanWorkloads todo: scanWorkload function
func (s scanner) ScanWorkloads(k8sApi *kubernetes.API, newmapper newmapper.MapperWorkload, clusterName string) ([]workloads.Data, error) {
	var ecstData = make([]workloads.Data, 0)

	mappedWorkloadEcst, err := newmapper.MapWorkloads(k8sApi, clusterName)
	if err != nil {
		return nil, err
	}

	// create ECST discovery item with Data object
	ecstDiscoveryData := workloads.Data{
		Workload: mappedWorkloadEcst,
	}
	ecstData = append(ecstData, ecstDiscoveryData)
	return ecstData, nil
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
