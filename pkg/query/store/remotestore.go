package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/query"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
)

type RemoteStoreOpts struct {
	Log               logr.Logger
	Address           string
	Token             string
	RemoteStoreWriter RemoteStoreWriter
}

type RemoteStore struct {
	log         logr.Logger
	storeWriter RemoteStoreWriter
}

func (r RemoteStore) StoreAccessRules(ctx context.Context, roles []models.AccessRule) error {
	err := r.storeWriter.StoreAccessRules(ctx, roles)
	if err != nil {
		return fmt.Errorf("cannot remote write access rules: %w", err)
	}
	return nil
}

func (r RemoteStore) StoreObjects(ctx context.Context, objects []models.Object) error {
	err := r.storeWriter.StoreObjects(ctx, objects)
	if err != nil {
		return fmt.Errorf("cannot remote store objects: %w", err)
	}
	return nil
}

func (r RemoteStore) DeleteObjects(ctx context.Context, object []models.Object) error {
	r.log.Info("in delete objects")
	return fmt.Errorf("not implemented")
}

func (r RemoteStore) GetObjects(ctx context.Context, q Query) ([]models.Object, error) {
	r.log.Info("in get objects")
	return nil, fmt.Errorf("not implemented")
}

func (r RemoteStore) GetAccessRules(ctx context.Context) ([]models.AccessRule, error) {
	r.log.Info("in get access rule")
	return nil, fmt.Errorf("not implemented")
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . RemoteStoreWriter
type RemoteStoreWriter interface {
	StoreWriter
	GetUrl() string
}

func NewRemoteStore(opts RemoteStoreOpts) (Store, error) {

	if err := validateOptions(opts); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	writer, err := NewRemoteStoreWriter(opts)
	if err != nil {
		return nil, fmt.Errorf("cannot create remote store writer: %w", err)
	}

	remoteStore := RemoteStore{
		log:         opts.Log,
		storeWriter: writer,
	}

	remoteStore.log.Info("created remote writer", "url", remoteStore.storeWriter.GetUrl())

	return remoteStore, nil
}

func NewRemoteStoreWriter(opts RemoteStoreOpts) (RemoteStoreWriter, error) {
	if opts.RemoteStoreWriter != nil {
		return opts.RemoteStoreWriter, nil
	}

	writer, err := newHttpRemoteStore(opts)
	if err != nil {
		return nil, fmt.Errorf("cannot create http remote writer: %w", err)
	}
	return writer, nil
}

func validateOptions(opts RemoteStoreOpts) error {
	//valid if already using a writer
	if opts.RemoteStoreWriter != nil {
		return nil
	}

	if opts.Address == "" {
		return fmt.Errorf("url cannot be empty")
	}

	if opts.Token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	return nil
}

type httpRemoteStore struct {
	url   string
	token string
	log   logr.Logger
}

func (h httpRemoteStore) StoreAccessRules(ctx context.Context, roles []models.AccessRule) error {

	storeAccessRulesPath := "v1/query/access-rules"
	storeAccessRulesUrl := fmt.Sprintf("%s/%s", h.url, storeAccessRulesPath)

	req := &pb.StoreAccessRulesRequest{
		Rules: convertToPbAccessRule(roles),
	}

	bodyAsJson, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed json marshalling: %w", err)
	}
	request, err := http.NewRequest("POST", storeAccessRulesUrl, bytes.NewBuffer(bodyAsJson))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Cookie", h.token)

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
	request.Header.Set("Cookie", h.token)

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

func (h httpRemoteStore) DeleteObjects(ctx context.Context, object []models.Object) error {
	return fmt.Errorf("not supported")
}

func (h httpRemoteStore) GetUrl() string {
	return h.url
}

func newHttpRemoteStore(opts RemoteStoreOpts) (httpRemoteStore, error) {

	return httpRemoteStore{
		url:   opts.Address,
		token: opts.Token,
		log:   opts.Log,
	}, nil

}

// default remote store writer implementation
// it should get a client to the queyr service endpoint doing the storage
// from the options
func newGrpcRemoteStore(opts RemoteStoreOpts) (RemoteStoreWriter, error) {
	log := opts.Log

	//TODO review
	var dialOpts []grpc.DialOption
	dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.Dial(opts.Address, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to server:%w", err)
	}

	//defer conn.Close()
	queryClient := pb.NewQueryClient(conn)
	return grpcRemoteStore{
		queryClient: queryClient, log: log,
	}, nil
}

type grpcRemoteStore struct {
	serverAddr  string
	queryClient pb.QueryClient
	log         logr.Logger
}

func (g grpcRemoteStore) StoreAccessRules(ctx context.Context, rules []models.AccessRule) error {
	request := &pb.StoreAccessRulesRequest{
		Rules: convertToPbAccessRule(rules),
	}
	_, err := g.queryClient.StoreAccessRules(ctx, request)
	if err != nil {
		return fmt.Errorf("query client store access rules failed: %w", err)
	}
	return nil
}

func (g grpcRemoteStore) StoreObjects(ctx context.Context, objects []models.Object) error {
	request := &pb.StoreObjectsRequest{
		Objects: convertToPbObject(objects),
	}
	_, err := g.queryClient.StoreObjects(ctx, request)
	if err != nil {
		return fmt.Errorf("query client store objects failed: %w", err)
	}
	return nil
}

func (g grpcRemoteStore) DeleteObjects(ctx context.Context, object []models.Object) error {
	return fmt.Errorf("not implemented delete objects")
}

func (g grpcRemoteStore) GetUrl() string {
	return g.serverAddr
}

func convertToPbObject(obj []models.Object) []*pb.Object {
	pbObjects := []*pb.Object{}

	for _, o := range obj {
		pbObjects = append(pbObjects, &pb.Object{
			Kind:      o.Kind,
			Name:      o.Name,
			Namespace: o.Namespace,
			Cluster:   o.Cluster,
			Status:    o.Status,
		})
	}

	return pbObjects
}

func convertToPbAccessRule(rules []models.AccessRule) []*pb.AccessRule {
	pbRules := []*pb.AccessRule{}

	for _, r := range rules {
		rule := &pb.AccessRule{
			Principal:       r.Principal,
			Namespace:       r.Namespace,
			Cluster:         r.Cluster,
			AccessibleKinds: []string{},
		}

		rule.AccessibleKinds = append(rule.AccessibleKinds, r.AccessibleKinds...)

		pbRules = append(pbRules, rule)

	}
	return pbRules
}
