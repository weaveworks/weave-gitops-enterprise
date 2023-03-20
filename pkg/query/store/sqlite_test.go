package store

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

	db, err := CreateSQLiteDB(dbDir)
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
			name:        "role_bindings table",
			tableName:   "role_bindings",
			desiredCols: []string{"id", "cluster", "namespace", "kind", "name", "role_ref_name", "role_ref_kind"},
		},
		{
			name:        "roles table",
			tableName:   "roles",
			desiredCols: []string{"id", "cluster", "namespace", "kind", "name"},
		},
		{
			name:        "subjects table",
			tableName:   "subjects",
			desiredCols: []string{"id", "namespace", "kind", "name", "role_binding_id"},
		},
		{
			name:        "policy_rules table",
			tableName:   "policy_rules",
			desiredCols: []string{"id", "role_id", "api_groups", "resources", "verbs"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
		})

	}
}
