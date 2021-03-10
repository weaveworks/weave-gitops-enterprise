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
func ConvertEvent(wkpEvent payload.KubernetesEvent) models.Event {
	event := wkpEvent.Event
	eventJSONbytes := toJSON(&event)

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
		RawEvent:     datatypes.JSON(eventJSONbytes),
	}
	return result
}

// ConvertClusterInfo returns a models.ClusterInfo from a NATS message with cluster info
func ConvertClusterInfo(clusterInfo payload.ClusterInfo) models.ClusterInfo {
	cluster := clusterInfo.Cluster
	result := models.ClusterInfo{
		Token: clusterInfo.Token,
		UID:   types.UID(cluster.ID),
		Type:  cluster.Type,
	}
	return result
}

// ConvertNodeInfo returns a models.Node from a NATS message with node info
func ConvertNodeInfo(clusterInfo payload.ClusterInfo, clusterID types.UID) []models.NodeInfo {
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
	return result
}

// ConvertAlert returns a models.Alert from a NATS message with alert info
func ConvertAlert(token string, gAlert *ammodels.GettableAlert) models.Alert {
	alert := gAlert
	alertJSONbytes := toJSON(alert)
	annotationsJSONBytes := toJSON(alert.Annotations)
	labelsJSONBytes := toJSON(alert.Labels)

	flattenedInhibitedBy := SerializeStringSlice(alert.Status.InhibitedBy)
	flattenedSilencedBy := SerializeStringSlice(alert.Status.SilencedBy)
	Severity := alert.Alert.Labels["severity"]

	result := models.Alert{
		ClusterToken: token,
		Annotations:  datatypes.JSON(annotationsJSONBytes),
		EndsAt:       time.Time(*alert.EndsAt),
		Fingerprint:  *alert.Fingerprint,
		InhibitedBy:  flattenedInhibitedBy,
		SilencedBy:   flattenedSilencedBy,
		Severity:     Severity,
		State:        *alert.Status.State,
		StartsAt:     time.Time(*alert.StartsAt),
		UpdatedAt:    time.Time(*alert.UpdatedAt),
		GeneratorURL: alert.Alert.GeneratorURL.String(),
		Labels:       datatypes.JSON(labelsJSONBytes),
		RawAlert:     datatypes.JSON(alertJSONbytes),
	}

	return result
}

// ConvertFluxInfo returns a models.FluxInfo from a NATS message with cluster info
func ConvertFluxInfo(fluxInfo payload.FluxInfo) []models.FluxInfo {
	result := []models.FluxInfo{}

	for _, fluxDeploymentInfo := range fluxInfo.Deployments {

		fluxRepoURL := ExtractRepoURLfromFluxArgs(fluxDeploymentInfo.Args)
		fluxRepoBranch := ExtractRepoBranchfromFluxArgs(fluxDeploymentInfo.Args)

		flattenedArgs := SerializeStringSlice(fluxDeploymentInfo.Args)
		result = append(result,
			models.FluxInfo{
				ClusterToken: fluxInfo.Token,
				Name:         fluxDeploymentInfo.Name,
				Namespace:    fluxDeploymentInfo.Namespace,
				Args:         flattenedArgs,
				Image:        fluxDeploymentInfo.Image,
				RepoURL:      fluxRepoURL,
				RepoBranch:   fluxRepoBranch,
			})
	}

	return result
}

func ConvertGitCommitInfo(commitInfo payload.GitCommitInfo) models.GitCommit {
	return models.GitCommit{
		ClusterToken:   commitInfo.Token,
		Sha:            commitInfo.Commit.Sha,
		AuthorName:     commitInfo.Commit.Author.Name,
		AuthorEmail:    commitInfo.Commit.Author.Email,
		AuthorDate:     commitInfo.Commit.Author.Date,
		CommitterName:  commitInfo.Commit.Committer.Name,
		CommitterEmail: commitInfo.Commit.Committer.Email,
		CommitterDate:  commitInfo.Commit.Committer.Date,
		Message:        commitInfo.Commit.Message,
	}
}

func ConvertWorkspaceInfo(workspaceInfo payload.WorkspaceInfo) []models.Workspace {
	result := []models.Workspace{}

	for _, ws := range workspaceInfo.Workspaces {
		result = append(result, models.Workspace{
			ClusterToken: workspaceInfo.Token,
			Name:         ws.Name,
			Namespace:    ws.Namespace,
		})
	}

	return result
}

// SerializeLabelSet flattens a labelset to a string
func SerializeLabelSet(labels ammodels.LabelSet) string {
	labelMap := map[string]string(labels)
	return SerializeStringMap(labelMap)
}

// SerializeStringSlice flattens a slice of strings to a string
func SerializeStringSlice(strSlice []string) string {
	str := strings.Join(strSlice, ",")

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

func toJSON(obj interface{}) []byte {
	output := bytes.NewBufferString("")
	encoder := json.NewEncoder(output)
	encoder.Encode(obj)
	return output.Bytes()
}

// SerializeEventToJSON serializes a v1.Event object to a byte array
func SerializeEventToJSON(e *v1.Event) []byte {
	return toJSON(e)
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

// ExtractRepoURLfromFluxArgs parses a string slice of flux start arguments and returns the repo URL
func ExtractRepoURLfromFluxArgs(args []string) string {
	gitURLPrefix := "--git-url="
	return ExtractArgValue(args, gitURLPrefix)
}

// ExtractRepoBranchfromFluxArgs parses a string slice of flux start arguments and returns the repo branch
func ExtractRepoBranchfromFluxArgs(args []string) string {
	gitBranchPrefix := "--git-branch="
	return ExtractArgValue(args, gitBranchPrefix)
}

// ExtractArgValue returns a given argument value from a string slice of flux run arguments
func ExtractArgValue(args []string, arg string) string {
	var result string

	for _, item := range args {
		if strings.HasPrefix(item, arg) {
			result = strings.TrimPrefix(item, arg)
			return result
		}
	}
	return ""
}
