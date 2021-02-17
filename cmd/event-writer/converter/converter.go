package converter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	ammodels "github.com/prometheus/alertmanager/api/v2/models"
	"github.com/weaveworks/wks/common/database/models"
	"github.com/weaveworks/wks/common/messaging/payload"
	"gorm.io/datatypes"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
)

// ConvertEvent returns a models.Event from a payload.KubernetesEvent which wraps a v1.Event object
func ConvertEvent(wkpEvent payload.KubernetesEvent) (models.Event, error) {
	event := wkpEvent.Event
	eventJSONbytes, err := SerializeEventToJSON(&event)
	if err != nil {
		return models.Event{}, err
	}
	eventJSON, err := datatypes.JSON.MarshalJSON(eventJSONbytes)
	if err != nil {
		return models.Event{}, err
	}

	flattenedLabels := SerializeStringMap(event.ObjectMeta.Labels)
	flattenedAnnotations := SerializeStringMap(event.ObjectMeta.Annotations)
	creationTimestamp := datatypes.Date{}
	creationTimestamp.Scan(event.ObjectMeta.CreationTimestamp)
	registrationTimestamp := datatypes.Date{}
	registrationTimestamp.Scan(time.Now())

	result := models.Event{
		Token:        wkpEvent.Token,
		UID:          event.ObjectMeta.UID,
		CreatedAt:    creationTimestamp,
		RegisteredAt: registrationTimestamp,
		Name:         event.ObjectMeta.Name,
		Namespace:    event.ObjectMeta.Namespace,
		Labels:       flattenedLabels,
		Annotations:  flattenedAnnotations,
		ClusterName:  event.ClusterName,
		Reason:       event.Reason,
		Message:      event.Message,
		Type:         event.Type,
		RawEvent:     eventJSON,
	}
	return result, nil
}

// ConvertClusterInfo returns a models.ClusterInfo from a NATS message with cluster info
func ConvertClusterInfo(clusterInfo payload.ClusterInfo) (models.ClusterInfo, error) {
	cluster := clusterInfo.Cluster
	result := models.ClusterInfo{
		Token: clusterInfo.Token,
		UID:   types.UID(cluster.ID),
		Type:  cluster.Type,
	}
	return result, nil
}

// ConvertNodeInfo returns a models.Node from a NATS message with node info
func ConvertNodeInfo(clusterInfo payload.ClusterInfo, clusterID types.UID) ([]models.NodeInfo, error) {
	result := []models.NodeInfo{}
	for _, nodeInfo := range clusterInfo.Cluster.Nodes {
		result = append(result, models.NodeInfo{
			Token:          clusterInfo.Token,
			UID:            types.UID(nodeInfo.MachineID),
			ClusterInfoUID: clusterID,
			Name:           nodeInfo.Name,
			IsControlPlane: nodeInfo.IsControlPlane,
			KubeletVersion: nodeInfo.KubeletVersion,
		})
	}
	return result, nil
}

// ConvertAlert returns a models.Alert from a NATS message with alert info
func ConvertAlert(token string, gAlert *ammodels.GettableAlert) (models.Alert, error) {
	alert := gAlert
	alertJSONbytes, err := SerializeAlertToJSON(alert)
	if err != nil {
		return models.Alert{}, err
	}
	alertJSON, err := datatypes.JSON.MarshalJSON(alertJSONbytes)
	if err != nil {
		return models.Alert{}, err
	}

	flattenedLabels := SerializeLabelSet(alert.Alert.Labels)
	flattenedAnnotations := SerializeLabelSet(alert.Annotations)
	flattenedInhibitedBy := SerializeStringSlice(alert.Status.InhibitedBy)
	flattenedSilencedBy := SerializeStringSlice(alert.Status.SilencedBy)
	Severity := alert.Alert.Labels["severity"]

	result := models.Alert{
		Token:        token,
		Annotations:  flattenedAnnotations,
		EndsAt:       time.Time(*alert.EndsAt),
		Fingerprint:  *alert.Fingerprint,
		InhibitedBy:  flattenedInhibitedBy,
		SilencedBy:   flattenedSilencedBy,
		Severity:     Severity,
		State:        *alert.Status.State,
		StartsAt:     time.Time(*alert.StartsAt),
		UpdatedAt:    time.Time(*alert.UpdatedAt),
		GeneratorURL: alert.Alert.GeneratorURL.String(),
		Labels:       flattenedLabels,
		RawAlert:     alertJSON,
	}

	return result, nil
}

// SerializeLabelSet flattens a labelset to a string
func SerializeLabelSet(labels ammodels.LabelSet) string {
	labelMap := map[string]string(labels)
	return SerializeStringMap(labelMap)
}

// SerializeStringSlice flattens a slice of strings to a string
func SerializeStringSlice(strSlice []string) string {
	str := strings.Join(strSlice, ", ")

	return str
}

// SerializeStringMap flattens a string-to-string map to a string
func SerializeStringMap(m map[string]string) string {
	format := "%s:%s,"
	b := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(b, format, key, value)
	}
	return b.String()
}

// SerializeAlertToJSON serializes a alert manager models.Alert object to a byte array
func SerializeAlertToJSON(a *ammodels.GettableAlert) ([]byte, error) {
	output := bytes.NewBufferString("")
	encoder := json.NewEncoder(output)
	encoder.Encode(a)
	return output.Bytes(), nil
}

// SerializeEventToJSON serializes a v1.Event object to a byte array
func SerializeEventToJSON(e *v1.Event) ([]byte, error) {
	output := bytes.NewBufferString("")
	encoder := json.NewEncoder(output)
	encoder.Encode(e)
	return output.Bytes(), nil
}

// DeserializeJSONToEvent constructs a v1.Event from a byte array and returns a pointer to it
func DeserializeJSONToEvent(b []byte) (*v1.Event, error) {
	decoder := serializer.NewCodecFactory(scheme.Scheme).UniversalDecoder()
	e := &v1.Event{}
	err := runtime.DecodeInto(decoder, b, e)
	if err != nil {
		return nil, err
	}
	return e, nil
}
