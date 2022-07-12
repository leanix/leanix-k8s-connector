package iris

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/leanix/leanix-k8s-connector/pkg/iris/models"
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

	id := fmt.Sprintf("%s:%s-%s", deployment.Namespace, deployment.Name, m.ClusterName)
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

type EventBuilder interface {
	Cluster(n *models.Namespace, d *models.Deployment) EventBuilder
	Build() models.DiscoveryItem
}

type eventBuilder struct {
	c models.Cluster
	d models.Deployment
	s models.Service
	p models.Properties
}

func New() EventBuilder {
	return &eventBuilder{}
}

func (eb *eventBuilder) Cluster(cluster models.Cluster, de models.Deployment) EventBuilder {
	eb.c = cluster
	eb.d = de
	return eb
}

func (eb *eventBuilder) Build() *models.DiscoveryItem {
	data := &models.Data{
		Cluster: eb.c,
	}
	return &models.DiscoveryItem{
		Data: *data,
	}
}
