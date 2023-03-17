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

	rows, err := sqlDB.Query("SELECT name FROM sqlite_master;")
	g.Expect(err).To(BeNil())

	defer rows.Close()

	result := []string{}

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		g.Expect(err).To(BeNil())
		result = append(result, name)
	}

	g.Expect(result).To(ContainElement("objects"))
	g.Expect(result).To(ContainElement("access_rules"))

	cols, err := sqlDB.Query(fmt.Sprintf("PRAGMA table_info(%s)", "objects"))
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

	g.Expect(columnNames).To(ContainElements("id", "cluster", "namespace", "kind", "name"))

}
