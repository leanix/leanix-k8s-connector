package newmapper

import (
	"github.com/leanix/leanix-k8s-connector/pkg/iris/models/workload"
	batchv1 "k8s.io/api/batch/v1"
	_ "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"reflect"
	"strings"
	"time"
)

func (m *mapworkload) MapCronJobsEcst(clusterName string, cronJobs *batchv1.CronJobList, services *v1.ServiceList) ([]workload.WorkloadEcst, error) {
	var groupedCronJobs []workload.WorkloadEcst

	for _, cronJob := range cronJobs.Items {
		// Check if any service has the exact same selector labels and use this as the service related to the deployment
		cronJobService := ResolveK8sServiceForK8sCronJob(services, cronJob)
		mappedCronJob := m.CreateCronjobEcst(clusterName, cronJob, cronJobService)
		groupedCronJobs = append(groupedCronJobs, mappedCronJob)
	}

	return groupedCronJobs, nil
}

// CreateCronjobEcst create a data object that contains name, labels, CronJobSchedule and more
func (m *mapworkload) CreateCronjobEcst(clusterName string, cronJob batchv1.CronJob, service string) workload.WorkloadEcst {
	mappedDeployment := workload.WorkloadEcst{
		ClusterName:  clusterName,
		WorkloadType: "cronjob",
		WorkloadName: cronJob.Name,
		ServiceName:  service,
		Containers: workload.Containers{
			Name:        cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Name,
			Image:       strings.Split(cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Image, ":")[0],
			Port:        cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Ports,
			K8sLimits:   CreateK8sResources(cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Resources.Limits),
			K8sRequests: CreateK8sResources(cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Resources.Requests),
		},
		WorkloadProperties: workload.Properties{
			Replicas:  cronJob.Status.String(),
			Schedule:  cronJob.Spec.Schedule,
			Labels:    cronJob.ObjectMeta.Labels,
			Timestamp: cronJob.CreationTimestamp.UTC().Format(time.RFC3339),
		},
	}
	return mappedDeployment
}

func ResolveK8sServiceForK8sCronJob(services *v1.ServiceList, cronJob batchv1.CronJob) string {
	cronJobService := ""
	for _, service := range services.Items {
		sharedLabelsCronJob := map[string]string{}
		sharedLabelsService := map[string]string{}
		for label := range service.Spec.Selector {
			if _, ok := cronJob.Spec.JobTemplate.Spec.Selector.MatchLabels[label]; ok {
				sharedLabelsCronJob[label] = cronJob.Spec.JobTemplate.Spec.Selector.MatchLabels[label]
				sharedLabelsService[label] = service.Spec.Selector[label]
			}
		}

		if len(sharedLabelsCronJob) != 0 && len(sharedLabelsService) != 0 && reflect.DeepEqual(sharedLabelsCronJob, sharedLabelsService) {
			cronJobService = service.Name
			break
		}
	}
	return cronJobService
}
