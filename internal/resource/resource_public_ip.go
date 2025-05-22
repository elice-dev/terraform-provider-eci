package resource

import (
	"context"
	"fmt"
	"terraform-provider-eci/internal/api"
	. "terraform-provider-eci/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const maxRetry = 10

type ResourcePublicIpModel struct {
	Id                         types.String `tfsdk:"id"`
	Tags                       types.Map    `tfsdk:"tags"`
	Created                    types.String `tfsdk:"created"`
	Modified                   types.String `tfsdk:"modified"`
	ZoneId                     types.String `tfsdk:"zone_id"`
	OrganizationId             types.String `tfsdk:"organization_id"`
	DR                         types.Bool   `tfsdk:"dr"`
	AttachedNetworkInterfaceId types.String `tfsdk:"attached_network_interface_id"`
	PoolId                     types.String `tfsdk:"pool_id"`
	DrPoolId                   types.String `tfsdk:"dr_pool_id"`
	Ip                         types.String `tfsdk:"ip"`
	DrIp                       types.String `tfsdk:"dr_ip"`
	Deleted                    types.String `tfsdk:"deleted"`
	Status                     types.String `tfsdk:"status"`
}

var _ resource.Resource = &ResourcePublicIp{}

type ResourcePublicIp struct {
	client *api.APIClient
}

func resourcePublicIpGetResponseToPublicIpModel(
	ctx context.Context,
	response *api.ResourcePublicIpGetResponse,
	data *ResourcePublicIpModel,
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
	data.DR = types.BoolValue(response.DR)
	data.AttachedNetworkInterfaceId = StringOrNull(response.AttachedNetworkInterfaceId)
	data.PoolId = types.StringValue(response.PoolId.String())
	data.DrPoolId = StringOrNull(response.DrPoolId)
	data.Deleted = StringOrNull(response.Deleted)
	data.Status = types.StringValue(response.Status)
	data.Ip = types.StringValue(response.Ip)
	data.DrIp = StringValOrNull(response.DrIp)

	return diag.Diagnostics{}
}

func NewResourcePublicIp() resource.Resource {
	return &ResourcePublicIp{}
}

func (r *ResourcePublicIp) Metadata(
	ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_public_ip"
}

func (r *ResourcePublicIp) Schema(
	ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Public IP",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "unique identifier of the public ip",
				Required:      false,
				Computed:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tags": schema.MapAttribute{ElementType: types.StringType, Required: true},
			"created": schema.StringAttribute{
				Description:   "the time when the public ip is created",
				Required:      false,
				Computed:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"modified": schema.StringAttribute{
				Description: "the last time when the public ip is modified",
				Computed:    true,
			},
			"zone_id": schema.StringAttribute{
				Description:   "id of zone that the public ip belongs to",
				Required:      false,
				Computed:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Description:   "id of organization that the public ip belongs to",
				Required:      false,
				Computed:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"attached_network_interface_id": schema.StringAttribute{
				Description: "id of network interface that the public ip attaches to",
				Required:    true,
			},
			"dr": schema.BoolAttribute{
				Description:   "whether to enable DR support",
				Required:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			},
			"pool_id":    schema.StringAttribute{Computed: true},
			"dr_pool_id": schema.StringAttribute{Computed: true},
			"deleted": schema.StringAttribute{
				Description: "the time when the public ip is deleted",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "status of the public ip",
				Computed:    true,
			},
			"ip": schema.StringAttribute{
				Description: "the public ip address",
				Computed:    true,
			},
			"dr_ip": schema.StringAttribute{
				Description: "the public ip address available in DR mode",
				Computed:    true,
			},
		},
	}
}

func (r *ResourcePublicIp) Configure(
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

func (r *ResourcePublicIp) Create(
	ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse,
) {
	var plan ResourcePublicIpModel
	var state ResourcePublicIpModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tags := map[string]string{}
	resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.PostPublicIp(plan.DR.ValueBool(), tags)

	if err != nil {
		addResourceError(
			&resp.Diagnostics,
			"failed to create a public ip",
			"",
			fmt.Errorf("API Error: %s", err.Error()),
		)
		return
	}

	id := response.Id.String()
	tflog.Info(ctx, fmt.Sprintf("created a public ip: %s", id))

	if !plan.AttachedNetworkInterfaceId.IsNull() {
		attachedNetworkInterfaceIdPtr := plan.AttachedNetworkInterfaceId.ValueStringPointer()
		_, err := r.client.PatchPublicIp(id, &attachedNetworkInterfaceIdPtr, nil)

		if err != nil {
			addResourceError(&resp.Diagnostics, "failed to patch public ip", id, err)
			return
		}

		tflog.Trace(
			ctx,
			fmt.Sprintf(
				"public ip (%s) attached to a public ip (%s)",
				id,
				*attachedNetworkInterfaceIdPtr,
			),
		)
	}

	getResponse, err := r.client.GetPublicIp(id)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to get public ip", id, err)
		return
	}

	resp.Diagnostics.Append(
		resourcePublicIpGetResponseToPublicIpModel(ctx, getResponse, &state)...,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if getResponse.Status == "active" {
		tflog.Info(ctx, fmt.Sprintf("public ip (%s) is active", id))
		return
	}

	_, diags := waitStatus(
		func() (*string, error) {
			getResponse, err := r.client.GetPublicIp(id)
			if err != nil {
				return nil, err
			}
			return &getResponse.Status, nil
		},
		[]string{"active"},
		maxRetry,
	)
	resp.Diagnostics.Append(diags...)
}

func (r *ResourcePublicIp) Read(
	ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse,
) {
	var state ResourcePublicIpModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	response, err := r.client.GetPublicIp(id)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to get a public ip", id, err)
		return
	}

	resp.Diagnostics.Append(
		resourcePublicIpGetResponseToPublicIpModel(ctx, response, &state)...,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourcePublicIp) Update(
	ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse,
) {
	var plan ResourcePublicIpModel
	var state ResourcePublicIpModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()

	var attachedMachineIdPtr *string = nil
	if !plan.AttachedNetworkInterfaceId.Equal(state.AttachedNetworkInterfaceId) {
		if !state.AttachedNetworkInterfaceId.IsNull() && !plan.AttachedNetworkInterfaceId.IsNull() {
			_, err := r.client.PatchPublicIp(id, &attachedMachineIdPtr, nil)
			if err != nil {
				addResourceError(&resp.Diagnostics, "failed to patch a public ip", id, err)
				return
			}
		}

		attachedMachineIdPtr = plan.AttachedNetworkInterfaceId.ValueStringPointer()
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

	_, err := r.client.PatchPublicIp(id, &attachedMachineIdPtr, tagsPtr)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to patch a public ip", id, err)
		return
	}
	tflog.Info(ctx, fmt.Sprintf("successfully patched a public ip: %s", id))

	getResponse, err := r.client.GetPublicIp(id)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to get a public ip", id, err)
		return
	}

	resp.Diagnostics.Append(
		resourcePublicIpGetResponseToPublicIpModel(ctx, getResponse, &state)...,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourcePublicIp) Delete(
	ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse,
) {
	var state ResourcePublicIpModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()

	if !state.AttachedNetworkInterfaceId.IsNull() {
		var attachedNetworkInterfaceId *string = nil
		_, err := r.client.PatchPublicIp(id, &attachedNetworkInterfaceId, nil)

		if err != nil {
			addResourceError(
				&resp.Diagnostics,
				"failed to detach public ip from a network interface",
				id,
				err,
			)
			return
		}
	}

	_, err := r.client.DeletePublicIp(id)
	successMessage, err := isResourceDeleted(err, "resource_public_ip", "deleted")

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to delete a public ip", id, err)
		return
	}

	tflog.Info(ctx, fmt.Sprintf("%s (public ip: %s)", successMessage, id))
}
