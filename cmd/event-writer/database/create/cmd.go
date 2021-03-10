package create

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
	"github.com/weaveworks/wks/common/database/utils"
)

// Cmd to create the MCCP database
var Cmd = &cobra.Command{
	Use:   "create",
	Short: "Create the MCCP database.",
	RunE: func(_ *cobra.Command, _ []string) error {
		return runCommand(globalParams)
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

type paramSet struct {
	URI      string
	Type     string
	Name     string
	User     string
	Password string
}

var globalParams paramSet

func init() {
	Cmd.Flags().StringVar(&globalParams.URI, "db-uri", os.Getenv("DB_URI"), "URI of the database")
	Cmd.Flags().StringVar(&globalParams.Type, "db-type", os.Getenv("DB_TYPE"), "database type, supported types [sqlite, postgres]")
	Cmd.Flags().StringVar(&globalParams.Name, "db-name", os.Getenv("DB_NAME"), "database name, applicable if type is postgres")
	Cmd.Flags().StringVar(&globalParams.User, "db-user", os.Getenv("DB_USER"), "database user")
	Cmd.Flags().StringVar(&globalParams.Password, "db-password", os.Getenv("DB_PASSWORD"), "database password")
}

func runCommand(globalParams paramSet) error {
	if globalParams.URI == "" {
		return errors.New("--db-uri not provided and $DB_URI not set")
	}
	if globalParams.Type == "" {
		return errors.New("--db-type not provided and $DB_TYPE not set")
	}

	db, err := utils.Open(globalParams.URI, globalParams.Type, globalParams.Name, globalParams.User, globalParams.Password)
	if err != nil {
		return err
	}
	// Set the Ref to the created DB
	err = utils.MigrateTables(db)
	if err != nil {
		return err
	}
	return nil
}
