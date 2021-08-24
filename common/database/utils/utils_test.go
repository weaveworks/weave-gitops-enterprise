package utils

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/wks/common/database/models"
	"gorm.io/gorm"
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

func TestCascadeDelete(t *testing.T) {
	db, err := OpenDebug("", true)
	assert.NoError(t, err)
	err = MigrateTables(db)
	assert.NoError(t, err)

	// Create cluster
	clusterWithPullRequest := models.Cluster{Name: "foo", PullRequests: []*models.PullRequest{{URL: "http://example.com"}}}
	result := db.Create(&clusterWithPullRequest)
	assert.NoError(t, result.Error)

	// Join row created okay
	joinRow := &ClusterPullRequests{}
	result = db.First(&joinRow)
	assert.NoError(t, result.Error)
	assert.Equal(t, 1, joinRow.ClusterID)
	assert.Equal(t, 1, joinRow.PullRequestID)

	// Delete cluster
	result = db.Delete(&models.Cluster{}, 1)
	assert.NoError(t, result.Error)

	// Make sure join table is cleaned up!
	joinRow = &ClusterPullRequests{}
	result = db.First(&joinRow)
	assert.Error(t, result.Error)
	assert.True(t, errors.Is(result.Error, gorm.ErrRecordNotFound), "Error was something other than not found! %v", result.Error)
}

// So we can query the able usually managed by gorm
type ClusterPullRequests struct {
	ClusterID     int `gorm:"primaryKey"`
	PullRequestID int `gorm:"primaryKey"`
}
