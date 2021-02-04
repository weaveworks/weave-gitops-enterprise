package utils

import (
	"os"
	"testing"

	"github.com/tj/assert"
	"github.com/weaveworks/wks/common/database/models"
)

var dbPath = "open-test.db"

func TestOpen(t *testing.T) {
	_, err := Open(dbPath)
	defer os.Remove(dbPath)

	assert.NoError(t, err)
	_, err = os.Stat(dbPath)
	assert.NoError(t, err)
}

func TestMigrateTables(t *testing.T) {
	testDB, err := Open(dbPath)
	defer os.Remove(dbPath)

	assert.NoError(t, err)

	err = MigrateTables(testDB)
	assert.NoError(t, err)

	testDB.Migrator().HasTable(&models.Event{})
	testDB.Migrator().HasTable(&models.Cluster{})
	testDB.Migrator().HasTable(&models.GitProvider{})
	testDB.Migrator().HasTable(&models.GitRepository{})
	testDB.Migrator().HasTable(&models.Workspace{})
}
