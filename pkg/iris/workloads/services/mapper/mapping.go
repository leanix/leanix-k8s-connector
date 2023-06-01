package mapper

import (
	workload "github.com/leanix/leanix-k8s-connector/pkg/iris/workloads/models"
	"github.com/leanix/leanix-k8s-connector/pkg/kubernetes"
	"github.com/leanix/leanix-k8s-connector/pkg/set"
	v1 "k8s.io/api/core/v1"
	"strings"
)

type WorkloadMapper interface {
	MapCluster(clusterName string, nodes *v1.NodeList) (workload.Cluster, error)
	MapWorkloads(cluster workload.Cluster) ([]workload.Data, error)
}

type mapperWorkload struct {
	KubernetesApi *kubernetes.API
	ClusterName   string
	WorkspaceId   string
	runId         string
}

func NewMapper(
	kubernetesApi *kubernetes.API,
	clusterName string,
	workspaceId string,
	runId string) WorkloadMapper {
	return &mapperWorkload{
		KubernetesApi: kubernetesApi,
		ClusterName:   clusterName,
		WorkspaceId:   workspaceId,
		runId:         runId,
	}
}

func (m *mapperWorkload) MapWorkloads(cluster workload.Cluster) ([]workload.Data, error) {

	var scannedWorkloads []workload.Data
	services, err := m.KubernetesApi.Services("")
	if err != nil {
		return nil, err
	}

	deployments, err := m.KubernetesApi.Deployments("")
	if err != nil {
		return nil, err
	}
	mappedDeployments, err := m.MapDeploymentsEcst(cluster, deployments, services)
	if err != nil {
		return nil, err
	}

	cronJobs, err := m.KubernetesApi.CronJobs("")
	if err != nil {
		return nil, err
	}
	mappedCronJobs, err := m.MapCronJobsEcst(cluster, cronJobs, services)
	if err != nil {
		return nil, err
	}

	statefulSets, err := m.KubernetesApi.StatefulSets("")
	if err != nil {
		return nil, err
	}
	MappedStatefulSets, err := m.MapStatefulSetsEcst(cluster, statefulSets, services)
	if err != nil {
		return nil, err
	}

	daemonSets, err := m.KubernetesApi.DaemonSets("")
	if err != nil {
		return nil, err
	}
	MappedDaemonSets, err := m.MapDaemonSetsEcst(cluster, daemonSets, services)
	if err != nil {
		return nil, err
	}

	scannedWorkloads = append(scannedWorkloads, mappedDeployments...)
	scannedWorkloads = append(scannedWorkloads, mappedCronJobs...)
	scannedWorkloads = append(scannedWorkloads, MappedStatefulSets...)
	scannedWorkloads = append(scannedWorkloads, MappedDaemonSets...)
	return scannedWorkloads, nil
}

func (m *mapperWorkload) MapCluster(clusterName string, nodes *v1.NodeList) (workload.Cluster, error) {
	items := nodes.Items
	if len(items) == 0 {
		return workload.Cluster{
			Name: clusterName,
		}, nil
	}
	os := set.NewStringSet()
	k8sVersion := set.NewStringSet()

	for _, n := range items {
		os.Add(n.Status.NodeInfo.OSImage)
		k8sVersion.Add(n.Status.NodeInfo.KubeletVersion)
	}
	return workload.Cluster{
		Name:       clusterName,
		OsImage:    strings.Join(os.Items(), ", "),
		K8sVersion: strings.Join(k8sVersion.Items(), ", "),
		NoOfNodes:  len(items),
	}, nil
}
