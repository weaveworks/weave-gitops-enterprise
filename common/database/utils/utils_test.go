package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

	assert.True(t, HasAllTables(testDB))
}
