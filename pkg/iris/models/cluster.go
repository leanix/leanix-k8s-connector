package models

type Cluster struct {
	Namespace   string       `json:"namespaceName"`
	Deployments []Deployment `json:"deployments"`
	Name        string       `json:"clusterName"`
	Os          string       `json:"os"`
	K8sVersion  string       `json:"k8sVersion"`
	NoOfNodes   string       `json:"noOfNodes"`
}

//Interface functions can go here
