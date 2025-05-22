package resource

import (
	"context"
	"fmt"
	"math"
	"terraform-provider-eci/internal/api"
	. "terraform-provider-eci/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ResourceSubnetModel struct {
	Id                types.String `tfsdk:"id"`
	Tags              types.Map    `tfsdk:"tags"`
	Created           types.String `tfsdk:"created"`
	Modified          types.String `tfsdk:"modified"`
	ZoneId            types.String `tfsdk:"zone_id"`
	OrganizationId    types.String `tfsdk:"organization_id"`
	AttachedNetworkId types.String `tfsdk:"attached_network_id"`
	Name              types.String `tfsdk:"name"`
	Purpose           types.String `tfsdk:"purpose"`
	NetworkGw         types.String `tfsdk:"network_gw"`
	Activated         types.String `tfsdk:"activated"`
	Deleted           types.String `tfsdk:"deleted"`
	Status            types.String `tfsdk:"status"`
}

var _ resource.Resource = &ResourceSubnet{}

type ResourceSubnet struct {
	client *api.APIClient
}

func resourceSubnetGetResponseToSubnetModel(
	ctx context.Context, response *api.ResourceSubnetGetResponse, data *ResourceSubnetModel,
) diag.Diagnostics {
	data.Id = types.StringValue(response.Id.String())
	tags, diags := types.MapValueFrom(ctx, types.StringType, response.Tags)

	if diags.HasError() {
		return diags
	}

	data.Tags = tags
	data.Created = types.StringValue(response.Created.String())
	data.Modified = StringOrNull(response.Modified)
	data.ZoneId = types.StringValue(response.ZoneId.String())
	data.OrganizationId = types.StringValue(response.OrganizationId.String())

	data.AttachedNetworkId = types.StringValue(
		string(response.AttachedNetworkId.String()),
	)
	data.Activated = StringOrNull(response.Activated)
	data.Deleted = StringOrNull(response.Deleted)
	data.Status = types.StringValue(response.Status)
	data.Name = types.StringValue(response.Name)
	data.Purpose = types.StringValue(response.Purpose)
	data.NetworkGw = types.StringValue(response.NetworkGw)

	return diag.Diagnostics{}
}

func NewResourceSubnet() resource.Resource {
	return &ResourceSubnet{}
}

func (r *ResourceSubnet) Metadata(
	ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

func (r *ResourceSubnet) Schema(
	ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Subnet",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "unique identifier of the subnet",
				Required:      false,
				Computed:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tags": schema.MapAttribute{
				Description: "User-defined metadata of key-value pairs",
				ElementType: types.StringType,
				Required:    true,
			},
			"created": schema.StringAttribute{
				Description:   "the time when the subnet is created",
				Required:      false,
				Computed:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"modified": schema.StringAttribute{
				Description: "last time when the subnet is modified",
				Required:    false,
				Computed:    true,
				Optional:    false,
			},
			"zone_id": schema.StringAttribute{
				Description:   "id of zone that the subnet belongs to",
				Required:      false,
				Computed:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Description:   "id of zone that the organization belongs to",
				Required:      false,
				Computed:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"attached_network_id": schema.StringAttribute{
				Description:   "id of virtual network that the subnet belongs to",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"activated": schema.StringAttribute{Required: false, Computed: true, Optional: false},
			"deleted": schema.StringAttribute{
				Required:      false,
				Computed:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"status": schema.StringAttribute{Computed: true, Required: false, Optional: false},
			"name": schema.StringAttribute{
				Description: "human-readable name of the subnet",
				Required:    true,
			},
			"purpose": schema.StringAttribute{
				Description:   "purpose of the subnet, e.g., `virtual_machine`",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"network_gw": schema.StringAttribute{
				Description:   "IPv4 interface address of the subnet, e.g., `192.168.0.1/24`",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *ResourceSubnet) Configure(
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

func (r *ResourceSubnet) Create(
	ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse,
) {
	var plan ResourceSubnetModel
	var state ResourceSubnetModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tags := map[string]string{}
	resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.PostSubnet(
		plan.Name.ValueString(),
		plan.AttachedNetworkId.ValueString(),
		plan.Purpose.ValueString(),
		plan.NetworkGw.ValueString(),
		tags,
	)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to create subnet", "", err)
		return
	}

	id := response.Id.String()
	plan.Id = types.StringValue(id)

	tflog.Info(ctx, fmt.Sprintf("successfully created a virtual subnet: %s", id))

	getResponse, err := r.client.GetSubnet(id)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to get subnet", id, err)
		return
	}

	resp.Diagnostics.Append(
		resourceSubnetGetResponseToSubnetModel(ctx, getResponse, &state)...,
	)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourceSubnet) Read(
	ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse,
) {
	var state ResourceSubnetModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	response, err := r.client.GetSubnet(id)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to get subnet", id, err)
		return
	}

	resp.Diagnostics.Append(
		resourceSubnetGetResponseToSubnetModel(ctx, response, &state)...,
	)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourceSubnet) Update(
	ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse,
) {
	var plan ResourceSubnetModel
	var state ResourceSubnetModel

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

	var tagsPtr *map[string]string = nil
	if !plan.Tags.Equal(state.Tags) {
		tags := map[string]string{}
		resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		tagsPtr = &tags
	}

	_, err := r.client.PatchSubnet(id, namePtr, tagsPtr)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to patch a subnet", id, err)
		return
	}
	tflog.Info(ctx, fmt.Sprintf("successfully patched a subnet: %s", id))

	getResponse, err := r.client.GetSubnet(id)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to get a subnet", id, err)
		return
	}

	resp.Diagnostics.Append(
		resourceSubnetGetResponseToSubnetModel(ctx, getResponse, &state)...,
	)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourceSubnet) Delete(
	ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse,
) {
	var state ResourceSubnetModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
	id := state.Id.ValueString()

	var err error
	for retryIndex := 0; retryIndex < 10; retryIndex += 1 {
		_, err = r.client.DeleteSubnet(id)
		successMessage, err := isResourceDeleted(err, "resource_subnet", "deleted")
		if err == nil {
			tflog.Info(ctx, fmt.Sprintf("%s (subnet: %s)", successMessage, id))
			return
		}

		time.Sleep(time.Duration(min(0.5+math.Pow(2, float64(retryIndex)), 10)) * time.Second)
	}

	addResourceError(&resp.Diagnostics, "failed to delete a subnet", id, err)
}
