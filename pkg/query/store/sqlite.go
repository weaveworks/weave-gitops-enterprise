package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"

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

func newSQLiteStore(location string, log logr.Logger) (*InMemoryStore, error) {
	if location == "" {
		return nil, fmt.Errorf("invalid location")
	}
	dbLocation, db, err := createDB(location, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
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

func (i InMemoryStore) GetObjects(ctx context.Context) ([]models.Object, error) {
	sqlStatement := `SELECT name,namespace,kind FROM documents`

	rows, err := i.db.QueryContext(ctx, sqlStatement)
	if err != nil {
		return []models.Object{}, fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()

	var objects []models.Object
	for rows.Next() {
		var object models.Object
		if err := rows.Scan(&object.Name, &object.Namespace, &object.Kind); err != nil {
			return nil, fmt.Errorf("failed to scan database: %w", err)
		}
		objects = append(objects, object)
	}

	return objects, nil
}

func (i InMemoryStore) GetAccessRules(ctx context.Context) ([]models.AccessRule, error) {
	return []models.AccessRule{}, nil
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
CREATE TABLE IF NOT EXISTS objects (
	name text,
	namespace text,
	kind text);`)

	if err != nil {
		return fmt.Errorf("failed to create documents table: %w", err)
	}

	return err
}

func createDB(cacheLocation string, log logr.Logger) (string, *sql.DB, error) {
	dbFileLocation := filepath.Join(cacheLocation, dbFile)
	// make sure the directory exists
	if err := os.MkdirAll(cacheLocation, os.ModePerm); err != nil {
		return "", nil, fmt.Errorf("failed to create cache directory: %w", err)
	}
	db, err := sql.Open("sqlite3", dbFileLocation)
	if err != nil {
		return "", nil, fmt.Errorf("failed to open database: %q, %w", cacheLocation, err)
	}
	// From the readme: https://github.com/mattn/go-sqlite3
	db.SetMaxOpenConns(1)
	if err := applySchema(db); err != nil {
		return "", db, err
	}

	return dbFileLocation, db, nil
}
