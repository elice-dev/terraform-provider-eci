package resource

import (
	"context"
	"fmt"
	"terraform-provider-eci/internal/api"
	. "terraform-provider-eci/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ResourceVirtualNetworkModel struct {
	Id             types.String `tfsdk:"id"`
	Tags           types.Map    `tfsdk:"tags"`
	Created        types.String `tfsdk:"created"`
	Modified       types.String `tfsdk:"modified"`
	ZoneId         types.String `tfsdk:"zone_id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	Deleted        types.String `tfsdk:"deleted"`
	Status         types.String `tfsdk:"status"`
	Name           types.String `tfsdk:"name"`
	NetworkCidr    types.String `tfsdk:"network_cidr"`
	FirewallRules  types.List   `tfsdk:"firewall_rules"`
}

type FirewallRuleModel struct {
	Proto       types.String `tfsdk:"proto"`
	Source      types.String `tfsdk:"source"`
	Destination types.String `tfsdk:"destination"`
	Port        types.Int64  `tfsdk:"port"`
	PortEnd     types.Int64  `tfsdk:"port_end"`
	Action      types.String `tfsdk:"action"`
	Comment     types.String `tfsdk:"comment"`
}

var _ resource.Resource = &ResourceVirtualNetwork{}

type ResourceVirtualNetwork struct {
	client *api.APIClient
}

func resourceVirtualNetworkGetResponseToVirtualNetworkModel(
	ctx context.Context,
	response *api.ResourceVirtualNetworkGetResponse,
	data *ResourceVirtualNetworkModel,
) diag.Diagnostics {
	data.Id = types.StringValue(response.Id.String())
	tags, diags := types.MapValueFrom(ctx, types.StringType, response.Tags)

	if diags.HasError() {
		return diags
	}

	data.Tags = tags
	data.Name = types.StringValue(response.Name)
	data.Created = types.StringValue(response.Created.String())
	data.Modified = StringOrNull(response.Modified)
	data.ZoneId = types.StringValue(response.ZoneId.String())
	data.OrganizationId = types.StringValue(response.OrganizationId.String())
	data.Deleted = StringOrNull(response.Deleted)
	data.Status = types.StringValue(string(response.Status))
	data.NetworkCidr = types.StringValue(response.NetworkCidr)

	objectType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"proto":       types.StringType,
			"source":      types.StringType,
			"destination": types.StringType,
			"port":        types.Int64Type,
			"port_end":    types.Int64Type,
			"action":      types.StringType,
			"comment":     types.StringType,
		},
	}

	firewallRules, diags := types.ListValueFrom(ctx, objectType, response.FirewallRules)
	data.FirewallRules = firewallRules

	return diags
}

func NewResourceVirtualNetwork() resource.Resource {
	return &ResourceVirtualNetwork{}
}

func (r *ResourceVirtualNetwork) Metadata(
	ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_virtual_network"
}

func (r *ResourceVirtualNetwork) Schema(
	ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Virtual Network",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "unique identifier of the virtual network",
				Computed:      true,
				Optional:      false,
				Required:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tags": schema.MapAttribute{
				Description: "User-defined metadata of key-value pairs",
				ElementType: types.StringType,
				Required:    true,
			},
			"created": schema.StringAttribute{
				Description:   "the time when the virtual network is created",
				Computed:      true,
				Optional:      false,
				Required:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"modified": schema.StringAttribute{
				Description: "the time when the virtual network is created",
				Computed:    true,
				Optional:    false,
				Required:    false,
			},
			"zone_id": schema.StringAttribute{
				Description:   "id of the zone that the virtual network belongs to",
				Required:      false,
				Computed:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Description:   "id of the organization that the virtual network belongs to",
				Required:      false,
				Computed:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"deleted": schema.StringAttribute{
				Description: "the time when the virtual network is deleted",
				Computed:    true,
				Required:    false,
				Optional:    false,
			},
			"status": schema.StringAttribute{
				Description: "status of the virtual network",
				Computed:    true,
				Required:    false,
				Optional:    false,
			},
			"name": schema.StringAttribute{
				Description: "human-readable name of the virtual network",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.LengthAtMost(256),
				},
			},
			"network_cidr": schema.StringAttribute{
				Description:   "CIDR of the virtual network (e.g., 192.168.0.0/16)",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"firewall_rules": schema.ListNestedAttribute{
				Description: "list of the firewall rules",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"proto": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf("ALL", "ICMP", "TCP", "UDP"),
							},
						},
						"source":      schema.StringAttribute{Required: true},
						"destination": schema.StringAttribute{Required: true},
						"port": schema.Int64Attribute{
							Optional: true,
							Validators: []validator.Int64{
								int64validator.AtLeast(0),
								int64validator.AtMost(65535),
							},
						},
						"port_end": schema.Int64Attribute{
							Optional: true,
							Validators: []validator.Int64{
								int64validator.AtLeast(0),
								int64validator.AtMost(65535),
							},
						},
						"action": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf("ACCEPT", "DROP"),
							},
						},
						"comment": schema.StringAttribute{
							Description: "human-readable comment of the firewall rule",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.LengthAtMost(256),
							},
						},
					},
				},
			},
		},
	}
}

func (r *ResourceVirtualNetwork) Configure(
	ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.APIClient)

	if !ok {
		resp.Diagnostics.AddError(
			"unexpected resource configure type",
			fmt.Sprintf(`expected *api.APIClient, got: %T.`, req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *ResourceVirtualNetwork) Create(
	ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse,
) {
	var plan ResourceVirtualNetworkModel
	var state ResourceVirtualNetworkModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tags := map[string]string{}
	resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.PostVirtualNetwork(
		plan.Name.ValueString(),
		plan.NetworkCidr.ValueString(),
		tags,
	)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to create a virtual network", "", err)
		return
	}

	id := response.Id.String()
	tflog.Info(ctx, fmt.Sprintf("successfully created a virtual network: %s", id))

	var rules = []api.NetworkFirewallRule{}
	if !plan.FirewallRules.IsUnknown() {
		resp.Diagnostics.Append(
			plan.FirewallRules.ElementsAs(ctx, &rules, false)...,
		)

		if resp.Diagnostics.HasError() {
			return
		}
	}

	_, err = r.client.PatchVirtualNetwork(id, nil, &rules, nil)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to patch a virtual network", "", err)
		return
	}

	getResponse, err := r.client.GetVirtualNetwork(id)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to get a virtual network", id, err)
		return
	}

	resp.Diagnostics.Append(
		resourceVirtualNetworkGetResponseToVirtualNetworkModel(ctx, getResponse, &state)...,
	)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourceVirtualNetwork) Read(
	ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse,
) {
	var state ResourceVirtualNetworkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	response, err := r.client.GetVirtualNetwork(id)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to get a virtual network", id, err)
		return
	}

	resp.Diagnostics.Append(
		resourceVirtualNetworkGetResponseToVirtualNetworkModel(ctx, response, &state)...,
	)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourceVirtualNetwork) Update(
	ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse,
) {
	var plan ResourceVirtualNetworkModel
	var state ResourceVirtualNetworkModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
	id := state.Id.ValueString()

	var namePtr *string = nil
	if !plan.Name.Equal(state.Name) {
		namePtr = plan.Name.ValueStringPointer()
	}

	var firewallRulesPtr *[]api.NetworkFirewallRule = nil
	if !plan.FirewallRules.Equal(state.FirewallRules) {
		var firewallRules []api.NetworkFirewallRule

		resp.Diagnostics.Append(plan.FirewallRules.ElementsAs(ctx, &firewallRules, false)...)

		if resp.Diagnostics.HasError() {
			return
		}
		firewallRulesPtr = &firewallRules
	}

	var tagsPtr *map[string]string = nil
	if !plan.Tags.Equal(state.Tags) {
		tags := map[string]string{}
		resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		tagsPtr = &tags
	}

	_, err := r.client.PatchVirtualNetwork(id, namePtr, firewallRulesPtr, tagsPtr)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to patch a virtual network", id, err)
		return
	}

	tflog.Info(ctx, fmt.Sprintf("successfully patched a virtual network: %s", id))

	getResponse, err := r.client.GetVirtualNetwork(id)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to get a virtual network", id, err)
		return
	}

	resp.Diagnostics.Append(
		resourceVirtualNetworkGetResponseToVirtualNetworkModel(
			ctx, getResponse, &state,
		)...,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourceVirtualNetwork) Delete(
	ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse,
) {
	var state ResourceVirtualNetworkModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	_, err := r.client.DeleteVirtualNetwork(id)

	successMessage, err := isResourceDeleted(err, "resource_virtual_network", "deleted")
	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to delete a virtual network", id, err)
		return
	}

	tflog.Info(ctx, fmt.Sprintf("%s (virtual network: %s)", successMessage, id))
}
