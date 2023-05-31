package models

type Data struct {
	Workload      Workload `json:"workload"`
	NamespaceName string   `json:"namespaceName"`
	ServiceName   string   `json:"serviceName"`
	Cluster       Cluster  `json:"cluster"`
	Timestamp     string   `json:"timestamp"`
}

type Workload struct {
	Name               string             `json:"name"`
	WorkloadType       string             `json:"type"`
	Labels             map[string]string  `json:"labels"`
	WorkloadProperties WorkloadProperties `json:"workloadProperties"`
}

type WorkloadProperties struct {
	Schedule       string     `json:"schedule"`
	Replicas       string     `json:"replicas"`
	UpdateStrategy string     `json:"updateStrategy"`
	Containers     Containers `json:"containers"`
}

type Containers struct {
	Name        string       `json:"name"`
	Image       string       `json:"image"`
	Port        interface{}  `json:"port"`
	K8sLimits   K8sResources `json:"k8sLimits"`
	K8sRequests K8sResources `json:"k8sRequests"`
}

type K8sResources struct {
	Cpu    string `json:"cpu"`
	Memory string `json:"memory"`
}

type Cluster struct {
	Name string `json:"name"`
	Os   string `json:"os"`
}
