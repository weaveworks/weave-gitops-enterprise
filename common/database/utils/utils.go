package utils

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/wks/common/database/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB ref contains a pointer to the MCCP database
var DB *gorm.DB

// Open creates the SQLite database or connects to an existing database
func Open(dbURI string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbURI), &gorm.Config{})
	if err != nil {
		return nil, errors.New("failed to connect to database")
	}
	// Set the global database Ref
	DB = db
	return db, nil
}

// MigrateTables creates the database tables given a gorm.DB
func MigrateTables(db *gorm.DB) error {
	// Migrate the schema
	err := db.AutoMigrate(&models.Event{})
	if err != nil {
		return errors.New("failed to create Events table")
	}
	log.Info("created Events table")

	err = db.AutoMigrate(&models.Cluster{})
	if err != nil {
		return errors.New("failed to create Clusters table")
	}
	log.Info("created Cluster table")

	err = db.AutoMigrate(&models.ClusterInfo{})
	if err != nil {
		return errors.New("failed to create ClusterInfo table")
	}
	log.Info("created ClusterInfo table")

	err = db.AutoMigrate(&models.NodeInfo{})
	if err != nil {
		return errors.New("failed to create NodeInfo table")
	}
	log.Info("created NodeInfo table")

	err = db.AutoMigrate(&models.Alert{})
	if err != nil {
		return errors.New("failed to create Alert table")
	}
	log.Info("created Alert table")

	err = db.AutoMigrate(&models.GitRepository{})
	if err != nil {
		return errors.New("failed to create GitRepository table")
	}
	log.Info("created GitRepository table")

	err = db.AutoMigrate(&models.GitProvider{})
	if err != nil {
		return errors.New("failed to create GitProviders table")
	}
	log.Info("created GitProviders table")

	err = db.AutoMigrate(&models.Workspace{})
	if err != nil {
		return errors.New("failed to create Workspaces table")
	}
	log.Info("created Workspaces table")

	err = db.AutoMigrate(&models.FluxInfo{})
	if err != nil {
		return errors.New("failed to create FluxInfo table")
	}
	log.Info("created FluxInfo table")
	return nil
}

// HasAllTables return true if the given DB has all tables defined in the models
func HasAllTables(db *gorm.DB) bool {
	if db.Migrator().HasTable(&models.Event{}) &&
		db.Migrator().HasTable(&models.Cluster{}) &&
		db.Migrator().HasTable(&models.ClusterInfo{}) &&
		db.Migrator().HasTable(&models.NodeInfo{}) &&
		db.Migrator().HasTable(&models.Alert{}) &&
		db.Migrator().HasTable(&models.GitProvider{}) &&
		db.Migrator().HasTable(&models.GitRepository{}) &&
		db.Migrator().HasTable(&models.FluxInfo{}) &&
		db.Migrator().HasTable(&models.Workspace{}) {
		return true
	}
	return false
}
