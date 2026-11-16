package resource

import (
	"context"
	"reflect"
	"terraform-provider-eci/internal/api"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func strPtr(s string) *string { return &s }

func sampleAPIRules() []api.LoadBalancerRule {
	return []api.LoadBalancerRule{
		{
			Proto: "TCP",
			Port:  443,
			Servers: []api.LoadBalancerRuleServer{
				{Ip: "10.0.0.5", NetworkInterfaceId: "nic-1", Port: 8443},
				{Ip: "10.0.0.6", NetworkInterfaceId: "nic-2", Port: 8443},
			},
			Healthcheck: api.LoadBalancerHealthCheck{
				Proto: "HTTP", Port: 8443, Interval: 5, Timeout: 3,
				HealthyThreshold: 2, UnhealthyThreshold: 3, Path: strPtr("/healthz"),
			},
			Algorithm: "leastconn",
		},
		{
			Proto:   "UDP",
			Port:    53,
			Servers: []api.LoadBalancerRuleServer{},
			Healthcheck: api.LoadBalancerHealthCheck{
				Proto: "TCP", Port: 53, Interval: 10, Timeout: 5,
				HealthyThreshold: 1, UnhealthyThreshold: 1,
			},
			Algorithm: "random",
		},
	}
}

func TestLoadBalancerRules_RoundTrip(t *testing.T) {
	want := sampleAPIRules()
	models := loadBalancerRulesFromAPI(want)

	if !models[1].Healthcheck.Path.IsNull() {
		t.Errorf("missing path should become null, got %v", models[1].Healthcheck.Path)
	}
	if models[0].Healthcheck.Path.ValueString() != "/healthz" {
		t.Errorf("path = %v", models[0].Healthcheck.Path)
	}

	got := loadBalancerRulesToAPI(models)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("round trip changed rules:\n got  %+v\n want %+v", got, want)
	}
}

func TestLoadBalancerRules_EmptyServersStayEmptyList(t *testing.T) {
	// The API takes "servers": [] but would reject null, so an empty list must not become nil.
	got := loadBalancerRulesToAPI(loadBalancerRulesFromAPI(sampleAPIRules()))
	if got[1].Servers == nil {
		t.Fatal("servers became nil")
	}
}

func modelWithRules(
	t *testing.T,
	name string,
	tags map[string]string,
	rules []api.LoadBalancerRule,
) *ResourceLoadBalancerModel {
	t.Helper()
	tagValue, diags := types.MapValueFrom(context.Background(), types.StringType, tags)
	if diags.HasError() {
		t.Fatalf("tags: %v", diags)
	}
	return &ResourceLoadBalancerModel{
		Name:  types.StringValue(name),
		Tags:  tagValue,
		Rules: loadBalancerRulesFromAPI(rules),
	}
}

func TestLoadBalancerPatch_OnlyChangedFields(t *testing.T) {
	ctx := context.Background()
	base := map[string]string{"env": "dev"}

	pathChanged := sampleAPIRules()
	pathChanged[1].Healthcheck.Path = strPtr("/ready")

	serverRemoved := sampleAPIRules()
	serverRemoved[0].Servers = serverRemoved[0].Servers[:1]

	cases := []struct {
		name                         string
		plan                         *ResourceLoadBalancerModel
		wantName, wantRules, wantTag bool
	}{
		{"no change", modelWithRules(t, "web", base, sampleAPIRules()), false, false, false},
		{"name", modelWithRules(t, "api", base, sampleAPIRules()), true, false, false},
		{
			"tags",
			modelWithRules(t, "web", map[string]string{}, sampleAPIRules()),
			false,
			false,
			true,
		},
		{
			"healthcheck path null to value",
			modelWithRules(t, "web", base, pathChanged),
			false,
			true,
			false,
		},
		{"server removed", modelWithRules(t, "web", base, serverRemoved), false, true, false},
		{"rule removed", modelWithRules(t, "web", base, sampleAPIRules()[:1]), false, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state := modelWithRules(t, "web", base, sampleAPIRules())
			namePtr, rulesPtr, tagsPtr, diags := loadBalancerPatch(ctx, tc.plan, state)
			if diags.HasError() {
				t.Fatalf("diags: %v", diags)
			}
			if (namePtr != nil) != tc.wantName {
				t.Errorf("name sent = %v, want %v", namePtr != nil, tc.wantName)
			}
			if (rulesPtr != nil) != tc.wantRules {
				t.Errorf("rules sent = %v, want %v", rulesPtr != nil, tc.wantRules)
			}
			if (tagsPtr != nil) != tc.wantTag {
				t.Errorf("tags sent = %v, want %v", tagsPtr != nil, tc.wantTag)
			}
			if rulesPtr != nil &&
				!reflect.DeepEqual(*rulesPtr, loadBalancerRulesToAPI(tc.plan.Rules)) {
				t.Errorf("sent rules differ from plan: %+v", *rulesPtr)
			}
		})
	}
}

func TestLoadBalancerGetResponseToModel_KeepsPublicIpId(t *testing.T) {
	clusterId := uuid.New()
	activated := time.Date(2026, 11, 16, 1, 2, 3, 0, time.UTC)
	response := &api.ResourceLoadBalancerGetResponse{
		Id:                  uuid.New(),
		Tags:                map[string]string{"a": "b"},
		Created:             activated,
		ZoneId:              uuid.New(),
		OrganizationId:      uuid.New(),
		Name:                "web",
		ShortId:             "lb-abc",
		AttachedNetworkId:   uuid.New(),
		KubernetesClusterId: &clusterId,
		Rules:               sampleAPIRules(),
		PricingType:         "ondemand",
		PricingId:           uuid.New(),
		Activated:           &activated,
		Status:              "activated",
	}

	data := ResourceLoadBalancerModel{PublicIpId: types.StringValue("pip-1")}
	if diags := resourceLoadBalancerGetResponseToModel(
		context.Background(),
		response,
		&data,
	); diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}

	if data.PublicIpId.ValueString() != "pip-1" {
		t.Errorf("public_ip_id must be kept from plan/state, got %v", data.PublicIpId)
	}
	if data.KubernetesClusterId.ValueString() != clusterId.String() {
		t.Errorf("kubernetes_cluster_id = %v", data.KubernetesClusterId)
	}
	if !data.Modified.IsNull() || !data.Deleted.IsNull() {
		t.Error("modified and deleted should be null")
	}
	if data.Status.ValueString() != "activated" || data.ShortId.ValueString() != "lb-abc" {
		t.Errorf("unexpected model: %+v", data)
	}
}

func TestLoadBalancerSchema_ReplaceOnlyImmutableFields(t *testing.T) {
	replaced, rejected := changeEffects(t, validSchema(t, NewResourceLoadBalancer()))

	want := map[string]bool{"attached_network_id": true, "public_ip_id": true, "pricing_id": true}
	if !reflect.DeepEqual(replaced, want) {
		t.Errorf("RequiresReplace attributes = %v, want %v", replaced, want)
	}
	if len(rejected) != 0 {
		t.Errorf("no attribute should reject changes, got %v", rejected)
	}
}
