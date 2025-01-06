package azaffinity

import (
	"context"
	"net"
	"testing"

	"github.com/coredns/caddy"
	"github.com/miekg/dns"
)

func TestSetup(t *testing.T) {
	c := caddy.NewTestController("dns", `azaffinity 10.0.0.0/24 ap-southeast-1a`)
	err := setup(c)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	c = caddy.NewTestController("dns", `azaffinity invalid-subnet ap-southeast-1a`)
	err = setup(c)
	if err == nil {
		t.Fatalf("Expected error for invalid subnet, got none")
	}

	c = caddy.NewTestController("dns", `azaffinity 10.0.0.0/24`)
	err = setup(c)
	if err == nil {
		t.Fatalf("Expected error for missing availability zone, got none")
	}
}

func TestServeDNS(t *testing.T) {
	azaffinity := AZAffinity{
		SubnetAZMap: map[string]string{
			"10.0.0.0/24": "ap-southeast-1a",
		},
	}

	tests := []struct {
		clientIP   string
		question   dns.Question
		expectName string
	}{
		{
			clientIP: "10.0.0.1",
			question: dns.Question{
				Name:  "internal-k8s-example-123.ap-southeast-1.elb.amazonaws.com.",
				Qtype: dns.TypeCNAME,
			},
			expectName: "ap-southeast-1a.internal-k8s-example-123.ap-southeast-1.elb.amazonaws.com.",
		},
		{
			clientIP: "192.168.1.1",
			question: dns.Question{
				Name:  "internal-k8s-example-123.ap-southeast-1.elb.amazonaws.com.",
				Qtype: dns.TypeCNAME,
			},
			expectName: "internal-k8s-example-123.ap-southeast-1.elb.amazonaws.com.",
		},
	}

	for _, test := range tests {
		w := &testResponseWriter{
			remoteAddr: test.clientIP + ":12345",
		}
		r := new(dns.Msg)
		r.SetQuestion(test.question.Name, test.question.Qtype)

		ctx := context.TODO()

		_, err := azaffinity.ServeDNS(ctx, w, r)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if r.Question[0].Name != test.expectName {
			t.Errorf("Expected question name %s, got %s", test.expectName, r.Question[0].Name)
		}
	}
}

type testResponseWriter struct {
	dns.ResponseWriter
	remoteAddr string
}

func (t *testResponseWriter) WriteMsg(m *dns.Msg) error {
	return nil
}

func (t *testResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (t *testResponseWriter) Close() error {
	return nil
}

func (t *testResponseWriter) TsigStatus() error {
	return nil
}

func (t *testResponseWriter) Hijack() {}

func (t *testResponseWriter) RemoteAddr() net.Addr {
    return &net.IPAddr{IP: net.ParseIP(t.remoteAddr)}
}
