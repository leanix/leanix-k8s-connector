package events

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/leanix/leanix-k8s-connector/pkg/iris/common/models"
	common "github.com/leanix/leanix-k8s-connector/pkg/iris/common/services"
	workload "github.com/leanix/leanix-k8s-connector/pkg/iris/workloads/models"
	"github.com/leanix/leanix-k8s-connector/pkg/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_eventProducer_filter_created(t *testing.T) {
	mockApi := mocks.NewIrisApi(t)
	newWorkload := map[string]workload.Data{
		"workload": {
			Workload: workload.Workload{
				ClusterName:  "testCluster1",
				WorkloadName: "testWorkload1",
				WorkloadType: "deployment",
				Containers: workload.Containers{
					Name:  "testContainer1",
					Image: "testImage1",
					Port:  "8080",
				},
				ServiceName: "serviceName1",
				Labels:      "k8s-app",
				Timestamp:   "",
				WorkloadProperties: workload.Properties{
					Schedule:       "testSchedule",
					Replicas:       "1",
					UpdateStrategy: "rollback",
				},
			},
		},
	}

	oldData := map[string]models.DiscoveryEvent{
		"testId2": {
			Body: models.DiscoveryBody{
				State: models.State{
					Name:           "",
					SourceInstance: "",
					SourceType:     "",
					Time:           "",
					Data: workload.Data{
						Workload: workload.Workload{
							ClusterName:  "testCluster1",
							WorkloadName: "testWorkload1",
							WorkloadType: "deployment",
							Containers: workload.Containers{
								Name:  "testContainer1",
								Image: "testImage1",
								Port:  "8080",
							},
							ServiceName: "serviceName1",
							Labels:      "k8s-app",
							Timestamp:   "",
							WorkloadProperties: workload.Properties{
								Schedule:       "testSchedule",
								Replicas:       "1",
								UpdateStrategy: "rollback",
							},
						}},
				},
			},
		},
	}
	//oldData map[string]models.DiscoveryEvent
	p := NewEventWorkloadProducer(mockApi, "testRunId", "testWorkspaceId")
	created, updated, _, err := p.FilterForChangedItems(newWorkload, oldData, "testConfigId")

	assert.NoError(t, err)
	assert.Empty(t, updated)
	assert.Len(t, created, 1)
	assert.Equal(t, "testCluster1", created[0].Body.State.Name)
}

func Test_eventProducer_filter_updated_no_change(t *testing.T) {
	mockApi := mocks.NewIrisApi(t)
	newData := map[string]workload.Data{
		"testId1": {
			Workload: workload.Workload{
				ClusterName:  "testCluster1",
				WorkloadName: "testWorkload1",
				WorkloadType: "deployment",
				Containers: workload.Containers{
					Name:  "testContainer1",
					Image: "testImage1",
					Port:  "8080",
				},
				ServiceName: "serviceName1",
				Labels:      "k8s-app",
				Timestamp:   "",
				WorkloadProperties: workload.Properties{
					Schedule:       "testSchedule",
					Replicas:       "1",
					UpdateStrategy: "rollback",
				},
			},
		}}
	oldData := map[string]models.DiscoveryEvent{
		"testId1": {
			Body: models.DiscoveryBody{
				State: models.State{
					Name:           "testCluster1",
					SourceInstance: "kubernetes",
					SourceType:     "cluster/testCluster1",
					Time:           "2023-05-04T11:03:50+02:00",
					Data: workload.Data{
						Workload: workload.Workload{
							ClusterName:  "testCluster1",
							WorkloadName: "testWorkload1",
							WorkloadType: "deployment",
							Containers: workload.Containers{
								Name:  "testContainer1",
								Image: "testImage1",
								Port:  "8080",
							},
							ServiceName: "serviceName1",
							Labels:      "k8s-app",
							Timestamp:   "",
							WorkloadProperties: workload.Properties{
								Schedule:       "testSchedule",
								Replicas:       "1",
								UpdateStrategy: "rollback",
							},
						},
					},
				},
			},
		},
	}

	//oldData map[string]models.DiscoveryEvent
	p := NewEventWorkloadProducer(mockApi, "testRunId", "testWorkspaceId")
	created, updated, filteredData, err := p.FilterForChangedItems(newData, oldData, "testConfigId")

	assert.NoError(t, err)
	assert.Empty(t, updated)
	assert.Empty(t, created)
	assert.Empty(t, filteredData)
}

func Test_eventProducer_filter_updated_changed(t *testing.T) {
	mockApi := mocks.NewIrisApi(t)
	newData := map[string]workload.Data{
		"testId1": {
			Workload: workload.Workload{
				ClusterName:  "testClusterName1",
				WorkloadName: "testWorkload1",
				WorkloadType: "deployment",
				Containers: workload.Containers{
					Name:  "testContainer1",
					Image: "testImage1",
					Port:  "8080",
				},
				ServiceName: "serviceName1",
				Labels:      "k8s-app",
				Timestamp:   "",
				WorkloadProperties: workload.Properties{
					Schedule:       "testSchedule",
					Replicas:       "1",
					UpdateStrategy: "rollback",
				},
			}},
	}
	oldData := map[string]models.DiscoveryEvent{
		"testId1": {
			Body: models.DiscoveryBody{
				State: models.State{
					Name:           "",
					SourceInstance: "",
					SourceType:     "",
					Time:           "",
					Data: workload.Data{
						Workload: workload.Workload{
							ClusterName:  "testCluster1",
							WorkloadName: "testWorkload1",
							WorkloadType: "deployment",
							Containers: workload.Containers{
								Name:  "testContainer1",
								Image: "testImage1",
								Port:  "8080",
							},
							ServiceName: "serviceName1",
							Labels:      "k8s-app",
							Timestamp:   "",
							WorkloadProperties: workload.Properties{
								Schedule:       "testSchedule",
								Replicas:       "1",
								UpdateStrategy: "rollback",
							},
						}},
				},
			},
		},
	}
	//oldData map[string]models.DiscoveryEvent
	p := &workloadEventProducer{
		irisApi:     mockApi,
		runId:       "testRunId",
		workspaceId: "testWorkspaceId",
	}
	created, updated, filteredData, err := p.FilterForChangedItems(newData, oldData, "testConfigId")

	assert.NoError(t, err)
	parsedData, err := common.ParseWorkloadData(updated[0])
	assert.NoError(t, err)
	assert.Empty(t, created)
	assert.Len(t, updated, 1)
	assert.Equal(t, "testImage1", parsedData.Workload.Containers.Image)
	assert.Empty(t, filteredData)
}

func Test_eventProducer_createECSTEvents(t *testing.T) {
	mockApi := mocks.NewIrisApi(t)
	id1 := sha256.Sum256([]byte(fmt.Sprintf("%s/%s/%s/%s", "workspace/testWorkspaceId/configuration/testConfigId", models.EventClassWorkload, "testCluster1", "testWorkload1")))
	id2 := sha256.Sum256([]byte(fmt.Sprintf("%s/%s/%s/%s", "workspace/testWorkspaceId/configuration/testConfigId", models.EventClassWorkload, "testCluster1", "testWorkload2")))
	id3 := sha256.Sum256([]byte(fmt.Sprintf("%s/%s/%s/%s", "workspace/testWorkspaceId/configuration/testConfigId", models.EventClassWorkload, "testCluster2", "testWorkload1")))
	id4 := sha256.Sum256([]byte(fmt.Sprintf("%s/%s/%s/%s", "workspace/testWorkspaceId/configuration/testConfigId", models.EventClassWorkload, "testCluster2", "testWorkload2")))
	newData := []workload.Workload{
		{
			ClusterName:  "testCluster1",
			WorkloadName: "testWorkload1",
			WorkloadType: "deployment",
			Containers: workload.Containers{
				Name:  "testContainer1",
				Image: "testImage1",
				Port:  "8080",
			},
			ServiceName: "serviceName1",
			Labels:      "k8s-app",
			Timestamp:   "",
			WorkloadProperties: workload.Properties{
				Schedule:       "testSchedule",
				Replicas:       "1",
				UpdateStrategy: "rollback",
			},
		},
		{
			ClusterName:  "testCluster1",
			WorkloadName: "testWorkload2",
			WorkloadType: "deployment",
			Containers: workload.Containers{
				Name:  "testContainer2",
				Image: "testImage2",
				Port:  "8080",
			},
			ServiceName: "serviceName2",
			Labels:      "k8s-app",
			Timestamp:   "",
			WorkloadProperties: workload.Properties{
				Schedule:       "testSchedule",
				Replicas:       "1",
				UpdateStrategy: "rollback",
			},
		},
		{
			ClusterName:  "testCluster2",
			WorkloadName: "testWorkload2",
			WorkloadType: "deployment",
			Containers: workload.Containers{
				Name:  "testContainer3",
				Image: "testImage3",
				Port:  "8080",
			},
			ServiceName: "serviceName3",
			Labels:      "k8s-app",
			Timestamp:   "",
			WorkloadProperties: workload.Properties{
				Schedule:       "testSchedule",
				Replicas:       "1",
				UpdateStrategy: "rollback",
			},
		},
	}
	oldData := []models.DiscoveryEvent{
		{
			HeaderProperties: models.HeaderProperties{
				Id:    hex.EncodeToString(id2[:]),
				Class: models.EventClassWorkload,
			},
			Body: models.DiscoveryBody{
				State: models.State{
					Name:           "",
					SourceInstance: "",
					SourceType:     "",
					Time:           "",
					Data: workload.Data{
						Workload: workload.Workload{
							ClusterName:  "testCluster1",
							WorkloadName: "testWorkload2",
							WorkloadType: "deployment",
							Containers: workload.Containers{
								Name:  "testContainer1",
								Image: "testImage1",
								Port:  "8080",
							},
							ServiceName: "serviceName1",
							Labels:      "k8s-app",
							Timestamp:   "",
							WorkloadProperties: workload.Properties{
								Schedule:       "testSchedule",
								Replicas:       "1",
								UpdateStrategy: "rollback",
							},
						}},
				},
			},
		},
		{
			HeaderProperties: models.HeaderProperties{
				Id:    hex.EncodeToString(id3[:]),
				Class: models.EventClassWorkload,
			},
			Body: models.DiscoveryBody{
				State: models.State{
					Name:           "",
					SourceInstance: "",
					SourceType:     "",
					Time:           "",
					Data: workload.Data{
						Workload: workload.Workload{
							ClusterName:  "testCluster2",
							WorkloadName: "testWorkload1",
							WorkloadType: "deployment",
							Containers: workload.Containers{
								Name:  "testContainer2",
								Image: "testImage2",
								Port:  "8080",
							},
							ServiceName: "serviceName2",
							Labels:      "k8s-app",
							Timestamp:   "",
							WorkloadProperties: workload.Properties{
								Schedule:       "testSchedule",
								Replicas:       "1",
								UpdateStrategy: "rollback",
							},
						}},
				},
			},
		},
		{
			HeaderProperties: models.HeaderProperties{
				Id:    hex.EncodeToString(id4[:]),
				Class: models.EventClassWorkload,
			},
			Body: models.DiscoveryBody{
				State: models.State{
					Name:           "",
					SourceInstance: "",
					SourceType:     "",
					Time:           "",
					Data: workload.Data{
						Workload: workload.Workload{
							ClusterName:  "testCluster2",
							WorkloadName: "testWorkload2",
							WorkloadType: "deployment",
							Containers: workload.Containers{
								Name:  "testContainer3",
								Image: "testImage3",
								Port:  "8080",
							},
							ServiceName: "serviceName3",
							Labels:      "k8s-app",
							Timestamp:   "",
							WorkloadProperties: workload.Properties{
								Schedule:       "testSchedule",
								Replicas:       "1",
								UpdateStrategy: "rollback",
							},
						}},
				},
			},
		},
	}
	//oldData map[string]models.DiscoveryEvent
	p := &workloadEventProducer{
		irisApi:     mockApi,
		runId:       "testRunId",
		workspaceId: "testWorkspaceId",
	}
	created, updated, deleted, err := p.CreateECSTWorkloadEvents(newData, oldData, "testConfigId")

	assert.NoError(t, err)
	assert.Len(t, created, 1)
	assert.Len(t, updated, 1)
	assert.Len(t, deleted, 1)
	// CREATED
	assert.Equal(t, hex.EncodeToString(id1[:]), created[0].HeaderProperties.Id)
	assert.Equal(t, models.EventTypeChange, created[0].HeaderProperties.Type)
	assert.Equal(t, models.EventActionCreated, created[0].HeaderProperties.Action)
	assert.Equal(t, "testCluster1", created[0].Body.State.Name)
	// UPDATED
	assert.Equal(t, models.EventTypeChange, updated[0].HeaderProperties.Type)
	assert.Equal(t, models.EventActionUpdated, updated[0].HeaderProperties.Action)
	parsedData, err := common.ParseWorkloadData(updated[0])
	assert.NoError(t, err)
	assert.Equal(t, "1", parsedData.Workload.WorkloadProperties.Replicas)
	// DELETED
	assert.Equal(t, models.EventTypeChange, deleted[0].HeaderProperties.Type)
	assert.Equal(t, models.EventActionDeleted, deleted[0].HeaderProperties.Action)
	parsedData, err = common.ParseWorkloadData(deleted[0])
	assert.NoError(t, err)
	assert.Equal(t, "testWorkload1", parsedData.Workload.WorkloadName)
	assert.Equal(t, "testCluster2", parsedData.Workload.ClusterName)
}

func Test_eventProducer_processECSTResults_empty(t *testing.T) {
	mockApi := mocks.NewIrisApi(t)

	var newData []workload.Workload
	var oldData []models.DiscoveryEvent
	p := NewEventWorkloadProducer(mockApi, "testRunId", "testWorkspaceId")
	err := p.ProcessWorkloads(newData, oldData, "testConfigId")

	assert.NoError(t, err)
}
