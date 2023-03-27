package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	pbp "github.com/weaveworks/weave-gitops-enterprise/pkg/api/pipelines"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/query"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"net/http"
)

type RemoteStore struct {
	log    logr.Logger
	url    string
	token  string
	client *http.Client
}

func (h RemoteStore) GetObjects(ctx context.Context, q Query, opts QueryOption) ([]models.Object, error) {

	ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("Cookie", h.token))

	// Replace the address with your own
	conn, err := grpc.Dial(h.url, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	pipelines := pbp.NewPipelinesClient(conn)
	req := &pbp.ListPipelinesRequest{}

	resp, err := pipelines.ListPipelines(ctx, req)
	if err != nil {
		return nil, err
	}
	h.log.Info("response pipelines", "resp", resp.String())
	return nil, nil
}

func (h RemoteStore) GetAccessRules(ctx context.Context) ([]models.AccessRule, error) {
	h.log.Info("get objects not supported for remote stores")
	return nil, nil
}

// TODO add test
func (h RemoteStore) DeleteRoles(ctx context.Context, roles []models.Role) error {
	deleteRolesPath := "v1/query/roles"
	deleteRolesUrl := fmt.Sprintf("%s/%s", h.url, deleteRolesPath)

	req := &pb.DeleteRolesRequest{
		Roles: convertToPbRole(roles),
	}

	bodyAsJson, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed json marshalling: %w", err)
	}
	request, err := http.NewRequest("DELETE", deleteRolesUrl, bytes.NewBuffer(bodyAsJson))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Cookie", fmt.Sprintf("id_token=%s", h.token))

	response, err := h.client.Do(request)
	if err != nil {
		return fmt.Errorf("failed http request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return fmt.Errorf(fmt.Sprintf("response error: %s", response.Status))
	}

	return nil
}

// TODO add test
func (h RemoteStore) DeleteRoleBindings(ctx context.Context, roleBindings []models.RoleBinding) error {
	deleteRoleBindingsPath := "v1/query/rolebindings"
	deleteRolesBindingUrl := fmt.Sprintf("%s/%s", h.url, deleteRoleBindingsPath)

	req := &pb.DeleteRoleBindingsRequest{
		Rolebindings: convertToPbRoleBinding(roleBindings),
	}

	bodyAsJson, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed json marshalling: %w", err)
	}
	request, err := http.NewRequest("DELETE", deleteRolesBindingUrl, bytes.NewBuffer(bodyAsJson))
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

// TODO refactor and make it more reliable, timeouts, retries!
func (h RemoteStore) StoreRoles(ctx context.Context, roles []models.Role) error {
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

func (h RemoteStore) StoreRoleBindings(ctx context.Context, roleBindings []models.RoleBinding) error {
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

func (h RemoteStore) StoreObjects(ctx context.Context, objects []models.Object) error {

	ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("Cookie", h.token))

	// Replace the address with your own
	conn, err := grpc.Dial(h.url, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	queryClient := pb.NewQueryClient(conn)
	req := &pb.StoreObjectsRequest{
		Objects: convertToPbObject(objects),
	}

	resp, err := queryClient.StoreObjects(ctx, req)
	if err != nil {
		return err
	}
	h.log.Info("response store objects", "resp", resp.String())
	return nil
}

func (h RemoteStore) DeleteObjects(ctx context.Context, objects []models.Object) error {
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

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . RemoteStoreWriter
type RemoteStoreWriter interface {
	StoreWriter
	GetUrl() string
}

func newRemoteStoreWithClient(opts StoreOpts, client *http.Client) (Store, error) {
	if client == nil {
		return nil, fmt.Errorf("invalid client")
	}
	remoteStore := RemoteStore{
		log:    opts.Log,
		client: client,
	}
	return remoteStore, nil
}

func newRemoteStore(opts StoreOpts) (Store, error) {
	if err := validateOptions(opts); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	remoteStore := RemoteStore{
		log:    opts.Log,
		url:    opts.Url,
		token:  opts.Token,
		client: &http.Client{},
	}

	remoteStore.log.Info("remote store created", "url", opts.Url)

	return remoteStore, nil
}

func validateOptions(opts StoreOpts) error {
	if opts.Url == "" {
		return fmt.Errorf("url cannot be empty")
	}

	if opts.Token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	return nil
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
