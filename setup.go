package logsql

import (
	"embed"
	"fmt"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

//go:embed migrations
var migrationFS embed.FS

// ...
func init() {
	// Register plugin
	plugin.Register("logsql", setup)
}

// ...
func setup(c *caddy.Controller) error {
	// Parse configuration
	db, err := parseConfig(c)
	if err != nil {
		return plugin.Error("logsql", fmt.Errorf("parse config: %w", err))
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		return plugin.Error("logsql", fmt.Errorf("run migrations: %w", err))
	}

	// ...
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return LogSql{Next: next, DB: db}
	})

	return nil
}

// parseConfig parses the configuration for the logsql plugin.
func parseConfig(c *caddy.Controller) (*sqlx.DB, error) {
	// Skip plugin name directive
	c.Next()

	// Read mandatory 'dialext argument
	if !c.NextArg() {
		return nil, fmt.Errorf("missing 'dialect' argument: %w", c.ArgErr())
	}

	dialect := c.Val()

	// Read mandatory 'DSN' argument
	if !c.NextArg() {
		return nil, fmt.Errorf("missing 'DSN' argument: %w", c.ArgErr())
	}

	dsn := c.Val()

	// Make sure no additional arguments are provided
	if c.NextArg() {
		return nil, fmt.Errorf("unexpected argument after DSN: %w", c.ArgErr())
	}

	// Create database connection
	db, err := sqlx.Open(dialect, dsn)
	if err != nil {
		return nil, fmt.Errorf("create database connection: %w", err)
	}

	return db, nil
}

// runMigrations runs database migrations using golang-migrate
func runMigrations(db *sqlx.DB) error {
	// Create source driver
	sd, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("create source driver: %w", err)
	}

	defer sd.Close()

	// Create database database driver
	dd, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("create database driver: %w", err)
	}

	// Create migrate instance
	mi, err := migrate.NewWithInstance("iofs", sd, "postgres", dd)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}

	// Run migrations
	err = mi.Up()
	if (err != nil) && (err != migrate.ErrNoChange) {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}
