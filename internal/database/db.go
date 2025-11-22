package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
)

func ConnectDB(ctx context.Context) (DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		viper.GetString("Database.Host"),
		viper.GetInt("Database.Port"),
		viper.GetString("Database.User"),
		viper.GetString("Database.Password"),
		viper.GetString("Database.Name"),
	)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	log.Println("Connected to database successfully")
	return pool, nil
}
