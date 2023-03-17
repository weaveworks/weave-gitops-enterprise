package sqlite

import (
	"database/sql"
	"os"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/go-logr/logr/testr"
)

func TestNewSQLiteStore(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)
	dbDir, err := os.MkdirTemp("", "db")
	g.Expect(err).To(BeNil())

	_, loc, err := NewStore(dbDir, log)
	g.Expect(err).To(BeNil())

	db, err := sql.Open("sqlite3", loc)
	g.Expect(err).To(BeNil())
	defer db.Close()

	rows, err := db.Query("SELECT name FROM sqlite_master;")
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

}
