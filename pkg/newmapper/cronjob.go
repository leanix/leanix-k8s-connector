package newmapper

import (
	"strings"
	"time"

	"github.com/leanix/leanix-k8s-connector/pkg/kubernetes"
	"github.com/leanix/leanix-k8s-connector/pkg/mapper"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

func GetCronJobs(clusterName string, namespaces *corev1.NamespaceList, kubernetesAPI *kubernetes.API) ([]mapper.KubernetesObject, error) {
	var allCronJobs []mapper.KubernetesObject
	for _, namespace := range namespaces.Items {
		cronJobs, err := kubernetesAPI.CronJobs(namespace.Name)
		if err != nil {
			return nil, err
		}
		mappedCronJobs, err := MapCronJobs(clusterName, cronJobs)
		if err != nil {
			return nil, err
		}
		allCronJobs = append(allCronJobs, mappedCronJobs...)

	}
	return allCronJobs, nil
}

func MapCronJobs(clusterName string, cronJobs *batchv1.CronJobList) ([]mapper.KubernetesObject, error) {
	var groupedCronJobs []mapper.KubernetesObject
	for _, cronJob := range cronJobs.Items {
		cronJob.ClusterName = clusterName
		cronJobArtifact, err := CronJobSoftwareArtifact(cronJob)
		if err != nil {
			return nil, err
		}
		mappedCronJob, err := CronJobDataMapping(cronJob)
		if err != nil {
			return nil, err
		}
		groupedCronJobs = append(groupedCronJobs, *mappedCronJob, *cronJobArtifact)
	}
	return groupedCronJobs, nil
}

func CronJobSoftwareArtifact(CronJob batchv1.CronJob) (*mapper.KubernetesObject, error) {
	var CronJobData map[string]interface{}
	CronJobData = make(map[string]interface{})
	CronJobId := CronJob.Namespace + "_" + CronJob.Name
	CronJobData["clusterName"] = CronJob.ClusterName
	CronJobData["name"] = CronJob.Namespace + ":" + CronJob.Name
	CronJobData["category"] = "Microservice"

	return &mapper.KubernetesObject{
		ID:   CronJobId,
		Type: "Microservice",
		Data: CronJobData,
	}, nil
}

//create a data object that contains name, labels, CronJobTime, namespace, version and image of the CronJob and returns as KubernetesObject
func CronJobDataMapping(CronJob batchv1.CronJob) (*mapper.KubernetesObject, error) {
	var CronJobData map[string]interface{}
	CronJobData = make(map[string]interface{})
	var version string
	var CronJobVersion string
	var CronJobVersionShort string
	if _, ok := CronJob.ObjectMeta.Labels["app.kubernetes.io/version"]; ok {
		CronJobVersion = CronJob.ObjectMeta.Labels["app.kubernetes.io/version"]
		CronJobVersionShort = CronJob.ObjectMeta.Labels["app.kubernetes.io/version"]
	} else {
		version = ""
		CronJobVersion = "unknown version"
		CronJobVersionShort = "unknown"
	}
	CronJobImage := strings.Split(CronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Image, ":")[0]
	CronJobId := CronJob.Namespace + ":" + CronJob.Name + "-" + CronJobVersionShort + "-" + CronJob.ClusterName
	CronJobData["name"] = CronJob.Namespace + ":" + CronJob.Name + " in " + CronJob.ClusterName
	CronJobData["longName"] = CronJob.Namespace + ":" + CronJob.Name + " (" + CronJobVersion + ")" + " in " + CronJob.ClusterName
	CronJobData["category"] = "CronJob"
	CronJobData["clusterName"] = CronJob.ClusterName
	CronJobData["version"] = version
	CronJobData["image"] = CronJobImage
	CronJobData["CronJobTime"] = CronJob.CreationTimestamp.UTC().Format(time.RFC3339)
	CronJobData["k8sName"] = CronJob.Name
	CronJobData["namespace"] = CronJob.Namespace
	CronJobData["labels"] = CronJob.ObjectMeta.Labels
	CronJobData["annotations"] = CronJob.ObjectMeta.Annotations
	CronJobData["k8sImage"] = CronJobImage
	CronJobData["limits"] = CronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Resources.Limits
	CronJobData["requests"] = CronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Resources.Requests
	CronJobData["replicas"] = CronJob.Status
	CronJobData["readyReplicas"] = CronJob.Status
	CronJobData["softwareArtifact"] = CronJob.Namespace + "_" + CronJob.Name
	return &mapper.KubernetesObject{
		ID:   CronJobId,
		Type: "CronJob",
		Data: CronJobData,
	}, nil
}
