package newmapper

import (
	"strings"
	"time"

	"github.com/leanix/leanix-k8s-connector/pkg/kubernetes"
	"github.com/leanix/leanix-k8s-connector/pkg/mapper"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func GetStatefulSets(clusterName string, namespaces *corev1.NamespaceList, kubernetesAPI *kubernetes.API) ([]mapper.KubernetesObject, error) {
	var allStatefulSets []mapper.KubernetesObject
	for _, namespace := range namespaces.Items {
		StatefulSets, err := kubernetesAPI.StatefulSets(namespace.Name)
		if err != nil {
			return nil, err
		}
		mappedStatefulSets, err := MapStatefulSets(clusterName, StatefulSets)
		if err != nil {
			return nil, err
		}
		allStatefulSets = append(allStatefulSets, mappedStatefulSets...)

	}
	return allStatefulSets, nil
}

func MapStatefulSets(clusterName string, StatefulSets *appsv1.StatefulSetList) ([]mapper.KubernetesObject, error) {
	var groupedStatefulSets []mapper.KubernetesObject
	for _, StatefulSet := range StatefulSets.Items {
		StatefulSet.ClusterName = clusterName
		StatefulSetArtifact, err := StatefulSetSoftwareArtifact(StatefulSet)
		if err != nil {
			return nil, err
		}
		mappedStatefulSet, err := StatefulSetDataMapping(StatefulSet)
		if err != nil {
			return nil, err
		}
		groupedStatefulSets = append(groupedStatefulSets, *mappedStatefulSet, *StatefulSetArtifact)
	}
	return groupedStatefulSets, nil
}

func StatefulSetSoftwareArtifact(StatefulSet appsv1.StatefulSet) (*mapper.KubernetesObject, error) {
	var StatefulSetData map[string]interface{}
	StatefulSetData = make(map[string]interface{})
	StatefulSetId := StatefulSet.Namespace + "_" + StatefulSet.Name
	StatefulSetData["clusterName"] = StatefulSet.ClusterName
	StatefulSetData["name"] = StatefulSet.Namespace + ":" + StatefulSet.Name
	StatefulSetData["category"] = "Microservice"

	return &mapper.KubernetesObject{
		ID:   StatefulSetId,
		Type: "Microservice",
		Data: StatefulSetData,
	}, nil
}

//create a data object that contains name, labels, StatefulSetime, namespace, version and image of the StatefulSet and returns as KubernetesObject
func StatefulSetDataMapping(StatefulSet appsv1.StatefulSet) (*mapper.KubernetesObject, error) {
	var StatefulSetData map[string]interface{}
	StatefulSetData = make(map[string]interface{})
	var version string
	var StatefulSetVersion string
	var StatefulSetVersionShort string
	if _, ok := StatefulSet.ObjectMeta.Labels["app.kubernetes.io/version"]; ok {
		StatefulSetVersion = StatefulSet.ObjectMeta.Labels["app.kubernetes.io/version"]
		StatefulSetVersionShort = StatefulSet.ObjectMeta.Labels["app.kubernetes.io/version"]
	} else {
		version = ""
		StatefulSetVersion = "unknown version"
		StatefulSetVersionShort = "unknown"
	}
	StatefulSetImage := strings.Split(StatefulSet.Spec.Template.Spec.Containers[0].Image, ":")[0]
	StatefulSetId := StatefulSet.Namespace + ":" + StatefulSet.Name + "-" + StatefulSetVersionShort + "-" + StatefulSet.ClusterName
	StatefulSetData["name"] = StatefulSet.Namespace + ":" + StatefulSet.Name + " in " + StatefulSet.ClusterName
	StatefulSetData["longName"] = StatefulSet.Namespace + ":" + StatefulSet.Name + " (" + StatefulSetVersion + ")" + " in " + StatefulSet.ClusterName
	StatefulSetData["category"] = "StatefulSet"
	StatefulSetData["clusterName"] = StatefulSet.ClusterName
	StatefulSetData["version"] = version
	StatefulSetData["image"] = StatefulSetImage
	StatefulSetData["StatefulSetTime"] = StatefulSet.CreationTimestamp.UTC().Format(time.RFC3339)
	StatefulSetData["k8sName"] = StatefulSet.Name
	StatefulSetData["namespace"] = StatefulSet.Namespace
	StatefulSetData["labels"] = StatefulSet.ObjectMeta.Labels
	StatefulSetData["annotations"] = StatefulSet.ObjectMeta.Annotations
	StatefulSetData["k8sImage"] = StatefulSetImage
	StatefulSetData["limits"] = StatefulSet.Spec.Template.Spec.Containers[0].Resources.Limits
	StatefulSetData["requests"] = StatefulSet.Spec.Template.Spec.Containers[0].Resources.Requests
	StatefulSetData["replicas"] = StatefulSet.Status
	StatefulSetData["readyReplicas"] = StatefulSet.Status
	return &mapper.KubernetesObject{
		ID:   StatefulSetId,
		Type: "StatefulSet",
		Data: StatefulSetData,
	}, nil
}
