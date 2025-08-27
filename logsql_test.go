package logsql

import (
	"context"
	"testing"

	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/miekg/dns"
)

// TestLogSqlServeDNS_ErrorHandling tests error handling and propagation in ServeDNS method.
func TestLogSqlServeDNS_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		nextHandler func() test.HandlerFunc
		expectError bool
		expectRcode int
	}{
		{
			name: "next handler error propagated",
			nextHandler: func() test.HandlerFunc {
				return test.HandlerFunc(func(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
					return dns.RcodeServerFailure, dns.ErrTime
				})
			},
			expectError: true,
			expectRcode: dns.RcodeServerFailure,
		},
		{
			name: "next handler success with database failure",
			nextHandler: func() test.HandlerFunc {
				return test.HandlerFunc(func(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
					resp := new(dns.Msg)
					resp.SetReply(r)
					resp.Answer = append(resp.Answer, &dns.A{
						Hdr: dns.RR_Header{Name: "test.example.org.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 300},
					})
					w.WriteMsg(resp)
					return dns.RcodeSuccess, nil
				})
			},
			expectError: false,
			expectRcode: dns.RcodeSuccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use closed database to simulate DB failure
			db, _ := sqlx.Open("sqlite3", ":memory:")
			db.Close()

			ls := NewLogSql(tt.nextHandler(), db)
			defer ls.Close()

			req := new(dns.Msg)
			req.SetQuestion("example.org.", dns.TypeA)
			rec := dnstest.NewRecorder(&test.ResponseWriter{})

			rcode, err := ls.ServeDNS(context.TODO(), rec, req)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if rcode != tt.expectRcode {
				t.Errorf("Expected rcode %d, got %d", tt.expectRcode, rcode)
			}
		})
	}
}

// TestInsertIntoDB_EdgeCases tests edge cases and error conditions for database insertions.
func TestInsertIntoDB_EdgeCases(t *testing.T) {
	db, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create test table
	_, err = db.Exec(
		`
			CREATE TABLE "answers" (
				"domain"              TEXT NOT NULL,
				"created_at"          TIMESTAMPTZ NOT NULL,
				"updated_at"          TIMESTAMPTZ NOT NULL,
				PRIMARY KEY ("domain")
			)
		`,
	)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	tests := []struct {
		name        string
		domains     []string
		expectError bool
		setupDB     func(*sqlx.DB)
	}{
		{
			name:    "empty domains slice",
			domains: []string{},
		},
		{
			name:    "nil domains slice",
			domains: nil,
		},
		{
			name:        "database connection closed",
			domains:     []string{"test.example.org."},
			expectError: true,
			setupDB: func(db *sqlx.DB) {
				db.Close()
			},
		},
		{
			name:        "invalid table schema",
			domains:     []string{"test.example.org."},
			expectError: true,
			setupDB: func(db *sqlx.DB) {
				db.Exec(`DROP TABLE "answers"`)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB := db
			if tt.setupDB != nil {
				// Create fresh DB for this test
				testDB, _ = sqlx.Open("sqlite3", ":memory:")
				testDB.Exec(
					`
						CREATE TABLE "answers" (
							"domain"              TEXT NOT NULL,
							"created_at"          TIMESTAMPTZ NOT NULL,
							"updated_at"          TIMESTAMPTZ NOT NULL,
							PRIMARY KEY ("domain")
						)
					`,
				)
				tt.setupDB(testDB)
				defer testDB.Close()
			}

			testLS := &LogSql{DB: testDB}
			err := testLS.insertIntoDB(tt.domains)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestLogSql_DatabaseErrors tests behavior with database connection errors.
func TestLogSql_DatabaseErrors(t *testing.T) {
	// Test with closed database connection
	db, _ := sqlx.Open("sqlite3", ":memory:")
	db.Close()

	ls := &LogSql{DB: db}
	err := ls.insertIntoDB([]string{"test.example.org."})

	if err == nil {
		t.Error("Expected error with closed database connection")
	}
}

