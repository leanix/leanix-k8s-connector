package iris

import (
	"fmt"

	"github.com/google/uuid"
	appsv1 "k8s.io/api/apps/v1"
)

const (
	typeAsK8sRuntime string = "leanix.vsm.item-discovered.runtimeObject"
	typeAsK8sService string = "leanix.vsm.item-discovered.kubernetesService"
)

type DiscoveryItem struct {
	ID      string      `json:"id"`
	Scope   string      `json:"scope"`
	Type    string      `json:"type"`
	Source  string      `json:"source"`
	Time    string      `json:"time"`
	Subject string      `json:"subject"`
	Data    interface{} `json:"data"`
}

func NewDeploymentEvent(m mapper, deployment appsv1.Deployment) *DiscoveryItem {

	deployment.ClusterName = m.ClusterName
	id := fmt.Sprintf("%s:%s-%s", deployment.Namespace, deployment.Name, deployment.ClusterName)
	Subject := fmt.Sprintf("deployment/%s", deployment.Name)
	scope := fmt.Sprintf("workspace/%s", m.WorkspaceId)
	Source := fmt.Sprintf("kubernetes/%s#%s", m.ClusterName, m.runId)
	Time := deployment.CreationTimestamp.String()
	return &DiscoveryItem{
		ID:      id,
		Scope:   scope,
		Type:    typeAsK8sRuntime,
		Source:  Source,
		Time:    Time,
		Subject: Subject,
		Data:    deployment,
	}
}

func NewSoftwareArtifactEvent(m mapper, deployment appsv1.Deployment) *DiscoveryItem {

	var deploymentData = make(map[string]interface{})

	deploymentData["clusterName"] = m.ClusterName
	deploymentData["name"] = deployment.Namespace + ":" + deployment.Name
	deploymentData["type"] = "namespaceBased"

	deployment.ClusterName = m.ClusterName
	id := fmt.Sprintf("%s-%s", deployment.Namespace, deployment.Name)
	Subject := fmt.Sprintf("softwareArtifact/%s", deployment.Name)
	scope := fmt.Sprintf("workspace/%s", m.WorkspaceId)
	Source := fmt.Sprintf("kubernetes/%s#%s", m.ClusterName, m.runId)
	Time := deployment.CreationTimestamp.String()
	return &DiscoveryItem{
		ID:      id,
		Scope:   scope,
		Type:    typeAsK8sService,
		Source:  Source,
		Time:    Time,
		Subject: Subject,
		Data:    deploymentData,
	}
}

// struct to extend Log with RunId
type Log struct {
	// Root        *Logger
	// Output      *Output
	RunId       string
	WorkspaceId string
}

func GenerateRunId() string {
	return uuid.New().String()
}
