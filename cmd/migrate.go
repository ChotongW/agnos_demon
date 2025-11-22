package cmd

import (
	"agnos_demo/internal/helpers"
	"agnos_demo/internal/migrations"

	"github.com/spf13/cobra"
)

var migrateDBCmd = &cobra.Command{
	Use:   "migrate-db",
	Short: "Run database migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		number, _ := cmd.Flags().GetInt("number")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		forceMigrate, _ := cmd.Flags().GetBool("force-migrate")

		logger, err := helpers.CreateLogger("info", nil, nil, "")
		if err != nil {
			return err
		}

		return migrations.Migrate(logger, dryRun, number, forceMigrate)
	},
}

func init() {
	rootCmd.AddCommand(migrateDBCmd)

	migrateDBCmd.Flags().Int("number", -1, "the migration to run forwards until; if not set, will run all migrations")
	migrateDBCmd.Flags().Bool("dry-run", false, "print out migrations to be applied without running them")
	migrateDBCmd.Flags().Bool("force-migrate", false, "drop all the tables before migrate the database")
}
