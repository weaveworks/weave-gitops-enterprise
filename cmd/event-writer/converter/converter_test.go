package converter

import (
	"testing"

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
