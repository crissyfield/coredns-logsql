package logsql

import (
	"github.com/miekg/dns"
)

// ...
type ResponseWriter struct {
	dns.ResponseWriter
	domains map[string]bool
}

// ...
func NewResponseWriter(w dns.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		domains:        make(map[string]bool),
	}
}

// WriteMsg writes a reply back to the client.
func (rw *ResponseWriter) WriteMsg(res *dns.Msg) error {
	for _, a := range res.Answer {
		rw.domains[a.Header().Name] = true
	}

	return rw.ResponseWriter.WriteMsg(res)
}

// ...
func (rw *ResponseWriter) Domains() []string {
	var domains []string
	for domain := range rw.domains {
		domains = append(domains, domain)
	}
	return domains
}
