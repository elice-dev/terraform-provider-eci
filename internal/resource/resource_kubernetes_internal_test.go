package resource

import (
	"context"
	"reflect"
	"strings"
	"terraform-provider-eci/internal/api"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestErrorOnChange(t *testing.T) {
	ctx := context.Background()
	existing := tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{})
	absent := tftypes.NewValue(tftypes.Object{}, nil)

	cases := []struct {
		name         string
		state, plan  tftypes.Value
		before, next types.String
		wantError    bool
	}{
		{"create", absent, existing, types.StringNull(), types.StringValue("s3cret"), false},
		{"destroy", existing, absent, types.StringValue("s3cret"), types.StringNull(), false},
		{
			"unchanged",
			existing,
			existing,
			types.StringValue("s3cret"),
			types.StringValue("s3cret"),
			false,
		},
		{
			"unknown plan",
			existing,
			existing,
			types.StringValue("s3cret"),
			types.StringUnknown(),
			false,
		},
		{
			"null in state (import)",
			existing,
			existing,
			types.StringNull(),
			types.StringValue("n3w"),
			false,
		},
		{
			"changed",
			existing,
			existing,
			types.StringValue("s3cret"),
			types.StringValue("n3w"),
			true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := planmodifier.StringRequest{
				Path:        path.Root("password"),
				State:       tfsdk.State{Raw: tc.state},
				Plan:        tfsdk.Plan{Raw: tc.plan},
				StateValue:  tc.before,
				PlanValue:   tc.next,
				ConfigValue: tc.next,
			}
			resp := &planmodifier.StringResponse{PlanValue: tc.next}
			errorOnChange("cannot change").PlanModifyString(ctx, req, resp)

			if resp.Diagnostics.HasError() != tc.wantError {
				t.Fatalf(
					"error = %v, want %v (%v)",
					resp.Diagnostics.HasError(),
					tc.wantError,
					resp.Diagnostics,
				)
			}
			if resp.RequiresReplace {
				t.Error("errorOnChange must never replace the resource")
			}
			for _, d := range resp.Diagnostics {
				if strings.Contains(d.Summary()+d.Detail(), "s3cret") ||
					strings.Contains(d.Summary()+d.Detail(), "n3w") {
					t.Errorf("diagnostic leaks the value: %s %s", d.Summary(), d.Detail())
				}
			}
		})
	}
}

func TestKubernetesClusterSchema_ChangeEffects(t *testing.T) {
	s := validSchema(t, NewResourceKubernetesCluster())
	replaced, rejected := changeEffects(t, s)

	wantReplaced := map[string]bool{
		"name": true, "attached_subnet_id": true, "username": true, "control_plane_spec": true,
	}
	if !reflect.DeepEqual(replaced, wantReplaced) {
		t.Errorf("RequiresReplace attributes = %v, want %v", replaced, wantReplaced)
	}
	wantRejected := map[string]bool{"target_version": true, "password": true}
	if !reflect.DeepEqual(rejected, wantRejected) {
		t.Errorf("rejected attributes = %v, want %v", rejected, wantRejected)
	}

	if !s.Attributes["password"].IsSensitive() {
		t.Error("password must be sensitive")
	}
	pricing := s.Attributes["control_plane_spec"].(schema.SingleNestedAttribute).
		Attributes["pricing_ids"].(schema.SingleNestedAttribute).Attributes
	if _, ok := pricing["load_balancer"]; !ok {
		t.Error("control plane pricing_ids must include load_balancer")
	}
}

func TestKubernetesNodePoolSchema_ChangeEffects(t *testing.T) {
	s := validSchema(t, NewResourceKubernetesNodePool())
	replaced, rejected := changeEffects(t, s)

	wantReplaced := map[string]bool{
		"cluster_id": true, "name": true, "username": true, "node_pool_spec": true,
	}
	if !reflect.DeepEqual(replaced, wantReplaced) {
		t.Errorf("RequiresReplace attributes = %v, want %v", replaced, wantReplaced)
	}
	if !reflect.DeepEqual(rejected, map[string]bool{"password": true}) {
		t.Errorf("rejected attributes = %v, want only password", rejected)
	}

	if !s.Attributes["password"].IsSensitive() {
		t.Error("password must be sensitive")
	}
	if !s.Attributes["target_version"].IsComputed() || s.Attributes["target_version"].IsRequired() {
		t.Error("node pool target_version follows the cluster and must be computed only")
	}
	pricing := s.Attributes["node_pool_spec"].(schema.SingleNestedAttribute).
		Attributes["pricing_ids"].(schema.SingleNestedAttribute).Attributes
	if _, ok := pricing["load_balancer"]; ok {
		t.Error("node pool pricing_ids must not include load_balancer")
	}
}

func TestKubernetesSpecs_RoundTrip(t *testing.T) {
	cp := api.ControlPlaneSpec{
		ImageId: "img-1", StorageSizeGib: 50,
		PricingIds: api.ControlPlanePricingIds{
			VmAllocation: "p-vm", BlockStorage: "p-bs", PublicIp: "p-ip", LoadBalancer: "p-lb",
		},
	}
	if got := controlPlaneSpecToAPI(controlPlaneSpecFromAPI(cp)); !reflect.DeepEqual(got, cp) {
		t.Errorf("control plane spec round trip = %+v, want %+v", got, cp)
	}

	np := api.NodePoolSpec{
		ImageId:        "img-2",
		StorageSizeGib: 100,
		PricingIds: api.NodePoolPricingIds{
			VmAllocation: "p-vm",
			BlockStorage: "p-bs",
			PublicIp:     "p-ip",
		},
	}
	if got := nodePoolSpecToAPI(nodePoolSpecFromAPI(np)); !reflect.DeepEqual(got, np) {
		t.Errorf("node pool spec round trip = %+v, want %+v", got, np)
	}
}

func TestKubernetesGetResponseToModel_KeepsPassword(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 16, 0, 0, 0, 0, time.UTC)
	endpoint := "https://10.0.0.1:6443"

	cluster := ResourceKubernetesClusterModel{Password: types.StringValue("s3cret")}
	diags := resourceKubernetesClusterGetResponseToModel(
		ctx,
		&api.ResourceKubernetesClusterGetResponse{
			Id:               uuid.New(),
			Tags:             map[string]string{},
			Created:          now,
			ZoneId:           uuid.New(),
			OrganizationId:   uuid.New(),
			Name:             "prod",
			TargetVersion:    "v1.36.0",
			CurrentVersion:   "v1.36.0",
			AttachedSubnetId: uuid.New(),
			TargetReplicas:   3,
			CurrentReplicas:  3,
			Username:         "admin",
			Endpoint:         &endpoint,
			Ready:            &now,
			Status:           "ready",
		},
		&cluster,
	)
	if diags.HasError() {
		t.Fatalf("cluster diags: %v", diags)
	}
	if cluster.Password.ValueString() != "s3cret" {
		t.Error("cluster password must be kept from plan/state")
	}
	if cluster.Endpoint.ValueString() != endpoint || cluster.Ready.IsNull() ||
		!cluster.Deleted.IsNull() {
		t.Errorf("unexpected cluster model: %+v", cluster)
	}

	pool := ResourceKubernetesNodePoolModel{Password: types.StringValue("s3cret")}
	diags = resourceKubernetesNodePoolGetResponseToModel(
		ctx,
		&api.ResourceKubernetesNodePoolGetResponse{
			Id:              uuid.New(),
			Tags:            map[string]string{},
			Created:         now,
			ZoneId:          uuid.New(),
			OrganizationId:  uuid.New(),
			ClusterId:       uuid.New(),
			Name:            "workers",
			TargetVersion:   "v1.36.0",
			CurrentVersion:  "v1.35.2",
			TargetReplicas:  4,
			CurrentReplicas: 2,
			Username:        "admin",
			Status:          "provisioning",
		},
		&pool,
	)
	if diags.HasError() {
		t.Fatalf("node pool diags: %v", diags)
	}
	if pool.Password.ValueString() != "s3cret" {
		t.Error("node pool password must be kept from plan/state")
	}
	if pool.TargetVersion.ValueString() != "v1.36.0" || pool.CurrentReplicas.ValueInt64() != 2 ||
		!pool.Ready.IsNull() {
		t.Errorf("unexpected node pool model: %+v", pool)
	}
}

func TestReplicasAndTagsPatch(t *testing.T) {
	ctx := context.Background()
	tags := func(m map[string]string) types.Map {
		v, _ := types.MapValueFrom(ctx, types.StringType, m)
		return v
	}
	base := tags(map[string]string{"env": "dev"})

	cases := []struct {
		name                           string
		planReplicas                   int64
		planTags                       types.Map
		wantReplicasSent, wantTagsSent bool
	}{
		{name: "no change", planReplicas: 3, planTags: base},
		{name: "scale", planReplicas: 1, planTags: base, wantReplicasSent: true},
		{
			name:         "tags cleared",
			planReplicas: 3,
			planTags:     tags(map[string]string{}),
			wantTagsSent: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			replicasPtr, tagsPtr, diags := replicasAndTagsPatch(
				ctx, types.Int64Value(tc.planReplicas), types.Int64Value(3), tc.planTags, base,
			)
			if diags.HasError() {
				t.Fatalf("diags: %v", diags)
			}
			if (replicasPtr != nil) != tc.wantReplicasSent {
				t.Errorf("replicas sent = %v, want %v", replicasPtr != nil, tc.wantReplicasSent)
			}
			if replicasPtr != nil && *replicasPtr != tc.planReplicas {
				t.Errorf("replicas = %d, want %d", *replicasPtr, tc.planReplicas)
			}
			if (tagsPtr != nil) != tc.wantTagsSent {
				t.Errorf("tags sent = %v, want %v", tagsPtr != nil, tc.wantTagsSent)
			}
			if tagsPtr != nil && len(*tagsPtr) != 0 {
				t.Errorf("cleared tags must be sent as an empty map, got %v", *tagsPtr)
			}
		})
	}
}

func TestWaitDeleted_TreatsNotFoundAsDeleted(t *testing.T) {
	notFound := func() (*string, error) {
		return nil, &api.APIError{HttpCode: 404}
	}
	if diags := waitDeleted(notFound, 1); diags.HasError() {
		t.Errorf("404 should count as deleted: %v", diags)
	}

	deleted := func() (*string, error) {
		s := "deleted"
		return &s, nil
	}
	if diags := waitDeleted(deleted, 1); diags.HasError() {
		t.Errorf("status deleted should finish: %v", diags)
	}
}

func TestWaitUntil(t *testing.T) {
	status := func(s string) func() (*string, error) {
		return func() (*string, error) { return &s, nil }
	}

	if diags := waitUntil(status("ready"), "ready", 1); diags.HasError() {
		t.Errorf("reaching the target should succeed: %v", diags)
	}

	// maxRetry is large, so passing means it stopped at the first poll instead of retrying.
	for _, s := range []string{"deleting", "deleted"} {
		diags := waitUntil(status(s), "ready", 1000)
		if !diags.HasError() || !strings.Contains(diags[len(diags)-1].Detail(), s) {
			t.Errorf("status %s should fail right away, got %v", s, diags)
		}
	}
}
