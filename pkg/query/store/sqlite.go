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

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// dbFile is the name of the sqlite3 database file
const dbFile = "resources.db"

type InMemoryStore struct {
	location string
	db       *gorm.DB
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

func (i InMemoryStore) StoreAccessRules(ctx context.Context, roles []models.AccessRule) error {
	return nil
}

func (i InMemoryStore) StoreObjects(ctx context.Context, objects []models.Object) error {
	for _, object := range objects {
		if err := object.Validate(); err != nil {
			return fmt.Errorf("invalid object: %w", err)
		}
	}

	// Note pointer to slice here.
	// https://gorm.io/docs/create.html#Batch-Insert
	result := i.db.Create(&objects)
	if result.Error != nil {
		return fmt.Errorf("failed to store object: %w", result.Error)
	}

	return nil
}

func (i InMemoryStore) GetObjects(ctx context.Context) ([]models.Object, error) {
	objects := []models.Object{}
	result := i.db.Limit(25).Find(&objects)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to query database: %w", result.Error)
	}

	return objects, nil
}

func (i InMemoryStore) GetAccessRules(ctx context.Context) ([]models.AccessRule, error) {
	return []models.AccessRule{}, nil
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
	cluster text,
	namespace text,
	kind text,
	name text,
	status text,
	message text);
`)

	if err != nil {
		return fmt.Errorf("failed to apply schema: %w", err)
	}
	return err
}

func createDB(cacheLocation string, log logr.Logger) (string, *gorm.DB, error) {
	dbFileLocation := filepath.Join(cacheLocation, dbFile)
	// make sure the directory exists
	if err := os.MkdirAll(cacheLocation, os.ModePerm); err != nil {
		return "", nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(dbFileLocation), &gorm.Config{})
	if err != nil {
		return "", nil, fmt.Errorf("failed to open database: %w", err)
	}
	// From the readme: https://github.com/mattn/go-sqlite3
	goDB, err := db.DB()
	if err != nil {
		return "", nil, fmt.Errorf("failed to golang sql database: %w", err)
	}

	goDB.SetMaxOpenConns(1)

	db.AutoMigrate(&models.Object{})

	return dbFileLocation, db, nil
}
