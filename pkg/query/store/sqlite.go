package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/weaveworks/weave-gitops/core/logger"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/sqliterator"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// dbFile is the name of the sqlite3 database file
const dbFile = "resources.db"

type SQLiteStore struct {
	db    *gorm.DB
	log   logr.Logger
	debug logr.Logger
}

func (i *SQLiteStore) DeleteAllRoles(ctx context.Context, clusters []string) error {
	for _, cluster := range clusters {
		where := i.db.Where(
			"cluster = ? ",
			cluster,
		)
		result := i.db.Unscoped().Delete(&models.Role{}, where)
		if result.Error != nil {
			return fmt.Errorf("failed to delete all objects: %w", result.Error)
		}
	}

	return nil
}

func (i *SQLiteStore) DeleteAllRoleBindings(ctx context.Context, clusters []string) error {
	for _, cluster := range clusters {
		where := i.db.Where(
			"cluster = ? ",
			cluster,
		)
		result := i.db.Unscoped().Delete(&models.RoleBinding{}, where)
		if result.Error != nil {
			return fmt.Errorf("failed to delete all objects: %w", result.Error)
		}
	}
	return nil
}

func (i *SQLiteStore) DeleteAllObjects(ctx context.Context, clusters []string) error {
	for _, cluster := range clusters {
		where := i.db.Where(
			"cluster = ? ",
			cluster,
		)
		result := i.db.Unscoped().Delete(&models.Object{}, where)
		if result.Error != nil {
			return fmt.Errorf("failed to delete all objects: %w", result.Error)
		}
	}

	return nil
}

func NewSQLiteStore(db *gorm.DB, log logr.Logger) (*SQLiteStore, error) {
	return &SQLiteStore{
		db:    db,
		log:   log.WithName("sqllite"),
		debug: log.WithName("sqllite").V(logger.LogLevelDebug),
	}, nil
}

func (i *SQLiteStore) StoreRoles(ctx context.Context, roles []models.Role) error {

	for _, role := range roles {
		if err := role.Validate(); err != nil {
			return fmt.Errorf("invalid role: %w", err)
		}

		role.ID = role.GetID()

		result := i.db.Unscoped().Delete(&role.PolicyRules, "role_id = ?", role.ID)
		if result.Error != nil {
			return fmt.Errorf("failed to delete policy rules: %w", result.Error)
		}

		m := i.db.Model(&role).Association("PolicyRules")

		if err := m.Delete(role.PolicyRules); err != nil {
			return fmt.Errorf("failed to delete policy rules: %w", err)
		}

		clauses := i.db.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "id"},
			},
			UpdateAll: true,
		})

		result = clauses.Create(&role)

		if result.Error != nil {
			return fmt.Errorf("failed to store role: %w", result.Error)
		}

		i.debug.Info("role stored", "role", role.GetID())
	}

	return nil
}

func (i *SQLiteStore) StoreRoleBindings(ctx context.Context, roleBindings []models.RoleBinding) error {
	for _, roleBinding := range roleBindings {
		if err := roleBinding.Validate(); err != nil {
			return fmt.Errorf("invalid role binding: %w", err)
		}

		roleBinding.ID = roleBinding.GetID()

		result := i.db.Unscoped().Delete(&roleBinding.Subjects, "role_binding_id = ?", roleBinding.ID)
		if result.Error != nil {
			return fmt.Errorf("failed to delete subjects: %w", result.Error)
		}

		m := i.db.Model(&roleBinding).Association("Subjects")

		if err := m.Delete(roleBinding.Subjects); err != nil {
			return fmt.Errorf("failed to delete subjects: %w", err)
		}

		clauses := i.db.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "id"},
			},
			UpdateAll: true,
		})

		result = clauses.Create(&roleBinding)
		if result.Error != nil {
			return fmt.Errorf("failed to store role binding: %w", result.Error)
		}
		i.debug.Info("rolebinding stored", "rolebinding", roleBinding.GetID())

	}

	return nil
}

func (i *SQLiteStore) StoreObjects(ctx context.Context, objects []models.Object) error {
	//do nothing if empty collection
	if len(objects) == 0 {
		return nil
	}
	// Gotta copy the objects because we need to set the ID
	rows := []models.Object{}

	for _, object := range objects {
		if err := object.Validate(); err != nil {
			return fmt.Errorf("invalid object: %w", err)
		}

		object.ID = object.GetID()
		rows = append(rows, object)
		i.debug.Info("storing object", "object", object.GetID())
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

	i.debug.Info("objects stored", "rows-affected", result.RowsAffected)
	return nil
}

func (i *SQLiteStore) GetObjects(ctx context.Context, ids []string, opts QueryOption) (Iterator, error) {
	// If offset is zero, it was not set.
	// -1 tells GORM to ignore the offset
	var offset int = -1
	var orderBy string = ""

	if opts != nil {
		if opts.GetOffset() != 0 {
			offset = int(opts.GetOffset())
		}

		if opts.GetOrderBy() != "" {
			orderBy = opts.GetOrderBy()
		}
	}

	tx := i.db.Model(&models.Object{})

	tx = tx.Offset(offset)
	tx = tx.Order(orderBy)

	if ids == nil {
		return sqliterator.New(tx)
	}

	tx = tx.Where("id IN ?", ids)

	if tx.Error != nil {
		return nil, fmt.Errorf("failed to execute query: %w", tx.Error)
	}
	i.debug.Info("objects retrieved", "numResults", tx.RowsAffected)
	return sqliterator.New(tx)
}

func (i *SQLiteStore) GetObjectByID(ctx context.Context, id string) (models.Object, error) {
	object := models.Object{}

	result := i.db.Model(&object).Where("id = ?", id).First(&object)
	if result.Error != nil {
		return models.Object{}, fmt.Errorf("failed to get object: %w", result.Error)
	}

	return object, nil
}

func (i *SQLiteStore) GetAllObjects(ctx context.Context) (Iterator, error) {
	var objects []models.Object
	result := i.db.Model(&models.Object{}).Find(&objects)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get objects: %w", result.Error)
	}

	return sqliterator.New(result)
}

func (i *SQLiteStore) GetRoles(ctx context.Context) ([]models.Role, error) {
	var roles []models.Role
	result := i.db.Model(&models.Role{}).Preload("PolicyRules").Find(&roles)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get roles: %w", result.Error)
	}
	return roles, nil
}

func (i *SQLiteStore) GetRoleBindings(ctx context.Context) ([]models.RoleBinding, error) {
	var rolebindings []models.RoleBinding
	result := i.db.Model(&models.RoleBinding{}).Preload("Subjects").Find(&rolebindings)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get rolebindings: %w", result.Error)
	}
	return rolebindings, nil
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

func (i *SQLiteStore) DeleteRoles(ctx context.Context, roles []models.Role) error {
	for _, role := range roles {
		if err := role.Validate(); err != nil {
			return fmt.Errorf("invalid role: %w", err)
		}

		where := i.db.Where(
			"id = ? ",
			role.GetID(),
		)
		result := i.db.Unscoped().Delete(&models.Role{}, where)
		if result.Error != nil {
			return fmt.Errorf("failed to delete role: %w", result.Error)
		}
	}

	return nil
}

func (i *SQLiteStore) DeleteRoleBindings(ctx context.Context, roleBindings []models.RoleBinding) error {
	for _, roleBinding := range roleBindings {
		if err := roleBinding.Validate(); err != nil {
			return fmt.Errorf("invalid role binding: %w", err)
		}

		where := i.db.Where(
			"id = ? ",
			roleBinding.GetID(),
		)
		result := i.db.Unscoped().Delete(&models.RoleBinding{}, where)
		if result.Error != nil {
			return fmt.Errorf("failed to delete role binding: %w", result.Error)
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
		return nil, fmt.Errorf("failed to create sql database: %w", err)
	}

	// From the readme: https://github.com/mattn/go-sqlite3
	goDB.SetMaxOpenConns(1)

	if err := db.AutoMigrate(&models.Object{}, &models.Role{}, &models.Subject{}, &models.RoleBinding{}, &models.PolicyRule{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}
