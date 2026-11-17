package resource

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func validSchema(t *testing.T, r fwresource.Resource) schema.Schema {
	t.Helper()
	ctx := context.Background()
	resp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diags: %v", resp.Diagnostics)
	}
	if diags := resp.Schema.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("invalid schema: %v", diags)
	}
	return resp.Schema
}

// changeEffects runs the plan modifiers of every top-level string and nested-object attribute
// for an update from one value to another, and reports which attributes make Terraform
// replace the resource and which are rejected with an error.
func changeEffects(t *testing.T, s schema.Schema) (replaced, rejected map[string]bool) {
	t.Helper()
	ctx := context.Background()
	existing := tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{})
	state := tfsdk.State{Raw: existing}
	plan := tfsdk.Plan{Raw: existing}

	replaced, rejected = map[string]bool{}, map[string]bool{}
	record := func(name string, requiresReplace bool, diags diag.Diagnostics) {
		if requiresReplace {
			replaced[name] = true
		}
		if diags.HasError() {
			rejected[name] = true
		}
	}

	for name, a := range s.Attributes {
		switch typed := a.(type) {
		case schema.StringAttribute:
			for _, m := range typed.PlanModifiers {
				req := planmodifier.StringRequest{
					State: state, Plan: plan,
					StateValue:  types.StringValue("a"),
					PlanValue:   types.StringValue("b"),
					ConfigValue: types.StringValue("b"),
				}
				resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
				m.PlanModifyString(ctx, req, resp)
				record(name, resp.RequiresReplace, resp.Diagnostics)
			}
		case schema.SingleNestedAttribute:
			attrTypes := map[string]attr.Type{"x": types.StringType}
			before := types.ObjectValueMust(
				attrTypes,
				map[string]attr.Value{"x": types.StringValue("a")},
			)
			after := types.ObjectValueMust(
				attrTypes,
				map[string]attr.Value{"x": types.StringValue("b")},
			)
			for _, m := range typed.PlanModifiers {
				req := planmodifier.ObjectRequest{
					State: state, Plan: plan,
					StateValue: before, PlanValue: after, ConfigValue: after,
				}
				resp := &planmodifier.ObjectResponse{PlanValue: req.PlanValue}
				m.PlanModifyObject(ctx, req, resp)
				record(name, resp.RequiresReplace, resp.Diagnostics)
			}
		}
	}
	return replaced, rejected
}
