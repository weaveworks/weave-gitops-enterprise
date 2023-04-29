package sqliterator

import (
	"database/sql"
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"gorm.io/gorm"
)

type iterator struct {
	result *gorm.DB
	rows   *sql.Rows
}

func New(result *gorm.DB) (*iterator, error) {
	rows, err := result.Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	return &iterator{
		result: result,
		rows:   rows,
	}, nil
}

func (i *iterator) Next() bool {
	return i.rows.Next()
}

func (i *iterator) Row() (models.Object, error) {
	var object models.Object

	if err := i.rows.Err(); err != nil {
		return models.Object{}, fmt.Errorf("iterator error: %w", err)
	}

	if err := i.result.ScanRows(i.rows, &object); err != nil {
		return models.Object{}, fmt.Errorf("failed to scan rows: %w", err)
	}

	return object, nil
}

func (i *iterator) All() ([]models.Object, error) {
	var objects []models.Object

	defer i.rows.Close()

	for i.rows.Next() {
		var object models.Object

		if err := i.result.ScanRows(i.rows, &object); err != nil {
			return nil, fmt.Errorf("failed to scan rows: %w", err)
		}

		objects = append(objects, object)
	}

	if err := i.rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	return objects, nil
}

func (i *iterator) Close() error {
	return i.rows.Close()
}
