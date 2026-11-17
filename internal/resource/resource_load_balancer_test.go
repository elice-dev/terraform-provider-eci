package resource_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type lbOverrides struct {
	rules         func(ruleType tftypes.Object) []tftypes.Value
	ruleProto     string
	rulePort      int64
	algorithm     string
	serverPort    int64
	healthPath    *string
	healthProto   string
	healthTimeout int64
}

func buildLoadBalancer(o lbOverrides) func(tftypes.Object) tftypes.Value {
	return func(typ tftypes.Object) tftypes.Value {
		rulesType := typ.AttributeTypes["rules"].(tftypes.List)
		ruleType := rulesType.ElementType.(tftypes.Object)
		serversType := ruleType.AttributeTypes["servers"].(tftypes.List)
		serverType := serversType.ElementType.(tftypes.Object)
		healthType := ruleType.AttributeTypes["healthcheck"].(tftypes.Object)

		pick := func(v, def string) string {
			if v == "" {
				return def
			}
			return v
		}
		pickN := func(v, def int64) int64 {
			if v == 0 {
				return def
			}
			return v
		}

		health := map[string]tftypes.Value{
			"proto":               str(pick(o.healthProto, "HTTP")),
			"port":                num(8080),
			"interval":            num(5),
			"timeout":             num(pickN(o.healthTimeout, 3)),
			"healthy_threshold":   num(2),
			"unhealthy_threshold": num(3),
		}
		if o.healthPath != nil {
			health["path"] = str(*o.healthPath)
		}

		rule := object(ruleType, map[string]tftypes.Value{
			"proto": str(pick(o.ruleProto, "TCP")),
			"port":  num(pickN(o.rulePort, 80)),
			"servers": tftypes.NewValue(serversType, []tftypes.Value{
				object(serverType, map[string]tftypes.Value{
					"ip":                   str("10.0.0.5"),
					"network_interface_id": str("nic-1"),
					"port":                 num(pickN(o.serverPort, 8080)),
				}),
			}),
			"healthcheck": object(healthType, health),
			"algorithm":   str(pick(o.algorithm, "roundrobin")),
		})

		rules := []tftypes.Value{rule}
		if o.rules != nil {
			rules = o.rules(ruleType)
		}

		return object(typ, map[string]tftypes.Value{
			"tags":                strMap(map[string]string{"env": "dev"}),
			"name":                str("web"),
			"attached_network_id": str("net-1"),
			"public_ip_id":        str("pip-1"),
			"pricing_id":          str("price-1"),
			"rules":               tftypes.NewValue(rulesType, rules),
		})
	}
}

func TestLoadBalancerConfig_Valid(t *testing.T) {
	path := "/healthz"
	assertValid(t, validateConfig(t, "eci_load_balancer", buildLoadBalancer(lbOverrides{})))
	assertValid(
		t,
		validateConfig(t, "eci_load_balancer", buildLoadBalancer(lbOverrides{healthPath: &path})),
	)
	assertValid(t, validateConfig(t, "eci_load_balancer", buildLoadBalancer(lbOverrides{
		ruleProto: "UDP", algorithm: "leastconn", healthProto: "TCP", rulePort: 65535,
	})))
}

func TestLoadBalancerConfig_Invalid(t *testing.T) {
	badPath := "healthz"
	cases := []struct {
		name string
		o    lbOverrides
		want string
	}{
		// The API upper-cases proto, so lower case would never match state after apply.
		{"lower case proto", lbOverrides{ruleProto: "tcp"}, "TCP"},
		{"unknown algorithm", lbOverrides{algorithm: "weighted"}, "roundrobin"},
		{"rule port negative", lbOverrides{rulePort: -1}, "between 1 and 65535"},
		{"server port too large", lbOverrides{serverPort: 70000}, "between 1 and 65535"},
		{"healthcheck proto UDP", lbOverrides{healthProto: "UDP"}, "HTTP"},
		{"healthcheck timeout negative", lbOverrides{healthTimeout: -1}, "at least 1"},
		{"healthcheck path without slash", lbOverrides{healthPath: &badPath}, "must start with /"},
		{
			"no rules",
			lbOverrides{rules: func(tftypes.Object) []tftypes.Value { return []tftypes.Value{} }},
			"at least 1",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertInvalid(
				t,
				validateConfig(t, "eci_load_balancer", buildLoadBalancer(tc.o)),
				tc.want,
			)
		})
	}
}

// A load balancer only accepts a public ip that is not attached to a network interface.
func TestPublicIpConfig_WithoutNetworkInterface(t *testing.T) {
	assertValid(t, validateConfig(t, "eci_public_ip", func(typ tftypes.Object) tftypes.Value {
		return object(typ, map[string]tftypes.Value{
			"tags":       strMap(map[string]string{}),
			"dr":         tftypes.NewValue(tftypes.Bool, false),
			"pricing_id": str("price-1"),
		})
	}))
}
