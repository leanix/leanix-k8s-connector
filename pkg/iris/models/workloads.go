package models

// workloads

type Deployment struct {
	Service        *Service             `json:"service"`
	Image          string               `json:"image"`
	DeploymentName string               `json:"deploymentName"`
	Labels         interface{}          `json:"labels"`
	Timestamp      string               `json:"time"`
	Properties     DeploymentProperties `json:"deploymentProperties"`
}

type Service struct {
	Name string `json:"name"`
}

type DeploymentEcst struct {
	ServiceName          string               `json:"serviceName"`
	Image                string               `json:"image"`
	DeploymentName       string               `json:"deploymentName"`
	Labels               interface{}          `json:"labels"`
	Timestamp            string               `json:"time"`
	DeploymentProperties DeploymentProperties `json:"deploymentProperties"`
}

// properties  for workloads

type DeploymentProperties struct {
	UpdateStrategy string       `json:"updateStrategy"`
	Replicas       string       `json:"replicas"`
	K8sLimits      K8sResources `json:"k8sLimits"`
	K8sRequests    K8sResources `json:"k8sRequests"`
}

type K8sResources struct {
	Cpu    string `json:"cpu"`
	Memory string `json:"memory"`
}

//Interface functions can go here
