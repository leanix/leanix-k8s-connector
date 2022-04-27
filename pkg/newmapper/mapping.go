package newmapper

import (
	"github.com/leanix/leanix-k8s-connector/pkg/kubernetes"
	"github.com/leanix/leanix-k8s-connector/pkg/mapper"
	"k8s.io/client-go/rest"
)

func ScanKubernetes(clusterName string, config *rest.Config) ([]mapper.KubernetesObject, error) {
	kubernetesAPI, err := kubernetes.NewAPI(config)
	if err != nil {
		return nil, err
	}
	var blacklistedNamespacesList []string
	var scannedObjects []mapper.KubernetesObject
	namespaces, err := kubernetesAPI.Namespaces(blacklistedNamespacesList)
	if err != nil {
		return nil, err
	}
	cluster, err := GetCluster(clusterName, kubernetesAPI)
	if err != nil {
		return nil, err
	}
	deployments, err := GetDeployments(clusterName, namespaces, kubernetesAPI)
	if err != nil {
		return nil, err
	}
	cronJobs, err := GetCronJobs(clusterName, namespaces, kubernetesAPI)
	if err != nil {
		return nil, err
	}
	statefulSets, err := GetStatefulSets(clusterName, namespaces, kubernetesAPI)
	if err != nil {
		return nil, err
	}

	scannedObjects = append(scannedObjects, *cluster)
	scannedObjects = append(scannedObjects, deployments...)
	scannedObjects = append(scannedObjects, cronJobs...)
	scannedObjects = append(scannedObjects, statefulSets...)
	return scannedObjects, nil
}
