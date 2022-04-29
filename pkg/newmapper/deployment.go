package newmapper

import (
	"strings"
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
		deploymentArtifact, err := DeploymentSoftwareArtifact(deployment)
		mappedDeployment, err := DeploymentDataMapping(deployment)
		if err != nil {
			return nil, err
		}
		groupDeployments = append(groupDeployments, *mappedDeployment, *deploymentArtifact)
	}
	return groupDeployments, nil
}

func DeploymentSoftwareArtifact(Deployment appsvs.Deployment) (*mapper.KubernetesObject, error) {
	var DeploymentData map[string]interface{}
	DeploymentData = make(map[string]interface{})
	DeploymentId := Deployment.Namespace + "_" + Deployment.Name
	DeploymentData["clusterName"] = Deployment.ClusterName
	DeploymentData["name"] = Deployment.Namespace + ":" + Deployment.Name
	DeploymentData["category"] = "Microservice"

	return &mapper.KubernetesObject{
		ID:   DeploymentId,
		Type: "Microservice",
		Data: DeploymentData,
	}, nil
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
	deploymentImage := strings.Split(deployment.Spec.Template.Spec.Containers[0].Image, ":")[0]
	deploymentId := deployment.Namespace + ":" + deployment.Name + "-" + deploymentVersionShort + "-" + deployment.ClusterName
	deploymentData["name"] = deployment.Namespace + ":" + deployment.Name + " in " + deployment.ClusterName
	deploymentData["longName"] = deployment.Namespace + ":" + deployment.Name + " (" + deploymentVersion + ")" + " in " + deployment.ClusterName
	deploymentData["category"] = "deployment"
	deploymentData["clusterName"] = deployment.ClusterName
	deploymentData["version"] = version
	deploymentData["image"] = deploymentImage
	deploymentData["deploymentTime"] = deployment.CreationTimestamp.UTC().Format(time.RFC3339)
	deploymentData["k8sName"] = deployment.Name
	deploymentData["namespace"] = deployment.Namespace
	deploymentData["labels"] = deployment.ObjectMeta.Labels
	deploymentData["annotations"] = deployment.ObjectMeta.Annotations
	deploymentData["updateStrategy"] = deployment.Spec.Strategy.Type
	deploymentData["k8sImage"] = deploymentImage
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
