package store

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"os"
	"path/filepath"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

// dbFile is the name of the sqlite3 database file
const dbFile = "resources.db"

type InMemoryStore struct {
	location string
	db       *sql.DB
	log      logr.Logger
}

func newInMemoryStore(location string, log logr.Logger) (*InMemoryStore, error) {
	if location == "" {
		return nil, fmt.Errorf("invalid location")
	}
	dbLocation, db, err := createDB(location, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache database: %w", err)
	}
	return &InMemoryStore{
		location: dbLocation,
		db:       db,
		log:      log,
	}, nil
}

// TODO batch
func (i InMemoryStore) StoreAccessRules(ctx context.Context, roles []models.AccessRule) error {
	for _, role := range roles {
		_, err := i.StoreAccessRule(ctx, role)
		//TODO capture error
		if err != nil {
			i.log.Error(err, "could not store object")
			return err
		}
	}
	return nil
}

func (i InMemoryStore) StoreAccessRule(ctx context.Context, roles models.AccessRule) (int64, error) {
	if ctx == nil {
		return -1, fmt.Errorf("invalid context")
	}

	sqlStatement := `INSERT INTO documents (name, namespace, kind) VALUES ($1, $2,$3)`
	result, err := i.db.ExecContext(
		ctx,
		sqlStatement, roles.Principal, roles.Namespace, roles.Cluster)
	if err != nil {
		return -1, err
	}
	return result.LastInsertId()
}

// TODO batch
func (i InMemoryStore) StoreObjects(ctx context.Context, objects []models.Object) error {

	for _, object := range objects {
		_, err := i.StoreObject(ctx, object)
		//TODO capture error
		if err != nil {
			i.log.Error(err, "could not store object")
			return err
		}

	}

	return nil
}

func (i InMemoryStore) StoreObject(ctx context.Context, object models.Object) (int64, error) {
	if ctx == nil {
		return -1, fmt.Errorf("invalid context")
	}

	if object.Name == "" || object.Kind == "" {
		return -1, fmt.Errorf("invalid object")
	}

	sqlStatement := `INSERT INTO documents (name, namespace, kind) VALUES ($1, $2,$3)`
	result, err := i.db.ExecContext(
		ctx,
		sqlStatement, object.Name, object.Namespace, object.Kind)
	if err != nil {
		return -1, err
	}
	return result.LastInsertId()
}

func (i InMemoryStore) GetObjects() ([]models.Object, error) {
	return []models.Object{}, fmt.Errorf("not implemented yet")
}

func (i InMemoryStore) GetAccessRules() ([]models.AccessRule, error) {
	return []models.AccessRule{}, fmt.Errorf("not implemented yet")
}

// TODO add unit tests
func (i InMemoryStore) CountObjects(ctx context.Context, kind string) (int64, error) {
	rows, err := i.db.QueryContext(ctx, "SELECT COUNT(*) FROM documents WHERE kind=?", kind)
	if err != nil {
		return 0, fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		var n int64
		if err := rows.Scan(&n); err != nil {
			return 0, fmt.Errorf("failed to scan database: %w", err)
		}
		count += n
	}
	return count, nil
}

func (i InMemoryStore) DeleteObject(ctx context.Context, object models.Object) error {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryStore) GetLocation() string {
	return i.location
}

func applySchema(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS documents (
	name text,
	namespace text,
	kind text);`)
	return err
}

func createDB(cacheLocation string, log logr.Logger) (string, *sql.DB, error) {
	dbFileLocation := filepath.Join(cacheLocation, dbFile)
	log.Info("db path", dbFileLocation)
	// make sure the directory exists
	if err := os.MkdirAll(cacheLocation, os.ModePerm); err != nil {
		return "", nil, fmt.Errorf("failed to create cache directory: %w", err)
	}
	db, err := sql.Open("sqlite3", dbFileLocation)
	if err != nil {
		return "", nil, fmt.Errorf("failed to open database: %q, %w", cacheLocation, err)
	}
	log.Info("db created")
	// From the readme: https://github.com/mattn/go-sqlite3
	db.SetMaxOpenConns(1)
	if err := applySchema(db); err != nil {
		return "", db, err
	}
	log.Info("schema created")
	return dbFileLocation, db, nil
}
