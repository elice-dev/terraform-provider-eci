package api

import (
	"errors"
	"net/http"
	"testing"
)

const lbId = "33333333-3333-3333-3333-333333333333"

func testLBRules() []LoadBalancerRule {
	path := "/healthz"
	return []LoadBalancerRule{
		{
			Proto: "TCP",
			Port:  80,
			Servers: []LoadBalancerRuleServer{
				{
					Ip:                 "10.0.0.5",
					NetworkInterfaceId: "44444444-4444-4444-4444-444444444444",
					Port:               8080,
				},
			},
			Healthcheck: LoadBalancerHealthCheck{
				Proto: "HTTP", Port: 8080, Interval: 5, Timeout: 3,
				HealthyThreshold: 2, UnhealthyThreshold: 3, Path: &path,
			},
			Algorithm: "roundrobin",
		},
		{
			Proto:   "UDP",
			Port:    53,
			Servers: []LoadBalancerRuleServer{},
			Healthcheck: LoadBalancerHealthCheck{
				Proto: "TCP", Port: 53, Interval: 10, Timeout: 5,
				HealthyThreshold: 1, UnhealthyThreshold: 1,
			},
			Algorithm: "leastconn",
		},
	}
}

func TestPostLoadBalancer_SendsFullBody(t *testing.T) {
	client, fs := newFakeClient(t, http.StatusOK, `{"id":"`+lbId+`"}`)

	res, err := client.PostLoadBalancer(
		"web", "net-1", "pip-1", testLBRules(), "price-1", map[string]string{"env": "dev"},
	)
	if err != nil {
		t.Fatalf("PostLoadBalancer: %v", err)
	}
	if res.Id.String() != lbId {
		t.Errorf("Id = %s", res.Id)
	}

	req := fs.only()
	assertRequest(t, req, http.MethodPost, "/user/resource/network/load_balancer")
	// The second rule has no healthcheck path, so "path" must be omitted, not sent as null.
	assertJSONEqual(t, "body", req.Body, `{
		"zone_id": "`+testZoneId+`",
		"name": "web",
		"attached_network_id": "net-1",
		"public_ip_id": "pip-1",
		"pricing_id": "price-1",
		"tags": {"env": "dev"},
		"rules": [
			{"proto": "TCP", "port": 80, "algorithm": "roundrobin",
			 "servers": [{"ip": "10.0.0.5", "network_interface_id": "44444444-4444-4444-4444-444444444444", "port": 8080}],
			 "healthcheck": {"proto": "HTTP", "port": 8080, "interval": 5, "timeout": 3,
			                 "healthy_threshold": 2, "unhealthy_threshold": 3, "path": "/healthz"}},
			{"proto": "UDP", "port": 53, "algorithm": "leastconn", "servers": [],
			 "healthcheck": {"proto": "TCP", "port": 53, "interval": 10, "timeout": 5,
			                 "healthy_threshold": 1, "unhealthy_threshold": 1}}
		]
	}`)
}

func TestPatchLoadBalancer_SendsOnlySetFields(t *testing.T) {
	name := "renamed"
	rules := testLBRules()[:1]
	tags := map[string]string{}

	cases := []struct {
		name  string
		nameP *string
		rules *[]LoadBalancerRule
		tags  *map[string]string
		want  string
	}{
		{"nothing", nil, nil, nil, `{}`},
		{"name only", &name, nil, nil, `{"name": "renamed"}`},
		{"empty tags are sent", nil, nil, &tags, `{"tags": {}}`},
		{
			"rules",
			nil,
			&rules,
			nil,
			`{"rules": [{"proto": "TCP", "port": 80, "algorithm": "roundrobin",
			"servers": [{"ip": "10.0.0.5", "network_interface_id": "44444444-4444-4444-4444-444444444444", "port": 8080}],
			"healthcheck": {"proto": "HTTP", "port": 8080, "interval": 5, "timeout": 3,
			                "healthy_threshold": 2, "unhealthy_threshold": 3, "path": "/healthz"}}]}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, fs := newFakeClient(t, http.StatusOK, `{"id":"`+lbId+`"}`)
			if _, err := client.PatchLoadBalancer(lbId, tc.nameP, tc.rules, tc.tags); err != nil {
				t.Fatalf("PatchLoadBalancer: %v", err)
			}
			req := fs.only()
			assertRequest(t, req, http.MethodPatch, "/user/resource/network/load_balancer/"+lbId)
			if req.Body == nil {
				req.Body = map[string]interface{}{}
			}
			assertJSONEqual(t, "body", req.Body, tc.want)
		})
	}
}

func TestGetLoadBalancer_ParsesResponse(t *testing.T) {
	client, fs := newFakeClient(t, http.StatusOK, `{
		"id": "`+lbId+`", "tags": {"a": "b"}, "created": "`+testCreated+`", "modified": null,
		"zone_id": "`+testZoneId+`", "organization_id": "`+testOrgId+`", "user_id": null,
		"name": "web", "short_id": "lb-abc", "attached_network_id": "55555555-5555-5555-5555-555555555555",
		"kubernetes_cluster_id": null,
		"rules": [{"proto": "TCP", "port": 80, "algorithm": "random", "servers": [],
		           "healthcheck": {"proto": "TCP", "port": 80, "interval": 5, "timeout": 3,
		                           "healthy_threshold": 2, "unhealthy_threshold": 3, "path": null}}],
		"pricing_type": "ondemand", "pricing_id": "66666666-6666-6666-6666-666666666666",
		"activated": "`+testCreated+`", "deleting": null, "deleted": null, "status": "activated"
	}`)

	lb, err := client.GetLoadBalancer(lbId)
	if err != nil {
		t.Fatalf("GetLoadBalancer: %v", err)
	}
	assertRequest(t, fs.only(), http.MethodGet, "/user/resource/network/load_balancer/"+lbId)

	if lb.Status != "activated" || lb.ShortId != "lb-abc" || lb.KubernetesClusterId != nil {
		t.Errorf("unexpected scalar fields: %+v", lb)
	}
	if len(lb.Rules) != 1 || lb.Rules[0].Healthcheck.Path != nil ||
		lb.Rules[0].Algorithm != "random" {
		t.Errorf("unexpected rules: %+v", lb.Rules)
	}
	if lb.Activated == nil {
		t.Error("Activated should be set")
	}
}

func TestDeleteLoadBalancer(t *testing.T) {
	client, fs := newFakeClient(t, http.StatusOK, `{"id":"`+lbId+`","status":"deleting"}`)

	res, err := client.DeleteLoadBalancer(lbId)
	if err != nil {
		t.Fatalf("DeleteLoadBalancer: %v", err)
	}
	assertRequest(t, fs.only(), http.MethodDelete, "/user/resource/network/load_balancer/"+lbId)
	if res.Status != "deleting" {
		t.Errorf("Status = %s", res.Status)
	}
}

func TestLoadBalancer_APIErrorIsReturned(t *testing.T) {
	client, _ := newFakeClient(t, http.StatusNotFound,
		`{"code":"RESOURCE_NOT_FOUND","message":"load balancer not found"}`)

	_, err := client.GetLoadBalancer(lbId)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *APIError", err)
	}
	if apiErr.HttpCode != http.StatusNotFound || !apiErr.IsCode("RESOURCE_NOT_FOUND") {
		t.Errorf("unexpected APIError: %v", apiErr)
	}
}
