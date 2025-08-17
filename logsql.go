package logsql

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/coredns/coredns/plugin"
	"github.com/jmoiron/sqlx"
	"github.com/miekg/dns"
)

// LogSql ...
type LogSql struct {
	Next plugin.Handler
	DB   *sqlx.DB
}

// Request ...
type Request struct {
	Domain    string    `db:"domain"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Name implements the plugin Handler interface.
func (ls LogSql) Name() string {
	return "logsql"
}

// ServeDNS implements the plugin.Handler interface.
func (ls LogSql) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	// Insert into database
	err := ls.insertIntoDB(ctx, r.Question)
	if err != nil {
		slog.Error("logsql: failed to insert request into database: ", slog.Any("question", r.Question), slog.Any("error", err))
	}

	// Call next plugin
	return plugin.NextOrFailure(ls.Name(), ls.Next, ctx, w, r)
}

// ...
func (ls LogSql) insertIntoDB(ctx context.Context, question []dns.Question) error {
	// Early return if there are no questions in the request
	if len(question) == 0 {
		return nil
	}

	// Insert record into database
	requests := make([]Request, 0, len(question))
	for _, q := range question {
		requests = append(requests, Request{
			Domain:    q.Name,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}

	query, args, err := sqlx.Named(
		`
			INSERT INTO "requests" (
				"domain",
				"created_at",
				"updated_at"
			)
			VALUES (
				:domain,
				:created_at,
				:updated_at
			)
			ON CONFLICT ("domain") DO UPDATE
			SET
				"updated_at" = EXCLUDED."updated_at"
		`,
		requests,
	)

	if err != nil {
		return fmt.Errorf("failed to prepare named query: %w", err)
	}

	_, err = ls.DB.ExecContext(
		ctx,
		ls.DB.Rebind(query),
		args...,
	)

	if err != nil {
		return fmt.Errorf("failed to execute named query: %w", err)
	}

	return nil
}
