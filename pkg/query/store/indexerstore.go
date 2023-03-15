package store

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	_ "github.com/mattn/go-sqlite3"
	"github.com/meilisearch/meilisearch-go"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"time"
)

type IndexerStore struct {
	log       logr.Logger
	client    *meilisearch.Client
	indexName string
}

// TODO extract indexName
func newIndexerStore(url string, log logr.Logger) (*IndexerStore, error) {
	if url == "" {
		return nil, fmt.Errorf("invalid url")
	}

	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host: url,
	})

	if client == nil {
		return nil, fmt.Errorf("could not create indexer client")
	}

	log.Info("created indexer store", "store", url)
	return &IndexerStore{
		client: client,
		log:    log,

		indexName: "applications",
	}, nil
}

func (i IndexerStore) StoreAccessRules(roles []models.AccessRule) error {
	return fmt.Errorf("not implemented yet")
}

func (i IndexerStore) StoreObjects(ctx context.Context, objects []models.Object) error {

	if objects == nil {
		return fmt.Errorf("invalid objects")
	}

	//TODO use batch addition
	for _, object := range objects {
		_, err := i.StoreObject(ctx, object)
		//TODO capture error
		if err != nil {
			i.log.Error(err, "could not store objects")
			return err
		}

	}

	return nil
}

func (i IndexerStore) StoreObject(ctx context.Context, object models.Object) (int64, error) {

	if ctx == nil {
		return -1, fmt.Errorf("invalid context")
	}

	if object.Namespace == "" || object.Name == "" || object.Kind == "" {
		return -1, fmt.Errorf("invalid object")
	}

	if object.Id == "" {
		object.Id = fmt.Sprintf("%s-%s-%s", object.Cluster, object.Namespace, object.Name)
		i.log.Info("setting default id", "id", object.Id)
	}
	var objectAsMap map[string]interface{}
	data, _ := json.Marshal(object)
	json.Unmarshal(data, &objectAsMap)
	objects := []map[string]interface{}{
		objectAsMap,
	}
	taskInfo, err := i.client.Index(i.indexName).AddDocuments(objects)
	if err != nil {
		return 0, fmt.Errorf("could not add documents:%w", err)
	}
	//TODO improve this piece
	i.log.Info("submitted request with id", "taskuid", taskInfo.TaskUID)
	for {
		task, err := i.client.GetTask(taskInfo.TaskUID)
		if err != nil {
			return 0, fmt.Errorf("could not add documents:%w", err)
		}
		i.log.Info("task info", "status", task.Status, "error", task.Error.Message)

		if task.Status == meilisearch.TaskStatusFailed || task.Status == meilisearch.TaskStatusSucceeded {
			break
		}

		time.Sleep(1 * time.Second)
	}

	return taskInfo.TaskUID, nil
}

func (i IndexerStore) GetObjects() ([]models.Object, error) {
	return []models.Object{}, fmt.Errorf("not implemented yet")
}

func (i IndexerStore) GetAccessRules() ([]models.AccessRule, error) {
	return []models.AccessRule{}, fmt.Errorf("not implemented yet")
}

// TODO add unit tests
func (i IndexerStore) CountObjects(ctx context.Context, kind string) (int64, error) {

	return -1, fmt.Errorf("not implememted yet")
}

func (i IndexerStore) DeleteObject(ctx context.Context, object models.Object) error {
	//TODO implement me
	panic("implement me")
}
