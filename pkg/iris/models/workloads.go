package models

// workloads

type Deployment struct {
	Service    Service     `json:"service"`
	Name       string      `json:"name"`
	Labels     interface{} `json:"labels"`
	Timestamp  string      `json:"time"`
	Properties `json:"properties"`
}

type Service struct {
	Name string `json:"name"`
}

// properties  for workloads

type Properties struct {
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
