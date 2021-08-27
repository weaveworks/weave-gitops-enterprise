package converter

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	ammodels "github.com/prometheus/alertmanager/api/v2/models"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/models"
	"github.com/weaveworks/weave-gitops-enterprise/common/messaging/payload"
	"gorm.io/datatypes"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func newEvent(reason, kind, namespace, name string) payload.KubernetesEvent {
	event := v1.Event{
		Reason: reason,
		InvolvedObject: v1.ObjectReference{
			Kind:      kind,
			Namespace: namespace,
			Name:      name,
		},
	}
	ret := payload.KubernetesEvent{
		Event: event,
	}
	return ret
}

func newClusterInfo(id, typ string) payload.ClusterInfo {
	return payload.ClusterInfo{
		Cluster: payload.Cluster{
			ID:   id,
			Type: typ,
		},
	}
}

func newNodeInfo(id, name, kubeletVersion string, isControlPlane bool) payload.Node {
	return payload.Node{
		MachineID:      id,
		Name:           name,
		KubeletVersion: kubeletVersion,
		IsControlPlane: isControlPlane,
	}
}

func newAlert(generatorURL, finPrint, state string, start, end, update time.Time,
	annot, labels ammodels.LabelSet, inhibitedBy, silencedBy []string,
	receivers []*ammodels.Receiver) ammodels.GettableAlert {
	startDate := strfmt.DateTime(start)
	endDate := strfmt.DateTime(end)
	updatedDate := strfmt.DateTime(update)

	alertStatus := ammodels.AlertStatus{
		InhibitedBy: inhibitedBy,
		SilencedBy:  silencedBy,
		State:       &state,
	}

	alertStruct := ammodels.Alert{
		GeneratorURL: strfmt.URI(generatorURL),
		Labels:       labels,
	}

	alert := ammodels.GettableAlert{
		Annotations: annot,
		EndsAt:      &endDate,
		Fingerprint: &finPrint,
		Receivers:   receivers,
		StartsAt:    &startDate,
		Status:      &alertStatus,
		UpdatedAt:   &updatedDate,
		Alert:       alertStruct,
	}

	return alert
}

func TestToJSONErr(t *testing.T) {
	logsHook := logrustest.NewGlobal()
	defer logsHook.Reset()

	// un-encodable object
	value := toJSON(make(chan int))
	assert.Equal(t, value, []byte{})
	assertErrorsLoggedContains(t, logsHook, "Error converting %!s(chan int")
}

func TestSerializeLabelSet(t *testing.T) {
	testLabelSet := ammodels.LabelSet{"label1": "1", "label2": "test_label", "label3": "foo"}
	flattenedLabels := SerializeLabelSet(testLabelSet)

	assert.Contains(t, flattenedLabels, "label1:1,")
	assert.Contains(t, flattenedLabels, "label2:test_label,")
	assert.Contains(t, flattenedLabels, "label3:foo,")
}

func TestSerializeStringSlice(t *testing.T) {
	testStringSlice := []string{"str1", "str2", "str3"}
	flattenedStringSlice := SerializeStringSlice(testStringSlice)
	t.Log(testStringSlice)
	t.Log(flattenedStringSlice)

	assert.Contains(t, flattenedStringSlice, "str1,")
	assert.Contains(t, flattenedStringSlice, ",str2,")
	assert.Contains(t, flattenedStringSlice, ",str3")
}

func TestSerializeStringMap(t *testing.T) {
	testStringMap := map[string]string{"label1": "1", "label2": "test_label", "label3": "foo"}
	flattenedMap := SerializeStringMap(testStringMap)

	assert.Contains(t, flattenedMap, "label1:1,")
	assert.Contains(t, flattenedMap, "label2:test_label,")
	assert.Contains(t, flattenedMap, "label3:foo,")
}

func TestSerializeEventToJSON(t *testing.T) {
	reason := "FailedToCreateContainer"
	kind := "Pod"
	namespace := "kube-system"
	name := "weave-net-5zqlf"
	testEvent := newEvent(reason, kind, namespace, name)

	// Serialize a v1.Event to the gorm JSON datatype as byte array
	b := SerializeEventToJSON(&testEvent.Event)
	assert.Contains(t, string(b), reason)
	assert.Contains(t, string(b), namespace)
	assert.Contains(t, string(b), name)
}

func TestDeserializeEventToJSON(t *testing.T) {
	reason := "FailedToCreateContainer"
	kind := "Pod"
	namespace := "kube-system"
	name := "weave-net-5zqlf"
	testEvent := newEvent(reason, kind, namespace, name)

	// First serialize a v1.Event to the gorm JSON datatype as byte array
	b := SerializeEventToJSON(&testEvent.Event)

	// Recreate the v1.Event struct and compare the fields
	parsedEvent, err := DeserializeJSONToEvent(b)
	assert.NoError(t, err)
	assert.Equal(t, parsedEvent.Reason, reason)
	assert.Equal(t, parsedEvent.InvolvedObject.Namespace, namespace)
	assert.Equal(t, parsedEvent.InvolvedObject.Name, name)
}

func TestConvertEvent(t *testing.T) {
	reason := "FailedToCreateContainer"
	kind := "Pod"
	namespace := "kube-system"
	name := "weave-net-5zqlf"
	testEvent := newEvent(reason, kind, namespace, name)

	// Convert v1.Event to models.Event
	dbEvent := ConvertEvent(testEvent)

	// Ensure that the reason is written correctly
	assert.Equal(t, dbEvent.Reason, reason)

	// Deserialize the JSON part of the model.Event to a v1.Event struct
	// and ensure that the JSON contains the correct values
	parsedEvent, err := DeserializeJSONToEvent(dbEvent.RawEvent)
	assert.NoError(t, err)

	assert.Equal(t, parsedEvent.Reason, testEvent.Event.Reason)
	assert.Equal(t, parsedEvent.InvolvedObject.Kind, testEvent.Event.InvolvedObject.Kind)
	assert.Equal(t, parsedEvent.InvolvedObject.Namespace, testEvent.Event.InvolvedObject.Namespace)
	assert.Equal(t, parsedEvent.InvolvedObject.Name, testEvent.Event.InvolvedObject.Name)
}

func TestConvertClusterInfo(t *testing.T) {
	testClusterInfo := newClusterInfo("8cb9581a-1de1-4a7b-ab2d-16791acc8f74", "existingInfra")

	// Convert payload.ClusterInfo to models.ClusterInfo
	dbClusterInfo := ConvertClusterInfo(testClusterInfo)

	assert.Equal(t, testClusterInfo.Cluster.ID, string(dbClusterInfo.UID))
	assert.Equal(t, testClusterInfo.Cluster.Type, dbClusterInfo.Type)
}

func TestConvertNodeInfo(t *testing.T) {
	testClusterInfo := newClusterInfo("8cb9581a-1de1-4a7b-ab2d-16791acc8f74", "existingInfra")

	// Convert payload.ClusterInfo to models.ClusterInfo
	dbClusterInfo := ConvertClusterInfo(testClusterInfo)

	cp := newNodeInfo("3f28d1dd7291784ed454f52ba0937337", "foo-wks-1", "v1.19.3", true)
	worker := newNodeInfo("953089b9924d3a45febe69bc3add4683", "foo-wks-2", "v1.19.4", false)
	testClusterInfo.Cluster.Nodes = append(testClusterInfo.Cluster.Nodes, cp)
	testClusterInfo.Cluster.Nodes = append(testClusterInfo.Cluster.Nodes, worker)

	// Convert payload.ClusterInfo to models.ClusterInfo
	dbNodeInfo := ConvertNodeInfo(testClusterInfo, dbClusterInfo.UID)

	assert.Len(t, dbNodeInfo, 2)
	assert.Equal(t, cp.MachineID, string(dbNodeInfo[0].UID))
	assert.Equal(t, cp.Name, dbNodeInfo[0].Name)
	assert.Equal(t, cp.IsControlPlane, dbNodeInfo[0].IsControlPlane)
	assert.Equal(t, cp.KubeletVersion, dbNodeInfo[0].KubeletVersion)
	assert.Equal(t, dbClusterInfo.UID, dbNodeInfo[0].ClusterInfoUID)

	assert.Equal(t, worker.MachineID, string(dbNodeInfo[1].UID))
	assert.Equal(t, worker.Name, dbNodeInfo[1].Name)
	assert.Equal(t, worker.IsControlPlane, dbNodeInfo[1].IsControlPlane)
	assert.Equal(t, worker.KubeletVersion, dbNodeInfo[1].KubeletVersion)
	assert.Equal(t, dbClusterInfo.UID, dbNodeInfo[1].ClusterInfoUID)
}

func TestConvertFluxInfo(t *testing.T) {
	logs := []payload.FluxLogInfo{
		{
			Timestamp: "2021-03-04T02:45:05.544496465Z",
			URL:       "ssh://git@github.com/foot/wk-test.git",
			Branch:    "master",
			Head:      "abc",
			Event:     "",
		},
	}

	output := &bytes.Buffer{}
	encoder := json.NewEncoder(output)
	encoder.Encode(logs)
	logsJSONBytes := output.Bytes()

	tests := []struct {
		message payload.FluxInfo
		result  []models.FluxInfo
	}{
		{
			payload.FluxInfo{
				// empty list of flux deployments
				Token:       "derp",
				Deployments: []payload.FluxDeploymentInfo{},
			},
			[]models.FluxInfo{},
		},
		{
			payload.FluxInfo{
				// list of flux deployments with 1 item
				Token: "derp",
				Deployments: []payload.FluxDeploymentInfo{
					{
						Name:      "flux",
						Namespace: "wkp-flux",
						Args: []string{
							"--memcached-service=",
							"--ssh-keygen-dir=/var/fluxd/keygen",
							"--sync-garbage-collection=true",
							"--git-poll-interval=10s",
							"--sync-interval=10s",
							"--manifest-generation=true",
							"--listen-metrics=:3031",
							"--git-url=git@github.com:dinosk/fluxes-1.git",
							"--git-branch=master",
							"--registry-exclude-image=*"},
						Image: "docker.io/weaveworks/wkp-jk-init:v2.0.3-RC.1-2-gd677dc0a",
						Syncs: []payload.FluxLogInfo{
							{
								Timestamp: "2021-03-04T02:45:05.544496465Z",
								URL:       "ssh://git@github.com/foot/wk-test.git",
								Branch:    "master",
								Head:      "abc",
								Event:     "",
							},
						},
					},
				},
			},
			[]models.FluxInfo{
				{
					ClusterToken: "derp",
					Name:         "flux",
					Namespace:    "wkp-flux",
					Args:         "--memcached-service=,--ssh-keygen-dir=/var/fluxd/keygen,--sync-garbage-collection=true,--git-poll-interval=10s,--sync-interval=10s,--manifest-generation=true,--listen-metrics=:3031,--git-url=git@github.com:dinosk/fluxes-1.git,--git-branch=master,--registry-exclude-image=*",
					Image:        "docker.io/weaveworks/wkp-jk-init:v2.0.3-RC.1-2-gd677dc0a",
					RepoURL:      "git@github.com:dinosk/fluxes-1.git",
					RepoBranch:   "master",
					Syncs:        datatypes.JSON(logsJSONBytes),
				},
			},
		},
		{
			payload.FluxInfo{
				// list of flux deployments with 3 items
				Token: "derp",
				Deployments: []payload.FluxDeploymentInfo{
					{
						Name:      "flux",
						Namespace: "wkp-flux",
						Args: []string{
							"--memcached-service=",
							"--ssh-keygen-dir=/var/fluxd/keygen",
							"--sync-garbage-collection=true",
							"--git-poll-interval=10s",
							"--sync-interval=10s",
							"--manifest-generation=true",
							"--listen-metrics=:3031",
							"--git-url=git@github.com:dinosk/fluxes-1.git",
							"--git-branch=master",
							"--registry-exclude-image=*"},
						Image: "docker.io/weaveworks/wkp-jk-init:v2.0.3-RC.1-2-gd677dc0a",
						Syncs: []payload.FluxLogInfo{
							{
								Timestamp: "2021-03-04T02:45:05.544496465Z",
								URL:       "ssh://git@github.com/foot/wk-test.git",
								Branch:    "master",
								Head:      "abc",
								Event:     "",
							},
						},
					},
					{
						Name:      "flux-2",
						Namespace: "wkp-workspaces",
						Args: []string{
							"--ssh-keygen-dir=/var/fluxd/keygen",
							"--git-poll-interval=3s",
							"--sync-interval=10s",
							"--manifest-generation=true",
							"--git-url=git@github.com:weaveworks/foo.git",
							"--git-branch=main",
							"--registry-exclude-image=*"},
						Image: "myuser/custom-flux:gd677dc0a",
						Syncs: []payload.FluxLogInfo{
							{
								Timestamp: "2021-03-04T02:45:05.544496465Z",
								URL:       "ssh://git@github.com/foot/wk-test.git",
								Branch:    "master",
								Head:      "abc",
								Event:     "",
							},
						},
					},
					{
						Name:      "flux-3",
						Namespace: "default",
						Args: []string{
							"--memcached-service=",
							"--ssh-keygen-dir=/var/fluxd/keygen",
							"--sync-garbage-collection=true",
							"--git-poll-interval=1s",
							"--sync-interval=1s",
							"--manifest-generation=true",
							"--listen-metrics=:3031",
							"--git-url=git@github.com:test/test-flux.git",
							"--git-branch=dev",
							"--registry-exclude-image=*"},
						Image: "docker.io/weaveworks/flux:latest",
						Syncs: []payload.FluxLogInfo{
							{
								Timestamp: "2021-03-04T02:45:05.544496465Z",
								URL:       "ssh://git@github.com/foot/wk-test.git",
								Branch:    "master",
								Head:      "abc",
								Event:     "",
							},
						},
					},
				},
			},
			[]models.FluxInfo{
				{
					ClusterToken: "derp",
					Name:         "flux",
					Namespace:    "wkp-flux",
					Args:         "--memcached-service=,--ssh-keygen-dir=/var/fluxd/keygen,--sync-garbage-collection=true,--git-poll-interval=10s,--sync-interval=10s,--manifest-generation=true,--listen-metrics=:3031,--git-url=git@github.com:dinosk/fluxes-1.git,--git-branch=master,--registry-exclude-image=*",
					Image:        "docker.io/weaveworks/wkp-jk-init:v2.0.3-RC.1-2-gd677dc0a",
					RepoURL:      "git@github.com:dinosk/fluxes-1.git",
					RepoBranch:   "master",
					Syncs:        datatypes.JSON(logsJSONBytes),
				},
				{
					ClusterToken: "derp",
					Name:         "flux-2",
					Namespace:    "wkp-workspaces",
					Args:         "--ssh-keygen-dir=/var/fluxd/keygen,--git-poll-interval=3s,--sync-interval=10s,--manifest-generation=true,--git-url=git@github.com:weaveworks/foo.git,--git-branch=main,--registry-exclude-image=*",
					Image:        "myuser/custom-flux:gd677dc0a",
					RepoURL:      "git@github.com:weaveworks/foo.git",
					RepoBranch:   "main",
					Syncs:        datatypes.JSON(logsJSONBytes),
				},
				{
					ClusterToken: "derp",
					Name:         "flux-3",
					Namespace:    "default",
					Args:         "--memcached-service=,--ssh-keygen-dir=/var/fluxd/keygen,--sync-garbage-collection=true,--git-poll-interval=1s,--sync-interval=1s,--manifest-generation=true,--listen-metrics=:3031,--git-url=git@github.com:test/test-flux.git,--git-branch=dev,--registry-exclude-image=*",
					Image:        "docker.io/weaveworks/flux:latest",
					RepoURL:      "git@github.com:test/test-flux.git",
					RepoBranch:   "dev",
					Syncs:        datatypes.JSON(logsJSONBytes),
				},
			},
		},
	}

	for _, test := range tests {
		result := ConvertFluxInfo(test.message)
		assert.Equal(t, test.result, result)
	}
}

func TestConvertAlert(t *testing.T) {
	// Alert dates
	startDate := time.Now()
	endDate := startDate.Add(time.Duration(60) * time.Minute)
	updatedDate := startDate.Add(time.Duration(30) * time.Minute)

	annot := ammodels.LabelSet{
		"summary":     "Instance down",
		"description": "Instance has been down for more than 5 minutes.",
	}
	labls := ammodels.LabelSet{
		"severity": "critical",
	}

	var strSlice = []string{"Test1", "Test2", "Test3"}

	receiverName := "My Receiver 1"
	receivr := ammodels.Receiver{
		Name: &receiverName,
	}
	receivrs := []*ammodels.Receiver{&receivr}

	alert := newAlert("example.com", "Test Fingerprint", "active",
		startDate, endDate, updatedDate, annot, labls, strSlice, strSlice, receivrs)

	// Convert payload.PrometheusAlerts to models.Alert
	dbAlert := ConvertAlert("derp", &alert)

	assert.Equal(t, "derp", dbAlert.ClusterToken)
	assert.Equal(t, "active", dbAlert.State)
	assert.Equal(t, endDate, dbAlert.EndsAt)
	assert.Equal(t, startDate, dbAlert.StartsAt)
	assert.Equal(t, updatedDate, dbAlert.UpdatedAt)
	assert.Equal(t, "Test Fingerprint", dbAlert.Fingerprint)
	assert.Equal(t, "example.com", dbAlert.GeneratorURL)
	assert.Equal(t, "Test1,Test2,Test3", dbAlert.InhibitedBy)
	assert.Equal(t, "Test1,Test2,Test3", dbAlert.SilencedBy)
	assert.Equal(t, "critical", dbAlert.Severity)
}

type Labels struct {
	Severity string `json:"severity"`
}

type Annotations struct {
	Summary     string `json:"summary"`
	Description string `json:"description"`
}

func TestExtractRepoURLfromFluxArgs(t *testing.T) {
	tests := []struct {
		args      []string
		parsedURL string
	}{
		{
			args: []string{
				"--memcached-service=",
				"--ssh-keygen-dir=/var/fluxd/keygen",
				"--sync-garbage-collection=true",
				"--git-poll-interval=10s",
				"--sync-interval=10s",
				"--manifest-generation=true",
				"--listen-metrics=:3031",
				"--git-url=git@github.com:ww/fluxes-1.git",
				"--git-branch=master",
				"--registry-exclude-image=*",
			},
			parsedURL: "git@github.com:ww/fluxes-1.git",
		},
		{
			args: []string{
				"--memcached-service=",
				"--ssh-keygen-dir=/var/fluxd/keygen",
				"--sync-garbage-collection=true",
				"--git-poll-interval=10s",
				"--sync-interval=10s",
				"--manifest-generation=true",
				"--listen-metrics=:3031",
				"--git-url=https://github.com/ww/fluxes-2",
				"--git-branch=master",
				"--registry-exclude-image=*",
			},
			parsedURL: "https://github.com/ww/fluxes-2",
		},
		{
			args: []string{
				"--memcached-service=",
				"--ssh-keygen-dir=/var/fluxd/keygen",
				"--sync-garbage-collection=true",
				"--git-poll-interval=10s",
				"--sync-interval=10s",
				"--manifest-generation=true",
				"--listen-metrics=:3031",
				"--git-branch=master",
				"--registry-exclude-image=*",
			},
			parsedURL: "",
		},
	}

	for _, test := range tests {
		result := ExtractRepoURLfromFluxArgs(test.args)
		assert.Equal(t, test.parsedURL, result)
	}
}

func TestConvertGitCommitInfo(t *testing.T) {
	commit := payload.GitCommitInfo{
		Token: "derp",
		Commit: payload.CommitView{
			Sha:     "123",
			Message: "GitOps ftw",
			Author: payload.UserView{
				Name:  "foo",
				Email: "foo@weave.works",
				Date:  time.Now().UTC(),
			},
			Committer: payload.UserView{
				Name:  "bar",
				Email: "bar@weave.works",
				Date:  time.Now().UTC(),
			},
		},
	}

	dbCommit := ConvertGitCommitInfo(commit)

	assert.Equal(t, commit.Token, dbCommit.ClusterToken)
	assert.Equal(t, commit.Commit.Sha, dbCommit.Sha)
	assert.Equal(t, commit.Commit.Message, dbCommit.Message)
	assert.Equal(t, commit.Commit.Author.Name, dbCommit.AuthorName)
	assert.Equal(t, commit.Commit.Author.Email, dbCommit.AuthorEmail)
	assert.Equal(t, commit.Commit.Author.Date, dbCommit.AuthorDate)
	assert.Equal(t, commit.Commit.Committer.Name, dbCommit.CommitterName)
	assert.Equal(t, commit.Commit.Committer.Email, dbCommit.CommitterEmail)
	assert.Equal(t, commit.Commit.Committer.Date, dbCommit.CommitterDate)
}

func TestConvertWorkspaceInfo(t *testing.T) {
	info := payload.WorkspaceInfo{
		Token: "derp",
		Workspaces: []payload.Workspace{
			{
				Name:      "foo-ws",
				Namespace: "foo-ns",
			},
			{
				Name:      "bar-ws",
				Namespace: "bar-ns",
			},
		},
	}

	dbWorkspaces := ConvertWorkspaceInfo(info)

	assert.Equal(t, info.Token, dbWorkspaces[0].ClusterToken)
	assert.Equal(t, info.Workspaces[0].Name, dbWorkspaces[0].Name)
	assert.Equal(t, info.Workspaces[0].Namespace, dbWorkspaces[0].Namespace)
	assert.Equal(t, info.Token, dbWorkspaces[1].ClusterToken)
	assert.Equal(t, info.Workspaces[1].Name, dbWorkspaces[1].Name)
	assert.Equal(t, info.Workspaces[1].Namespace, dbWorkspaces[1].Namespace)
}

func assertErrorsLoggedContains(t *testing.T, logsHook *logrustest.Hook, needle string) {
	actualErrorLogs := ""
	for _, en := range logsHook.Entries {
		if en.Level == logrus.ErrorLevel {
			actualErrorLogs = actualErrorLogs + en.Message
		}
	}
	assert.Contains(t, actualErrorLogs, needle)
}

func TestConvertCAPIClusterInfo(t *testing.T) {
	info := payload.CAPIClusterInfo{
		Token: "derp",
		CAPIClusters: []payload.CAPICluster{
			{
				Name:          "foo-ws",
				Namespace:     "foo-ns",
				CAPIVersion:   "foo-version",
				EncodedObject: `"hi"`,
			},
			{
				Name:        "bar-ws",
				Namespace:   "bar-ns",
				CAPIVersion: "bar-version",
			},
		},
	}

	clusters := ConvertCAPIClusterInfo(info)

	assert.Equal(t, info.Token, clusters[0].ClusterToken)
	assert.Equal(t, info.CAPIClusters[0].Name, clusters[0].Name)
	assert.Equal(t, info.CAPIClusters[0].Namespace, clusters[0].Namespace)
	assert.Equal(t, `"hi"`, string(clusters[0].Object))

	assert.Equal(t, info.Token, clusters[1].ClusterToken)
	assert.Equal(t, info.CAPIClusters[1].Name, clusters[1].Name)
	assert.Equal(t, info.CAPIClusters[1].Namespace, clusters[1].Namespace)
}
