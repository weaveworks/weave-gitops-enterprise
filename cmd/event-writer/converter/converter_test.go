package converter

import (
	"testing"

	"github.com/weaveworks/wks/common/messaging/payload"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func newEvent(reason, kind, namespace, name string) v1.Event {
	event := v1.Event{
		Reason: reason,
		InvolvedObject: v1.ObjectReference{
			Kind:      kind,
			Namespace: namespace,
			Name:      name,
		},
	}
	return event
}

func newClusterInfo(id, typ string) payload.ClusterInfo {
	return payload.ClusterInfo{
		ID:   id,
		Type: typ,
	}
}

func newNodeInfo(id, name, kubeletVersion string, isControlPlane bool) payload.NodeInfo {
	return payload.NodeInfo{
		MachineID:      id,
		Name:           name,
		KubeletVersion: kubeletVersion,
		IsControlPlane: isControlPlane,
	}
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
	b, err := SerializeEventToJSON(&testEvent)
	assert.NoError(t, err)
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
	b, err := SerializeEventToJSON(&testEvent)
	assert.NoError(t, err)

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
	dbEvent, err := ConvertEvent(testEvent)
	assert.NoError(t, err)

	// Ensure that the reason is written correctly
	assert.Equal(t, dbEvent.Reason, reason)

	// Deserialize the JSON part of the model.Event to a v1.Event struct
	// and ensure that the JSON contains the correct values
	parsedEvent, err := DeserializeJSONToEvent(dbEvent.RawEvent)
	assert.NoError(t, err)

	assert.Equal(t, parsedEvent.Reason, testEvent.Reason)
	assert.Equal(t, parsedEvent.InvolvedObject.Kind, testEvent.InvolvedObject.Kind)
	assert.Equal(t, parsedEvent.InvolvedObject.Namespace, testEvent.InvolvedObject.Namespace)
	assert.Equal(t, parsedEvent.InvolvedObject.Name, testEvent.InvolvedObject.Name)
}

func TestConvertClusterInfo(t *testing.T) {
	testClusterInfo := newClusterInfo("8cb9581a-1de1-4a7b-ab2d-16791acc8f74", "existingInfra")

	// Convert payload.ClusterInfo to models.ClusterInfo
	dbClusterInfo, err := ConvertClusterInfo(testClusterInfo)
	assert.NoError(t, err)

	assert.Equal(t, testClusterInfo.ID, string(dbClusterInfo.UID))
	assert.Equal(t, testClusterInfo.Type, dbClusterInfo.Type)
}

func TestConvertNodeInfo(t *testing.T) {
	testClusterInfo := newClusterInfo("8cb9581a-1de1-4a7b-ab2d-16791acc8f74", "existingInfra")

	// Convert payload.ClusterInfo to models.ClusterInfo
	dbClusterInfo, err := ConvertClusterInfo(testClusterInfo)
	assert.NoError(t, err)

	cp := newNodeInfo("3f28d1dd7291784ed454f52ba0937337", "foo-wks-1", "v1.19.3", true)
	worker := newNodeInfo("953089b9924d3a45febe69bc3add4683", "foo-wks-2", "v1.19.4", false)
	testClusterInfo.Nodes = append(testClusterInfo.Nodes, cp)
	testClusterInfo.Nodes = append(testClusterInfo.Nodes, worker)

	// Convert payload.ClusterInfo to models.ClusterInfo
	dbNodeInfo, err := ConvertNodeInfo(testClusterInfo, dbClusterInfo.UID)
	assert.NoError(t, err)

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
