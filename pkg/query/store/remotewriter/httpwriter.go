package remotewriter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/query"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/models"
	"net/http"
)

type httpRemoteStore struct {
	url   string
	token string
	log   logr.Logger
}

// TODO refactor and make it more reliable, timeouts, retries!
func (h httpRemoteStore) StoreRoles(ctx context.Context, roles []models.Role) error {
	storeRolesPath := "v1/query/roles"
	storeRolesUrl := fmt.Sprintf("%s/%s", h.url, storeRolesPath)

	req := &pb.StoreRolesRequest{
		Roles: convertToPbRole(roles),
	}

	bodyAsJson, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed json marshalling: %w", err)
	}
	request, err := http.NewRequest("POST", storeRolesUrl, bytes.NewBuffer(bodyAsJson))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Cookie", fmt.Sprintf("id_token=%s", h.token))

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("failed http request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return fmt.Errorf(fmt.Sprintf("response error: %s", response.Status))
	}

	return nil
}

func (h httpRemoteStore) StoreRoleBindings(ctx context.Context, roleBindings []models.RoleBinding) error {
	storeRoleBindingsPath := "v1/query/rolebindings"
	storeRoleBindingsUrl := fmt.Sprintf("%s/%s", h.url, storeRoleBindingsPath)

	req := &pb.StoreRoleBindingsRequest{
		Rolebindings: convertToPbRoleBinding(roleBindings),
	}

	bodyAsJson, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed json marshalling: %w", err)
	}
	request, err := http.NewRequest("POST", storeRoleBindingsUrl, bytes.NewBuffer(bodyAsJson))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Cookie", fmt.Sprintf("id_token=%s", h.token))

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("failed http request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return fmt.Errorf(fmt.Sprintf("response error: %s", response.Status))
	}

	return nil
}

func (h httpRemoteStore) StoreObjects(ctx context.Context, objects []models.Object) error {
	storeObjectsPath := "v1/query/objects"
	storeObjectsUrl := fmt.Sprintf("%s/%s", h.url, storeObjectsPath)

	req := &pb.StoreObjectsRequest{
		Objects: convertToPbObject(objects),
	}

	bodyAsJson, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed json marshalling: %w", err)
	}
	request, err := http.NewRequest("POST", storeObjectsUrl, bytes.NewBuffer(bodyAsJson))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Cookie", fmt.Sprintf("id_token=%s", h.token))

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("failed http request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return fmt.Errorf(fmt.Sprintf("response error: %s", response.Status))
	}

	return nil
}

func (h httpRemoteStore) DeleteObjects(ctx context.Context, objects []models.Object) error {
	deleteObjectsPath := "v1/query/objects"
	deleteObjectsUrl := fmt.Sprintf("%s/%s", h.url, deleteObjectsPath)

	req := &pb.DeleteObjectsRequest{
		Objects: convertToPbObject(objects),
	}

	bodyAsJson, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed json marshalling: %w", err)
	}
	request, err := http.NewRequest("DELETE", deleteObjectsUrl, bytes.NewBuffer(bodyAsJson))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Cookie", fmt.Sprintf("id_token=%s", h.token))

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("failed http request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return fmt.Errorf(fmt.Sprintf("response error: %s", response.Status))
	}

	return nil
}

func (h httpRemoteStore) GetUrl() string {
	return h.url
}

type RemoteWriterOpts struct {
	Log   logr.Logger
	Url   string
	Token string
}

func NewHttpRemoteStore(opts RemoteWriterOpts) (httpRemoteStore, error) {

	return httpRemoteStore{
		url:   opts.Url,
		token: opts.Token,
		log:   opts.Log,
	}, nil

}

func convertToPbObject(obj []models.Object) []*pb.Object {
	pbObjects := []*pb.Object{}

	for _, o := range obj {
		pbObjects = append(pbObjects, &pb.Object{
			Cluster:    o.Cluster,
			Namespace:  o.Namespace,
			ApiGroup:   o.APIGroup,
			ApiVersion: o.APIVersion,
			Kind:       o.Kind,
			Name:       o.Name,
			Status:     o.Status,
			Message:    o.Message,
		})
	}

	return pbObjects
}

func convertToPbRole(roles []models.Role) []*pb.Role {
	pbRoles := []*pb.Role{}

	for _, role := range roles {
		rules := []*pb.PolicyRule{}
		for _, pr := range role.PolicyRules {
			rule := pb.PolicyRule{
				ApiGroups: pr.APIGroups,
				Resources: pr.Resources,
				Verbs:     pr.Verbs,
				RoleId:    pr.RoleID,
			}
			rules = append(rules, &rule)
		}
		pbRoles = append(pbRoles, &pb.Role{
			Cluster:     role.Cluster,
			Namespace:   role.Namespace,
			Kind:        role.Kind,
			Name:        role.Name,
			PolicyRules: rules,
		})
	}
	return pbRoles
}

func convertToPbRoleBinding(roleBindings []models.RoleBinding) []*pb.RoleBinding {
	pbRoleBindings := []*pb.RoleBinding{}

	for _, roleBinding := range roleBindings {
		subjects := []*pb.Subject{}
		for _, subject := range roleBinding.Subjects {
			subject := pb.Subject{
				Kind:          subject.Kind,
				Name:          subject.Name,
				Namespace:     subject.Namespace,
				ApiGroup:      subject.APIGroup,
				RoleBindingId: subject.RoleBindingID,
			}
			subjects = append(subjects, &subject)
		}
		pbRoleBindings = append(pbRoleBindings, &pb.RoleBinding{
			Cluster:     roleBinding.Cluster,
			Namespace:   roleBinding.Namespace,
			Kind:        roleBinding.Kind,
			Name:        roleBinding.Name,
			RoleRefName: roleBinding.RoleRefName,
			RoleRefKind: roleBinding.RoleRefKind,
			Subjects:    subjects,
		})
	}
	return pbRoleBindings
}
