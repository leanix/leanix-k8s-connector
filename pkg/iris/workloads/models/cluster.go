package models

type Cluster struct {
	Type           string       `json:"workloadType"`
	Workload       []Workload   `json:"workload"`
	ClusterName    string       `json:"clusterName"`
	Os             string       `json:"os"`
	K8sVersion     string       `json:"k8sVersion"`
	Containers     []Containers `json:"containers"` //todo: check if Container is required
	Schedule       string       `json:"schedule"`
	Replicas       string       `json:"replicas"`
	UpdateStrategy string       `json:"updateStrategy"`
	Labels         string       `json:"labels"`
	Timestamp      string       `json:"timestamp"`
}
