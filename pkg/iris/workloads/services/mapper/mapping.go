package mapper

import (
	"github.com/leanix/leanix-k8s-connector/pkg/iris/workloads/models"
	"github.com/leanix/leanix-k8s-connector/pkg/kubernetes"
)

type MapperWorkload interface {
	MapWorkloads(k8sApi *kubernetes.API, clusterName string) ([]models.Workload, error)
}

type mapworkload struct {
	KubernetesApi *kubernetes.API
	ClusterName   string
	WorkspaceId   string
	runId         string
}

func NewMapper(
	kubernetesApi *kubernetes.API,
	clusterName string,
	workspaceId string,
	runId string) MapperWorkload {
	return &mapworkload{
		KubernetesApi: kubernetesApi,
		ClusterName:   clusterName,
		WorkspaceId:   workspaceId,
		runId:         runId,
	}
}

func (m *mapworkload) MapWorkloads(k8sApi *kubernetes.API, clusterName string) ([]models.Workload, error) {

	var scannedWorkloads []models.Workload
	services, err := k8sApi.Services("")
	if err != nil {
		return nil, err
	}

	deployments, err := k8sApi.Deployments("")
	if err != nil {
		return nil, err
	}
	mappedDeployments, err := m.MapDeploymentsEcst(clusterName, deployments, services)
	if err != nil {
		return nil, err
	}

	cronJobs, err := k8sApi.CronJobs("")
	if err != nil {
		return nil, err
	}
	mappedCronJobs, err := m.MapCronJobsEcst(clusterName, cronJobs, services)
	if err != nil {
		return nil, err
	}

	statefulSets, err := k8sApi.StatefulSets("")
	if err != nil {
		return nil, err
	}
	MappedStatefulSets, err := m.MapStatefulSetsEcst(clusterName, statefulSets, services)
	if err != nil {
		return nil, err
	}

	daemonSets, err := k8sApi.DaemonSets("")
	if err != nil {
		return nil, err
	}
	MappedDaemonSets, err := m.MapDaemonSetsEcst(clusterName, daemonSets, services)
	if err != nil {
		return nil, err
	}

	scannedWorkloads = append(scannedWorkloads, mappedDeployments...)
	scannedWorkloads = append(scannedWorkloads, mappedCronJobs...)
	scannedWorkloads = append(scannedWorkloads, MappedStatefulSets...)
	scannedWorkloads = append(scannedWorkloads, MappedDaemonSets...)
	return scannedWorkloads, nil
}
