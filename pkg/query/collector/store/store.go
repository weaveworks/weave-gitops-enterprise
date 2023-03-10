package store

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"os"
	"path/filepath"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type Document struct {
	Name      string
	Namespace string
	Kind      string
}

type StoreWriter interface {
	Add(ctx context.Context, document Document) (int64, error)
	Delete(ctx context.Context, document Document) error
}

type StoreReader interface {
	Count(ctx context.Context, kind string) (int64, error)
	GetAll(ctx context.Context) ([]Document, error)
}

type Store interface {
	StoreReader
	StoreWriter
}

// dbFile is the name of the sqlite3 database file
const dbFile = "charts.db"

type InMemoryStore struct {
	location string
	db       *sql.DB
	log      logr.Logger
}

// TODO add unit tests
func (i InMemoryStore) Count(ctx context.Context, kind string) (int64, error) {
	rows, err := i.db.QueryContext(ctx, "SELECT COUNT(*) FROM documents WHERE kind=?", kind)
	if err != nil {
		return 0, fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		var n int64
		if err := rows.Scan(&n); err != nil {
			return 0, fmt.Errorf("failed to scan database: %w", err)
		}
		count += n
	}

	return count, nil
}

func (i InMemoryStore) GetAll(ctx context.Context) ([]Document, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryStore) Add(ctx context.Context, document Document) (int64, error) {
	if ctx == nil {
		return -1, fmt.Errorf("invalid context")
	}

	if document.Namespace == "" || document.Name == "" || document.Kind == "" {
		return -1, fmt.Errorf("invalid document")
	}

	sqlStatement := `INSERT INTO documents (name, namespace, kind) VALUES ($1, $2,$3)`
	result, err := i.db.ExecContext(
		ctx,
		sqlStatement, document.Name, document.Namespace, document.Kind)
	if err != nil {
		return -1, err
	}
	return result.LastInsertId()
}

func (i InMemoryStore) Delete(ctx context.Context, document Document) error {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryStore) GetLocation() string {
	return i.location
}

func NewInMemoryStore(location string, log logr.Logger) (*InMemoryStore, error) {
	if location == "" {
		return nil, fmt.Errorf("invalid location")
	}
	dbLocation, db, err := createDB(location, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache database: %w", err)
	}
	return &InMemoryStore{
		location: dbLocation,
		db:       db,
		log:      log,
	}, nil
}

func applySchema(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS documents (
	name text,
	namespace text,
	kind text);`)
	return err
}

func createDB(cacheLocation string, log logr.Logger) (string, *sql.DB, error) {
	dbFileLocation := filepath.Join(cacheLocation, dbFile)
	log.Info("db path", dbFileLocation)
	// make sure the directory exists
	if err := os.MkdirAll(cacheLocation, os.ModePerm); err != nil {
		return "", nil, fmt.Errorf("failed to create cache directory: %w", err)
	}
	db, err := sql.Open("sqlite3", dbFileLocation)
	if err != nil {
		return "", nil, fmt.Errorf("failed to open database: %q, %w", cacheLocation, err)
	}
	log.Info("db created")
	// From the readme: https://github.com/mattn/go-sqlite3
	db.SetMaxOpenConns(1)
	if err := applySchema(db); err != nil {
		return "", db, err
	}
	log.Info("schema created")
	return dbFileLocation, db, nil
}
