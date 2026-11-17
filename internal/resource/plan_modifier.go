package resource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var _ planmodifier.String = errorOnChangeModifier{}

// errorOnChangeModifier rejects in-place changes that the API cannot apply, instead of
// replacing the resource (e.g. a whole kubernetes cluster).
type errorOnChangeModifier struct {
	reason string
}

func errorOnChange(reason string) planmodifier.String {
	return errorOnChangeModifier{reason: reason}
}

func (m errorOnChangeModifier) Description(_ context.Context) string {
	return "Changing this attribute after creation is an error: " + m.reason
}

func (m errorOnChangeModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m errorOnChangeModifier) PlanModifyString(
	_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse,
) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}
	if req.PlanValue.IsUnknown() || req.PlanValue.Equal(req.StateValue) {
		return
	}

	// Values are not included because the attribute may be sensitive.
	resp.Diagnostics.AddAttributeError(
		req.Path,
		"attribute cannot be changed after creation",
		m.reason,
	)
}
