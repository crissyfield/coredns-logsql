package logsql

import (
	"github.com/miekg/dns"
)

// ...
type ResponseWriter struct {
	dns.ResponseWriter
	Domains []string
}

// WriteMsg writes a reply back to the client.
func (rw *ResponseWriter) WriteMsg(res *dns.Msg) error {
	for _, a := range res.Answer {
		rw.Domains = append(rw.Domains, a.Header().Name)
	}

	return rw.ResponseWriter.WriteMsg(res)
}
