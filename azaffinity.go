package azaffinity

import (
	"context"
	"fmt"
	"net"
	"regexp"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

type AZAffinity struct {
	Next        plugin.Handler
	SubnetAZMap map[string]string
}

var internalAwsAlbDnsPattern = regexp.MustCompile(`^internal-k8s-.*\.ap-southeast-1.elb\.amazonaws\.com$`)

func init() {
	caddy.RegisterPlugin("azaffinity", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func (a AZAffinity) Name() string { return "azaffinity" }

func setup(c *caddy.Controller) error {
	azaffinity := AZAffinity{SubnetAZMap: make(map[string]string)}

	for c.Next() {
		for c.NextBlock() {
			subnet := c.Val()
			if _, _, err := net.ParseCIDR(subnet); err != nil {
				return plugin.Error("azaffinity", fmt.Errorf("invalid subnet: %s", subnet))
			}

			if !c.NextArg() {
				return c.ArgErr()
			}

			az := c.Val()
			if az == "" {
				return plugin.Error("azaffinity", fmt.Errorf("availability zone cannot be empty"))
			}
			azaffinity.SubnetAZMap[subnet] = az
		}
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		azaffinity.Next = next
		return azaffinity
	})

	return nil
}

func (a AZAffinity) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	clientIP, _, _ := net.SplitHostPort(state.IP())
	az, err := a.getAvailabilityZone(clientIP)
	if err != nil {
		return plugin.NextOrFailure(a.Name(), a.Next, ctx, w, r)
	}

	for _, question := range r.Question {
		if !internalAwsAlbDnsPattern.MatchString(question.Name) {
			continue
		}

		// Append the availability zone to the record
		question.Name = az + "." + question.Name
	}

	return plugin.NextOrFailure(a.Name(), a.Next, ctx, w, r)
}

func (a AZAffinity) getAvailabilityZone(ip string) (string, error) {
	for cidr, az := range a.SubnetAZMap {
		if _, subnet, err := net.ParseCIDR(cidr); err == nil {
			if subnet.Contains(net.ParseIP(ip)) {
				return az, nil
			}
		}
	}
	return "", fmt.Errorf("no availability zone found for IP %s", ip)
}
