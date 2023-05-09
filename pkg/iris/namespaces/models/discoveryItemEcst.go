package models

type Data struct {
	Cluster ClusterEcst `json:"cluster"`
}

type ClusterEcst struct {
	Namespace   string           `json:"namespaceName"`
	Deployments []DeploymentEcst `json:"deployments"`
	Name        string           `json:"clusterName"`
	Os          string           `json:"os"`
	K8sVersion  string           `json:"k8sVersion"`
	NoOfNodes   string           `json:"noOfNodes"`
}
