package newmapper

import (
	"time"

	"github.com/leanix/leanix-k8s-connector/pkg/kubernetes"
	"github.com/leanix/leanix-k8s-connector/pkg/mapper"
	appsvs "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func GetDeployments(clusterName string, namespaces *corev1.NamespaceList, kubernetesAPI *kubernetes.API) ([]mapper.KubernetesObject, error) {
	var allDeployments []mapper.KubernetesObject
	for _, namespace := range namespaces.Items {
		deployments, err := kubernetesAPI.Deployments(namespace.Name)
		if err != nil {
			return nil, err
		}

		if err != nil {
			return nil, err
		}
		mappedDeployments, err := MapDeployments(clusterName, deployments)
		if err != nil {
			return nil, err
		}
		allDeployments = append(allDeployments, mappedDeployments...)

	}
	return allDeployments, nil
}

func MapDeployments(clusterName string, deployments *appsvs.DeploymentList) ([]mapper.KubernetesObject, error) {
	var groupDeployments []mapper.KubernetesObject
	for _, deployment := range deployments.Items {
		deployment.ClusterName = clusterName
		mappedDeployment, err := DeploymentDataMapping(deployment)
		if err != nil {
			return nil, err
		}
		groupDeployments = append(groupDeployments, *mappedDeployment)
	}
	return groupDeployments, nil
}

//create a data object that contains name, labels, deploymentime, namespace, version and image of the deployment and returns as KubernetesObject
func DeploymentDataMapping(deployment appsvs.Deployment) (*mapper.KubernetesObject, error) {
	var deploymentData map[string]interface{}
	deploymentData = make(map[string]interface{})
	var version string
	var deploymentVersion string
	var deploymentVersionShort string
	if _, ok := deployment.ObjectMeta.Labels["app.kubernetes.io/version"]; ok {
		deploymentVersion = deployment.ObjectMeta.Labels["app.kubernetes.io/version"]
		deploymentVersionShort = deployment.ObjectMeta.Labels["app.kubernetes.io/version"]
	} else {
		version = ""
		deploymentVersion = "unknown version"
		deploymentVersionShort = "unknown"
	}
	deploymentId := deployment.ClusterName + "/" + deployment.Namespace + "/" + deployment.Name + "/" + deploymentVersionShort
	deploymentData["name"] = deployment.Namespace + ":" + deployment.Name + " (" + deploymentVersion + ")" + " in " + deployment.ClusterName
	deploymentData["category"] = "deployment"
	deploymentData["clusterName"] = deployment.ClusterName
	deploymentData["version"] = version
	deploymentData["image"] = deployment.Spec.Template.Spec.Containers[0].Image
	deploymentData["deploymentTime"] = deployment.CreationTimestamp.UTC().Format(time.RFC3339)
	deploymentData["k8sName"] = deployment.Name
	deploymentData["namespace"] = deployment.Namespace
	deploymentData["labels"] = deployment.ObjectMeta.Labels
	deploymentData["annotations"] = deployment.ObjectMeta.Annotations
	deploymentData["updateStrategy"] = deployment.Spec.Strategy.Type
	deploymentData["k8sImage"] = deployment.Spec.Template.Spec.Containers[0].Image
	deploymentData["limits"] = deployment.Spec.Template.Spec.Containers[0].Resources.Limits
	deploymentData["requests"] = deployment.Spec.Template.Spec.Containers[0].Resources.Requests
	deploymentData["replicas"] = deployment.Status.Replicas
	deploymentData["readyReplicas"] = deployment.Status.ReadyReplicas
	return &mapper.KubernetesObject{
		ID:   deploymentId,
		Type: "Deployment",
		Data: deploymentData,
	}, nil
}
