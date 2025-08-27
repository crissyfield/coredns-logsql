package logsql

import (
	"github.com/miekg/dns"
)

// ResponseWriter wraps dns.ResponseWriter to capture domains from DNS responses.
type ResponseWriter struct {
	dns.ResponseWriter                 // Embedded dns.ResponseWriter
	domains            map[string]bool // Set to store unique domains
}

// NewResponseWriter creates a new ResponseWriter instance.
func NewResponseWriter(w dns.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		domains:        make(map[string]bool),
	}
}

// WriteMsg writes a reply back to the client.
func (rw *ResponseWriter) WriteMsg(res *dns.Msg) error {
	// Capture domains from the response answers
	for _, a := range res.Answer {
		rw.domains[a.Header().Name] = true
	}

	// Call the original WriteMsg method
	return rw.ResponseWriter.WriteMsg(res)
}

// Domains returns the list of unique domains captured.
func (rw *ResponseWriter) Domains() []string {
	// Convert map keys to slice
	var domains []string

	for domain := range rw.domains {
		domains = append(domains, domain)
	}

	return domains
}
