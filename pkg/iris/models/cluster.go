package models

type Cluster struct {
	Namespace   Namespace    `json:"namespace"`
	Deployments []Deployment `json:"deployment"`
	Name        string       `json:"name"`
	Os          string       `json:"os"`
	K8sVersion  string       `json:"k8sVersion"`
	NoOfNodes   string       `json:"noOfNodes"`
}

type Namespace struct {
	Name string `json:"name"`
}

//Interface functions can go here
