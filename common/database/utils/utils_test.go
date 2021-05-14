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

func TestGetSqliteUri(t *testing.T) {
	urlTests := []struct {
		description string
		uri         string
		busyTimeout string
		expected    string
		err         error
	}{
		{"basic", "/db/mccp.db", "5000", "/db/mccp.db?_busy_timeout=5000", nil},
		{"already there", "/db/mccp.db?_busy_timeout=10000", "5000", "/db/mccp.db?_busy_timeout=10000", nil},
		{"preserves existing", "/db/mccp.db?foo=bar", "5000", "/db/mccp.db?_busy_timeout=5000&foo=bar", nil},
		{"don't set for in memory", "", "5000", "", nil},
	}

	for _, tt := range urlTests {
		t.Run(tt.description, func(t *testing.T) {
			result, err := GetSqliteUri(tt.uri, tt.busyTimeout)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.err, err)
		})
	}
}
