package logsql

import (
	"context"
	"log/slog"

	"github.com/coredns/coredns/plugin"

	"github.com/miekg/dns"
)

// LogSql ...
type LogSql struct {
	Next plugin.Handler
}

// ServeDNS implements the plugin.Handler interface.
func (ls LogSql) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	a, err := plugin.NextOrFailure(ls.Name(), ls.Next, ctx, w, r)

	slog.Info("[logsql] ServeDNS called", slog.Any("msg", r), slog.Any("writer", w))

	return a, err
}

// Name implements the plugin Handler interface.
func (ls LogSql) Name() string {
	return "logsql"
}
