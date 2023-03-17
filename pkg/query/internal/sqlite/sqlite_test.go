package sqlite

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/gomega"
)

func TestNewSQLiteStore(t *testing.T) {
	g := NewGomegaWithT(t)
	dbDir, err := os.MkdirTemp("", "db")
	g.Expect(err).To(BeNil())

	db, err := CreateDB(dbDir)
	g.Expect(err).To(BeNil())

	sqlDB, err := db.DB()
	g.Expect(err).To(BeNil())

	tests := []struct {
		name        string
		tableName   string
		desiredCols []string
	}{
		{
			name:        "objects table",
			tableName:   "objects",
			desiredCols: []string{"id", "cluster", "namespace", "kind", "name", "status", "message"},
		},
		{
			name:        "access_rules table",
			tableName:   "access_rules",
			desiredCols: []string{"id", "cluster", "namespace", "principal", "accessible_kinds"},
		},
	}

	for _, tt := range tests {
		cols, err := sqlDB.Query(fmt.Sprintf("PRAGMA table_info(%s)", tt.tableName))
		g.Expect(err).To(BeNil())

		var columnNames []string
		for cols.Next() {
			var index int64
			var columnName string
			var dataType interface{}
			var nullable bool
			var defaultVal interface{}
			var autoIncrement bool

			err := cols.Scan(&index, &columnName, &dataType, &nullable, &defaultVal, &autoIncrement)
			g.Expect(err).To(BeNil())

			columnNames = append(columnNames, columnName)
		}
		g.Expect(columnNames).To(ContainElements(tt.desiredCols))
	}
}
