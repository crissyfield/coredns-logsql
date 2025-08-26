package logsql

import (
	"embed"
	"fmt"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations
var migrationFS embed.FS

func init() {
	// Register plugin
	plugin.Register("logsql", setup)
}

// setup sets up the logsql plugin.
func setup(c *caddy.Controller) error {
	// Create database connection from config
	db, err := createDBFromConfig(c)
	if err != nil {
		return plugin.Error("logsql", fmt.Errorf("create database from config: %w", err))
	}

	// Add plugin to the chain
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		// Create LogSql instance
		ls := NewLogSql(next, db)

		// Register shutdown function to close plugin
		c.OnShutdown(func() error {
			ls.Close()
			return nil
		})

		return ls
	})

	return nil
}

// createDBFromConfig creates a database connection from the plugin configuration.
func createDBFromConfig(c *caddy.Controller) (*sqlx.DB, error) {
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

	// Create database connection based on dialect
	var db *sqlx.DB

	switch dialect {
	case "postgres":
		// Postgres
		d, err := sqlx.Open("pgx", dsn)
		if err != nil {
			return nil, fmt.Errorf("create postgres connection: %w", err)
		}

		db = d

	case "sqlite3":
		// SQLite3
		d, err := sqlx.Open("sqlite3", dsn)
		if err != nil {
			return nil, fmt.Errorf("create sqlite3 connection: %w", err)
		}

		db = d

	default:
		// Unsupported dialect
		return nil, fmt.Errorf("unsupported dialect: %s", dialect)
	}

	// Run migrations
	if err := runMigrations(db, dialect); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return db, nil
}

// runMigrations runs database migrations using golang-migrate
func runMigrations(db *sqlx.DB, dialect string) error {
	// Create source driver
	sd, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("create source driver: %w", err)
	}

	defer sd.Close()

	// Create database driver based on dialect
	var dd database.Driver
	var databaseName string

	switch dialect {
	case "postgres":
		// Postgres
		dd, err = postgres.WithInstance(db.DB, &postgres.Config{})
		if err != nil {
			return fmt.Errorf("create postgres driver: %w", err)
		}

		databaseName = "postgres"

	case "sqlite3":
		// SQLite3
		dd, err = sqlite3.WithInstance(db.DB, &sqlite3.Config{})
		if err != nil {
			return fmt.Errorf("create sqlite3 driver: %w", err)
		}

		databaseName = "sqlite3"

	default:
		// Unsupported dialect
		return fmt.Errorf("unsupported dialect: %s", dialect)
	}

	// Create migrate instance
	mi, err := migrate.NewWithInstance("iofs", sd, databaseName, dd)
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
