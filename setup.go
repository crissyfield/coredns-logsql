package logsql

import (
	"log/slog"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// ...
func init() {
	slog.Info("[logsql] init")
	plugin.Register("logsql", setup)
}

// ...
func setup(c *caddy.Controller) error {
	// ...
	c.Next()

	db, err := parseConfig(c)
	if err != nil {
		return plugin.Error("logsql", err)
	}

	slog.Info("[logsql] setup", slog.Any("db", db))

	// ...
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		slog.Info("[logsql] add plugin")
		return LogSql{Next: next}
	})

	return nil
}

// parseConfig parses the configuration for the logsql plugin.
func parseConfig(c *caddy.Controller) (*sqlx.DB, error) {
	// ...
	if !c.NextArg() {
		return nil, c.ArgErr()
	}

	dialect := c.Val()

	// ...
	if !c.NextArg() {
		return nil, c.ArgErr()
	}

	arg := c.Val()

	// ...
	if c.NextArg() {
		return nil, c.ArgErr()
	}

	// ...
	db, err := sqlx.Open(dialect, arg)
	if err != nil {
		return nil, err
	}

	return db, nil
}
