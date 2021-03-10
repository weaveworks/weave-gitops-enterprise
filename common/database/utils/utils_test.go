package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var dbPath = "open-test.db"

func TestOpen(t *testing.T) {
	_, err := Open("/doesnotexist/test.db", "sqlite", "", "", "")
	assert.Error(t, err)

	_, err = Open(dbPath, "sqlite", "", "", "")
	defer os.Remove(dbPath)

	assert.NoError(t, err)
	_, err = os.Stat(dbPath)
	assert.NoError(t, err)

	_, err = Open("", "mysql", "", "", "")
	assert.Error(t, err)

	_, err = Open("nonexistent", "postgres", "postgres", "username", "password")
	assert.Error(t, err)
}

func TestOpenDebug(t *testing.T) {
	_, err := OpenDebug("/doesnotexist/test.db", true)
	assert.Error(t, err)

	_, err = OpenDebug("", true)
	assert.NoError(t, err)
}

func TestMigrateTables(t *testing.T) {
	testDB, err := Open(dbPath, "sqlite", "", "", "")
	defer os.Remove(dbPath)

	assert.NoError(t, err)
	assert.False(t, HasAllTables(testDB))

	err = MigrateTables(testDB)
	assert.NoError(t, err)

	assert.True(t, HasAllTables(testDB))
}
