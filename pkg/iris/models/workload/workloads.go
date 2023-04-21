package workload

// workloadEcst

type Service struct {
	Name string `json:"name"`
}

type WorkloadEcst struct {
	ClusterName        string     `json:"clusterName"`
	WorkloadType       string     `json:"workloadType"` //todo: check enum type instead of string
	WorkloadName       string     `json:"workloadName"`
	Containers         Containers `json:"containers"`
	ServiceName        string     `json:"serviceName"`
	WorkloadProperties Properties `json:"WorkloadProperties"`
}

// properties  for workloadEcst

type Properties struct {
	Schedule       string      `json:"schedule"`
	Replicas       string      `json:"replicas"`
	UpdateStrategy string      `json:"updateStrategy"`
	Labels         interface{} `json:"labels"`
	Timestamp      string      `json:"timestamp"`
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

//Interface functions can go here
