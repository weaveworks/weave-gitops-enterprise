package sqlite

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"

	_ "github.com/mattn/go-sqlite3"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// dbFile is the name of the sqlite3 database file
const dbFile = "resources.db"

type SQLiteStore struct {
	location string
	db       *gorm.DB
	log      logr.Logger
}

func NewStore(location string, log logr.Logger) (*SQLiteStore, string, error) {
	if location == "" {
		return nil, "", fmt.Errorf("invalid location")
	}
	dbLocation, db, err := createDB(location, log)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create database: %w", err)
	}
	return &SQLiteStore{
		location: dbLocation,
		db:       db,
		log:      log,
	}, dbLocation, nil
}

func (i SQLiteStore) StoreAccessRules(ctx context.Context, rules []models.AccessRule) error {
	for _, rule := range rules {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("invalid access rule: %w", err)
		}
	}

	result := i.db.Create(&rules)

	return result.Error
}

func (i SQLiteStore) StoreObjects(ctx context.Context, objects []models.Object) error {
	for _, object := range objects {
		if err := object.Validate(); err != nil {
			return fmt.Errorf("invalid object: %w", err)
		}
	}

	result := i.db.Create(objects)
	if result.Error != nil {
		return fmt.Errorf("failed to store object: %w", result.Error)
	}

	return nil
}

func (i SQLiteStore) GetObjects(ctx context.Context) ([]models.Object, error) {
	objects := []models.Object{}
	result := i.db.Find(&objects)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to query database: %w", result.Error)
	}

	return objects, nil
}

func (i SQLiteStore) GetAccessRules(ctx context.Context) ([]models.AccessRule, error) {
	rules := []models.AccessRule{}

	result := i.db.Find(&rules)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query database: %w", result.Error)
	}

	return rules, nil
}

func (i SQLiteStore) DeleteObject(ctx context.Context, object models.Object) error {
	//TODO implement me
	panic("implement me")
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

	goDB, err := db.DB()
	if err != nil {
		return "", nil, fmt.Errorf("failed to golang sql database: %w", err)
	}

	// From the readme: https://github.com/mattn/go-sqlite3
	goDB.SetMaxOpenConns(1)

	if err := db.AutoMigrate(&models.Object{}, &models.AccessRule{}); err != nil {
		return "", nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return dbFileLocation, db, nil
}
