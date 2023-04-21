package iris

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/leanix/leanix-k8s-connector/pkg/iris/models/namespace"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_eventProducer_filter_created(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockApi := NewMockAPI(ctrl)
	newData := map[string]models.Data{

		"testId1": {
			models.ClusterEcst{
				Namespace:   "testNamespace1",
				Deployments: make([]models.DeploymentEcst, 0),
				Name:        "testCluster1",
				Os:          "testOs1",
				K8sVersion:  "1.0.0",
				NoOfNodes:   "1",
			}},
	}
	oldData := map[string]models.DiscoveryEvent{
		"testId2": {
			Body: models.DiscoveryBody{
				State: models.State{
					Name:           "",
					SourceInstance: "",
					SourceType:     "",
					Time:           "",
					Data:           models.Data{},
				},
			},
		},
	}
	//oldData map[string]models.DiscoveryEvent
	p := NewEventProducer(mockApi, "testRunId", "testWorkspaceId")
	created, updated, _, err := p.FilterForChangedItems(newData, oldData, "testConfigId")

	assert.NoError(t, err)
	assert.Empty(t, updated)
	assert.Len(t, created, 1)
	assert.Equal(t, "testNamespace1", created[0].Body.State.Name)
}

func Test_eventProducer_filter_updated_no_change(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockApi := NewMockAPI(ctrl)
	newData := map[string]models.Data{

		"testId1": {
			models.ClusterEcst{
				Namespace:   "testNamespace1",
				Deployments: make([]models.DeploymentEcst, 0),
				Name:        "testCluster1",
				Os:          "testOs1",
				K8sVersion:  "1.0.0",
				NoOfNodes:   "1",
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
					Data: models.Data{
						Cluster: models.ClusterEcst{
							Namespace:   "testNamespace1",
							Deployments: make([]models.DeploymentEcst, 0),
							Name:        "testCluster1",
							Os:          "testOs1",
							K8sVersion:  "1.0.0",
							NoOfNodes:   "1",
						},
					},
				},
			},
		},
	}
	//oldData map[string]models.DiscoveryEvent
	p := NewEventProducer(mockApi, "testRunId", "testWorkspaceId")
	created, updated, filteredData, err := p.FilterForChangedItems(newData, oldData, "testConfigId")

	assert.NoError(t, err)
	assert.Empty(t, updated)
	assert.Empty(t, created)
	assert.Empty(t, filteredData)
}

func Test_eventProducer_filter_updated_changed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockApi := NewMockAPI(ctrl)
	newData := map[string]models.Data{

		"testId1": {
			models.ClusterEcst{
				Namespace:   "testNamespace1",
				Deployments: make([]models.DeploymentEcst, 0),
				Name:        "testCluster1",
				Os:          "testOs1",
				K8sVersion:  "1.5.0",
				NoOfNodes:   "1",
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
					Data: models.Data{
						Cluster: models.ClusterEcst{
							Namespace:   "testNamespace1",
							Deployments: make([]models.DeploymentEcst, 0),
							Name:        "testCluster1",
							Os:          "testOs1",
							K8sVersion:  "1.0.0",
							NoOfNodes:   "1",
						},
					},
				},
			},
		},
	}
	//oldData map[string]models.DiscoveryEvent
	p := &eventProducer{
		irisApi:     mockApi,
		runId:       "testRunId",
		workspaceId: "testWorkspaceId",
	}
	created, updated, filteredData, err := p.FilterForChangedItems(newData, oldData, "testConfigId")

	assert.NoError(t, err)
	assert.Empty(t, created)
	assert.Len(t, updated, 1)
	assert.Equal(t, "1.5.0", updated[0].Body.State.Data.Cluster.K8sVersion)
	assert.Empty(t, filteredData)
}

func Test_eventProducer_createECSTEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockApi := NewMockAPI(ctrl)
	id1 := sha256.Sum256([]byte(fmt.Sprintf("%s/%s/%s/%s", "workspace/testWorkspaceId/configuration/testConfigId", EventClassNamespace, "testCluster1", "testNamespace1")))
	id2 := sha256.Sum256([]byte(fmt.Sprintf("%s/%s/%s/%s", "workspace/testWorkspaceId/configuration/testConfigId", EventClassNamespace, "testCluster1", "testNamespace2")))
	id3 := sha256.Sum256([]byte(fmt.Sprintf("%s/%s/%s/%s", "workspace/testWorkspaceId/configuration/testConfigId", EventClassNamespace, "testCluster2", "testNamespace1")))
	id4 := sha256.Sum256([]byte(fmt.Sprintf("%s/%s/%s/%s", "workspace/testWorkspaceId/configuration/testConfigId", EventClassNamespace, "testCluster2", "testNamespace2")))
	newData := []models.Data{
		{
			// NEW item
			models.ClusterEcst{
				Name:        "testCluster1",
				Namespace:   "testNamespace1",
				Deployments: make([]models.DeploymentEcst, 0),
				Os:          "testOs1",
				K8sVersion:  "1.0.0",
				NoOfNodes:   "1",
			},
		},
		{
			// Existing but not changed
			models.ClusterEcst{
				Name:        "testCluster1",
				Namespace:   "testNamespace2",
				Deployments: make([]models.DeploymentEcst, 0),
				Os:          "testOs1",
				K8sVersion:  "1.0.0",
				NoOfNodes:   "1",
			},
		},
		{
			// Existing and changed
			models.ClusterEcst{
				Name:        "testCluster2",
				Namespace:   "testNamespace1",
				Deployments: make([]models.DeploymentEcst, 0),
				Os:          "testOs1",
				K8sVersion:  "1.0.0",
				NoOfNodes:   "5",
			},
		},
	}
	oldData := []models.DiscoveryEvent{
		{
			HeaderProperties: models.HeaderProperties{
				Id: hex.EncodeToString(id2[:]),
			},
			Body: models.DiscoveryBody{
				State: models.State{
					Name:           "",
					SourceInstance: "",
					SourceType:     "",
					Time:           "",
					Data: models.Data{Cluster: models.ClusterEcst{
						Name:        "testCluster1",
						Namespace:   "testNamespace2",
						Deployments: make([]models.DeploymentEcst, 0),
						Os:          "testOs1",
						K8sVersion:  "1.0.0",
						NoOfNodes:   "1",
					}},
				},
			},
		},
		{
			HeaderProperties: models.HeaderProperties{
				Id: hex.EncodeToString(id3[:]),
			},
			Body: models.DiscoveryBody{
				State: models.State{
					Name:           "",
					SourceInstance: "",
					SourceType:     "",
					Time:           "",
					Data: models.Data{Cluster: models.ClusterEcst{
						Name:        "testCluster2",
						Namespace:   "testNamespace1",
						Deployments: make([]models.DeploymentEcst, 0),
						Os:          "testOs1",
						K8sVersion:  "1.0.0",
						NoOfNodes:   "1",
					},
					},
				},
			},
		},
		{
			HeaderProperties: models.HeaderProperties{
				Id: hex.EncodeToString(id4[:]),
			},
			Body: models.DiscoveryBody{
				State: models.State{
					Name:           "",
					SourceInstance: "",
					SourceType:     "",
					Time:           "",
					Data: models.Data{Cluster: models.ClusterEcst{
						Name:        "testCluster2",
						Namespace:   "testNamespace2",
						Deployments: make([]models.DeploymentEcst, 0),
						Os:          "testOs1",
						K8sVersion:  "1.0.0",
						NoOfNodes:   "1",
					},
					},
				},
			},
		},
	}
	//oldData map[string]models.DiscoveryEvent
	p := &eventProducer{
		irisApi:     mockApi,
		runId:       "testRunId",
		workspaceId: "testWorkspaceId",
	}
	created, updated, deleted, err := p.createECSTEvents(newData, oldData, "testConfigId")

	assert.NoError(t, err)
	assert.Len(t, created, 1)
	assert.Len(t, updated, 1)
	assert.Len(t, deleted, 1)
	// CREATED
	assert.Equal(t, hex.EncodeToString(id1[:]), created[0].HeaderProperties.Id)
	assert.Equal(t, EventTypeChange, created[0].HeaderProperties.Type)
	assert.Equal(t, EventActionCreated, created[0].HeaderProperties.Action)
	assert.Equal(t, "testNamespace1", created[0].Body.State.Name)
	// UPDATED
	assert.Equal(t, hex.EncodeToString(id3[:]), updated[0].HeaderProperties.Id)
	assert.Equal(t, EventTypeChange, updated[0].HeaderProperties.Type)
	assert.Equal(t, EventActionUpdated, updated[0].HeaderProperties.Action)
	assert.Equal(t, "5", updated[0].Body.State.Data.Cluster.NoOfNodes)
	// DELETED
	assert.Equal(t, hex.EncodeToString(id4[:]), deleted[0].HeaderProperties.Id)
	assert.Equal(t, EventTypeChange, deleted[0].HeaderProperties.Type)
	assert.Equal(t, EventActionDeleted, deleted[0].HeaderProperties.Action)
	assert.Equal(t, "testCluster2", deleted[0].Body.State.Data.Cluster.Name)
	assert.Equal(t, "testNamespace2", deleted[0].Body.State.Data.Cluster.Namespace)
}

func Test_eventProducer_processECSTResults_empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockApi := NewMockAPI(ctrl)

	var newData []models.Data
	var oldData []models.DiscoveryEvent
	p := NewEventProducer(mockApi, "testRunId", "testWorkspaceId")
	err := p.ProcessResults(newData, oldData, "testConfigId")

	assert.NoError(t, err)
}
