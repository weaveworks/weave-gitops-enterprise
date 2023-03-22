package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// dbFile is the name of the sqlite3 database file
const dbFile = "resources.db"

type SQLiteStore struct {
	db *gorm.DB
}

func NewSQLiteStore(db *gorm.DB) (*SQLiteStore, error) {
	return &SQLiteStore{
		db: db,
	}, nil
}

func (i *SQLiteStore) StoreRoles(ctx context.Context, roles []models.Role) error {
	if len(roles) == 0 {
		return fmt.Errorf("empty role list")
	}

	rows := []models.Role{}

	for _, role := range roles {
		if err := role.Validate(); err != nil {
			return fmt.Errorf("invalid role: %w", err)
		}

		role.ID = role.GetID()
		rows = append(rows, role)
	}

	clauses := i.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "id"},
		},
		UpdateAll: true,
	})

	result := clauses.Create(&rows)

	return result.Error
}

func (i *SQLiteStore) StoreRoleBindings(ctx context.Context, roleBindings []models.RoleBinding) error {
	if len(roleBindings) == 0 {
		return fmt.Errorf("empty role binding list")
	}

	rows := []models.RoleBinding{}

	for _, roleBinding := range roleBindings {
		if err := roleBinding.Validate(); err != nil {
			return fmt.Errorf("invalid role binding: %w", err)
		}

		roleBinding.ID = roleBinding.GetID()
		rows = append(rows, roleBinding)
	}

	clauses := i.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "id"},
		},
		UpdateAll: true,
	})

	result := clauses.Create(&rows)

	return result.Error
}

func (i *SQLiteStore) StoreObjects(ctx context.Context, objects []models.Object) error {
	// Gotta copy the objects because we need to set the ID
	rows := []models.Object{}

	for _, object := range objects {
		if err := object.Validate(); err != nil {
			return fmt.Errorf("invalid object: %w", err)
		}

		object.ID = object.GetID()
		rows = append(rows, object)
	}

	clauses := i.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "id"},
		},
		UpdateAll: true,
	})

	result := clauses.Create(rows)
	if result.Error != nil {
		return fmt.Errorf("failed to store object: %w", result.Error)
	}

	return nil
}

func toSQLOperand(op QueryOperand) (string, error) {
	switch op {
	case OperandEqual:
		return "=", nil
	case OperandNotEqual:
		return "!=", nil
	default:
		return "", fmt.Errorf("unsupported operand: %s", op)
	}
}

func (i *SQLiteStore) GetObjects(ctx context.Context, q Query, opts QueryOption) ([]models.Object, error) {
	objects := []models.Object{}

	// If limit or offset are zero, they were not set.
	// -1 tells GORM to ignore the limit/offset
	var limit int = -1
	var offset int = -1

	if opts != nil {
		if opts.GetLimit() != 0 {
			limit = int(opts.GetLimit())
		}
		if opts.GetOffset() != 0 {
			offset = int(opts.GetOffset())
		}
	}

	tx := i.db.Limit(limit)
	tx = tx.Offset(offset)

	if q != nil && len(q) > 0 {
		for _, c := range q {

			if c.GetKey() == "" {
				continue
			}

			val := c.GetValue()
			op, err := toSQLOperand(QueryOperand(c.GetOperand()))
			if err != nil {
				return nil, err
			}

			queryString := fmt.Sprintf("%s %s ?", c.GetKey(), op)
			tx = tx.Where(queryString, val)
		}
	}

	result := tx.Find(&objects)

	return objects, result.Error
}

func (i *SQLiteStore) GetAccessRules(ctx context.Context) ([]models.AccessRule, error) {
	roles := []models.Role{}
	bindings := []models.RoleBinding{}

	result := i.db.Model(&models.Role{}).Preload("PolicyRules").Find(&roles)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get roles: %w", result.Error)
	}

	result = i.db.Model(&models.RoleBinding{}).Preload("Subjects").Find(&bindings)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get role bindings: %w", result.Error)
	}

	rules := DeriveAccessRules(roles, bindings)

	return rules, result.Error
}

func (i *SQLiteStore) DeleteObjects(ctx context.Context, objects []models.Object) error {
	for _, object := range objects {
		if err := object.Validate(); err != nil {
			return fmt.Errorf("invalid object: %w", err)
		}

		where := i.db.Where(
			"id = ? ",
			object.GetID(),
		)
		result := i.db.Unscoped().Delete(&models.Object{}, where)
		if result.Error != nil {
			return fmt.Errorf("failed to delete object: %w", result.Error)
		}
	}

	return nil
}

func CreateSQLiteDB(path string) (*gorm.DB, error) {
	dbFileLocation := filepath.Join(path, dbFile)
	// make sure the directory exists
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(dbFileLocation), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	goDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to golang sql database: %w", err)
	}

	// From the readme: https://github.com/mattn/go-sqlite3
	goDB.SetMaxOpenConns(1)

	if err := db.AutoMigrate(&models.Object{}, &models.Role{}, &models.Subject{}, &models.RoleBinding{}, &models.PolicyRule{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}
