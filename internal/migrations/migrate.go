package migrations

import (
	"context"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const batchSize = 1000

type Migration struct {
	Number   uint                                                `json:"number"`
	Name     string                                              `json:"name"`
	Forwards func(db *pgxpool.Pool, logger *logrus.Logger) error `json:"-"`
}

var Migrations []*Migration

func Migrate(logger *logrus.Logger, dryRun bool, number int, forceMigrate bool) error {
	if dryRun {
		logger.Infof("=== DRY RUN ===")
	}

	// Check for duplicate migration numbers
	migrationIDs := make(map[uint]struct{})
	for _, migration := range Migrations {
		if _, ok := migrationIDs[migration.Number]; ok {
			err := fmt.Errorf("duplicate migration number found: %d", migration.Number)
			logger.Errorf("Unable to apply migrations, err: %+v", err)
			return err
		}
		migrationIDs[migration.Number] = struct{}{}
	}

	// Sort migrations by number
	sort.Slice(Migrations, func(i, j int) bool {
		return Migrations[i].Number < Migrations[j].Number
	})

	// Connect to database
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		viper.GetString("Database.Host"),
		viper.GetInt("Database.Port"),
		viper.GetString("Database.User"),
		viper.GetString("Database.Password"),
		viper.GetString("Database.Name"),
	)

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}
	defer pool.Close()

	// Force migrate - drop and recreate schema
	if forceMigrate {
		logger.Infof("=== FORCE MIGRATE ===")
		if _, err := pool.Exec(ctx, "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"); err != nil {
			return fmt.Errorf("unable to reset schema: %w", err)
		}
	}

	// Create migrations table if not exists
	logger.Debugf("ensuring migrations table is present")
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS migrations (
			number BIGINT PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`
	if _, err := pool.Exec(ctx, createTableSQL); err != nil {
		return fmt.Errorf("unable to create migrations table: %w", err)
	}

	// Get latest migration
	var latestNumber uint
	err = pool.QueryRow(ctx, "SELECT COALESCE(MAX(number), 0) FROM migrations").Scan(&latestNumber)
	if err != nil {
		return fmt.Errorf("unable to find latest migration: %w", err)
	}

	if len(Migrations) == 0 {
		logger.Infof("no migrations to apply")
		return nil
	}

	if latestNumber >= Migrations[len(Migrations)-1].Number {
		logger.Infof("no migrations to apply - database is up to date")
		return nil
	}

	if number == -1 {
		number = int(Migrations[len(Migrations)-1].Number)
	}

	if uint(number) <= latestNumber && latestNumber > 0 {
		logger.Infof("no migrations to apply, specified number is less than or equal to latest migration")
		return nil
	}

	// Apply migrations
	for _, migration := range Migrations {
		if migration.Number > uint(number) {
			break
		}

		if migration.Number <= latestNumber {
			continue
		}

		migLogger := logger.WithFields(logrus.Fields{
			"migration_number": migration.Number,
			"migration_name":   migration.Name,
		})
		migLogger.Infof("applying migration %d: %q", migration.Number, migration.Name)

		if dryRun {
			continue
		}

		// Begin transaction
		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("unable to begin transaction: %w", err)
		}

		// Apply migration
		if err := migration.Forwards(pool, logger); err != nil {
			tx.Rollback(ctx)
			migLogger.Errorf("unable to apply migration, rolling back. err: %+v", err)
			return err
		}

		// Record migration
		_, err = tx.Exec(ctx, "INSERT INTO migrations (number, name) VALUES ($1, $2)", migration.Number, migration.Name)
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("unable to record migration: %w", err)
		}

		// Commit transaction
		if err := tx.Commit(ctx); err != nil {
			migLogger.Errorf("unable to commit transaction... err: %+v", err)
			return err
		}

		migLogger.Infof("migration %d applied successfully", migration.Number)
	}

	logger.Infof("all migrations applied successfully")
	return nil
}
