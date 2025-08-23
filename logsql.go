package logsql

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/coredns/coredns/plugin"
	"github.com/jmoiron/sqlx"
	"github.com/miekg/dns"
)

const (
	// DomainChannelBufferSize is the size of the channel buffer for domains.
	DomainChannelBufferSize = 1024
)

// LogSql is a plugin that logs DNS queries to a SQL database.
type LogSql struct {
	Next      plugin.Handler // Next plugin in the chain
	DB        *sqlx.DB       // Database connection
	domainsCh chan []string  // Channel for sending domains to background goroutine
	doneCh    chan struct{}  // Channel to signal background goroutine to stop
}

// Request represents a DNS request record in the database.
type Request struct {
	Domain    string    `db:"domain"`     // Domain name
	CreatedAt time.Time `db:"created_at"` // Timestamp of creation
	UpdatedAt time.Time `db:"updated_at"` // Timestamp of last update
}

// NewLogSql creates a new LogSql instance.
func NewLogSql(next plugin.Handler, db *sqlx.DB) *LogSql {
	// Initialize LogSql
	ls := &LogSql{
		Next:      next,
		DB:        db,
		domainsCh: make(chan []string, DomainChannelBufferSize),
		doneCh:    make(chan struct{}),
	}

	// Spin up background goroutine to handle database inserts
	go func() {
		defer close(ls.doneCh)

		for domains := range ls.domainsCh {
			err := ls.insertIntoDB(domains)
			if err != nil {
				slog.Error("[logsql] Failed to insert domains into database", slog.Any("error", err))
			}
		}
	}()

	return ls
}

// Name implements the plugin Handler interface.
func (ls LogSql) Name() string {
	return "logsql"
}

// Close gracefully shuts down the plugin.
func (ls *LogSql) Close() {
	// Signal background goroutine to stop and wait for it to finish
	close(ls.domainsCh)
	<-ls.doneCh
}

// ServeDNS implements the plugin.Handler interface.
func (ls LogSql) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	// Call next plugin
	rw := NewResponseWriter(w)

	val, err := plugin.NextOrFailure(ls.Name(), ls.Next, ctx, rw, r)

	// Send domains to background goroutine for database insertion
	ls.domainsCh <- rw.Domains()

	return val, err
}

// insertIntoDB inserts the given domains into the database.
func (ls LogSql) insertIntoDB(domains []string) error {
	// Early return if there are no domains
	if len(domains) == 0 {
		return nil
	}

	// Prepare requests
	requests := make([]Request, 0, len(domains))
	for _, d := range domains {
		requests = append(requests, Request{
			Domain:    d,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}

	// Prevent deadlocks by forcing a stable order
	slices.SortStableFunc(requests, func(l Request, r Request) int {
		return strings.Compare(l.Domain, r.Domain)
	})

	// Insert record into database
	query, args, err := sqlx.Named(
		`
			INSERT INTO "answers" (
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

	_, err = ls.DB.Exec(ls.DB.Rebind(query), args...)
	if err != nil {
		return fmt.Errorf("failed to execute named query: %w", err)
	}

	return nil
}
