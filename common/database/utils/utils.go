package utils

import (
	"errors"
	"fmt"
	"net/url"

	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/wks/common/database/models"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB ref contains a pointer to the MCCP database
var DB *gorm.DB

// Open creates the database or connects to an existing database
func Open(uri, dbType, dbName, user, password string) (*gorm.DB, error) {
	if dbType == "postgres" {
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable TimeZone=UTC", uri, user, password, dbName)
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to database %s with provided username and password", uri)
		}
		DB = db
		return db, nil
	} else if dbType == "sqlite" {
		return OpenDebug(uri, false)
	}
	return nil, fmt.Errorf("unsupported database type %s", dbType)
}

func OpenDebug(dbURI string, debug bool) (*gorm.DB, error) {
	config := &gorm.Config{}
	if debug {
		config = &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		}
	}
	db, err := gorm.Open(sqlite.Open(dbURI), config)

	if err != nil {
		return nil, errors.New("failed to connect to database")
	}
	// Set the global database Ref
	DB = db
	return db, nil
}

func GetSqliteUri(uri, dbBusyTimeout string) (string, error) {
	// Don't set for in mem
	if uri == "" {
		return "", nil
	}

	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	// if _busy_timeout not already set
	if u.Query().Get("_busy_timeout") == "" {
		values := u.Query()
		values.Set("_busy_timeout", dbBusyTimeout)
		u.RawQuery = values.Encode()
	}
	return u.String(), nil
}

// MigrateTables creates the database tables given a gorm.DB
func MigrateTables(db *gorm.DB) error {
	// Migrate the schema
	err := db.AutoMigrate(
		&models.Event{},
		&models.Cluster{},
		&models.ClusterStatus{},
		&models.ClusterInfo{},
		&models.NodeInfo{},
		&models.Alert{},
		&models.Workspace{},
		&models.FluxInfo{},
		&models.GitCommit{},
		&models.CAPICluster{},
		&models.PullRequest{},
		&models.PRCluster{},
	)
	if err != nil {
		return errors.New("failed to create tables")
	}

	clusterStatuses := []*models.ClusterStatus{
		{
			ID:     1,
			Status: "critical",
		},
		{
			ID:     2,
			Status: "alerting",
		},
		{
			ID:     3,
			Status: "lastSeen",
		},
		{
			ID:     4,
			Status: "pullRequestCreated",
		},
		{
			ID:     5,
			Status: "clusterFound",
		},
		{
			ID:     6,
			Status: "ready",
		},
		{
			ID:     7,
			Status: "notConnected",
		},
	}
	db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status"}),
	}).Create(clusterStatuses)

	log.Info("created all database tables")
	return nil
}

// HasAllTables return true if the given DB has all tables defined in the models
func HasAllTables(db *gorm.DB) bool {
	if db.Migrator().HasTable(&models.Event{}) &&
		db.Migrator().HasTable(&models.Cluster{}) &&
		db.Migrator().HasTable(&models.ClusterInfo{}) &&
		db.Migrator().HasTable(&models.NodeInfo{}) &&
		db.Migrator().HasTable(&models.Alert{}) &&
		db.Migrator().HasTable(&models.FluxInfo{}) &&
		db.Migrator().HasTable(&models.Workspace{}) &&
		db.Migrator().HasTable(&models.CAPICluster{}) &&
		db.Migrator().HasTable(&models.PullRequest{}) &&
		db.Migrator().HasTable(&models.PRCluster{}) {
		return true
	}
	return false
}
