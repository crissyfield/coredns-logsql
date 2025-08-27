package logsql

import (
	"testing"

	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
)

// TestResponseWriter_DomainDeduplication tests that duplicate domains in DNS responses are deduplicated.
func TestResponseWriter_DomainDeduplication(t *testing.T) {
	baseWriter := &test.ResponseWriter{}
	rw := NewResponseWriter(baseWriter)

	resp := &dns.Msg{
		Answer: []dns.RR{
			&dns.A{
				Hdr: dns.RR_Header{Name: "example.org.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 300},
			},
			&dns.A{
				Hdr: dns.RR_Header{Name: "example.org.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 600},
			},
		},
	}

	rw.WriteMsg(resp)
	domains := rw.Domains()

	if len(domains) != 1 {
		t.Errorf("Expected 1 unique domain, got %d: %v", len(domains), domains)
	}
}

type mockResponseWriter struct {
	*test.ResponseWriter
	errorToReturn error
}

func (m *mockResponseWriter) WriteMsg(msg *dns.Msg) error {
	return m.errorToReturn
}

// TestResponseWriter_WriteError tests that domains are captured even when WriteMsg returns an error.
func TestResponseWriter_WriteError(t *testing.T) {
	baseWriter := &mockResponseWriter{
		ResponseWriter: &test.ResponseWriter{},
		errorToReturn:  dns.ErrTime,
	}
	rw := NewResponseWriter(baseWriter)

	resp := &dns.Msg{
		Answer: []dns.RR{
			&dns.A{Hdr: dns.RR_Header{Name: "example.org.", Rrtype: dns.TypeA}},
		},
	}

	err := rw.WriteMsg(resp)
	if err != dns.ErrTime {
		t.Errorf("Expected dns.ErrTime, got %v", err)
	}

	domains := rw.Domains()
	if len(domains) != 1 {
		t.Errorf("Expected domains captured despite error, got: %v", domains)
	}
}

