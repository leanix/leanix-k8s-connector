package models

type Data struct {
	Workload Workload `json:"workload"`
}

type Service struct {
	Name string `json:"name"`
}

type Workload struct {
	ClusterName        string      `json:"clusterName"`
	WorkloadType       string      `json:"workloadType"`
	WorkloadName       string      `json:"workloadName"`
	Containers         Containers  `json:"containers"`
	ServiceName        string      `json:"serviceName"`
	Labels             interface{} `json:"labels"`
	Timestamp          string      `json:"timestamp"`
	WorkloadProperties Properties  `json:"WorkloadProperties"`
}

type Properties struct {
	Schedule       string `json:"schedule"`
	Replicas       string `json:"replicas"`
	UpdateStrategy string `json:"updateStrategy"`
}

type Containers struct {
	Name        string       `json:"containerName"`
	Image       string       `json:"image"`
	Port        interface{}  `json:"port"`
	K8sLimits   K8sResources `json:"k8sLimits"`
	K8sRequests K8sResources `json:"k8sRequests"`
}

type K8sResources struct {
	Cpu    string `json:"cpu"`
	Memory string `json:"memory"`
}
