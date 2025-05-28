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

type ResourceNetworkInterfaceModel struct {
	Id                types.String `tfsdk:"id"`
	Tags              types.Map    `tfsdk:"tags"`
	Created           types.String `tfsdk:"created"`
	Modified          types.String `tfsdk:"modified"`
	ZoneId            types.String `tfsdk:"zone_id"`
	OrganizationId    types.String `tfsdk:"organization_id"`
	AttachedSubnetId  types.String `tfsdk:"attached_subnet_id"`
	AttachedMachineId types.String `tfsdk:"attached_machine_id"`
	DR                types.Bool   `tfsdk:"dr"`
	Deleted           types.String `tfsdk:"deleted"`
	Status            types.String `tfsdk:"status"`
	Name              types.String `tfsdk:"name"`
	Ip                types.String `tfsdk:"ip"`
	Mac               types.String `tfsdk:"mac"`
}

var _ resource.Resource = &ResourceNetworkInterface{}

type ResourceNetworkInterface struct {
	client *api.APIClient
}

func resourceNetworkInterfaceGetResponseToNetworkInterfaceModel(
	ctx context.Context,
	response *api.ResourceNetworkInterfaceGetResponse,
	data *ResourceNetworkInterfaceModel,
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
	data.AttachedSubnetId = types.StringValue(response.AttachedSubnetId.String())
	data.DR = types.BoolValue(response.DR)
	data.AttachedMachineId = StringOrNull(response.AttachedMachineId)
	data.Deleted = StringOrNull(response.Deleted)
	data.Status = types.StringValue(response.Status)
	data.Name = types.StringValue(response.Name)
	data.Ip = types.StringValue(response.Ip)
	data.Mac = types.StringValue(response.Mac)

	return diag.Diagnostics{}
}

func NewResourceNetworkInterface() resource.Resource {
	return &ResourceNetworkInterface{}
}

func (r *ResourceNetworkInterface) Metadata(
	ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_network_interface"
}

func (r *ResourceNetworkInterface) Schema(
	ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Network Interface",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "unique identifier of the network interface",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tags": schema.MapAttribute{
				Description: "User-defined metadata of key-value pairs",
				ElementType: types.StringType,
				Required:    true,
			},
			"created": schema.StringAttribute{
				Description:   "the time when the network interface is created",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"modified": schema.StringAttribute{
				Description: "last time when the network interface is modified",
				Computed:    true,
			},
			"zone_id": schema.StringAttribute{
				Description:   "id of zone that the network interface belongs to",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Description:   "id of organization that the network interface belongs to",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"attached_subnet_id": schema.StringAttribute{
				Description:   "id of subnet that the network interface attaches to",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"attached_machine_id": schema.StringAttribute{
				Description: "id of virtual machine that the network interface attaches to",
				Required:    true,
			},
			"dr": schema.BoolAttribute{
				Description:   "whether to enable DR support",
				Required:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			},
			"deleted": schema.StringAttribute{
				Description:   "the time when the network interface is deleted",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"status": schema.StringAttribute{
				Description: "status of the network interface",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "human-readable name for the network interface",
				Required:    true,
			},
			"ip": schema.StringAttribute{
				Description: "IP address that the network interface uses",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"mac": schema.StringAttribute{
				Description: "MAC address that the network interface uses",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ResourceNetworkInterface) Configure(
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

func (r *ResourceNetworkInterface) Create(
	ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse,
) {
	var plan ResourceNetworkInterfaceModel
	var state ResourceNetworkInterfaceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var ipPtr *string = nil
	if !plan.Ip.IsUnknown() {
		ipPtr = plan.Ip.ValueStringPointer()
	}

	var macPtr *string = nil
	if !plan.Mac.IsUnknown() {
		macPtr = plan.Mac.ValueStringPointer()
	}

	tagsPtr := map[string]string{}
	resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tagsPtr, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.PostNetworkInterface(
		plan.Name.ValueString(),
		plan.AttachedSubnetId.ValueString(),
		plan.DR.ValueBool(),
		ipPtr,
		macPtr,
		tagsPtr,
	)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to create network interface", "", err)
		return
	}

	id := response.Id.String()
	tflog.Info(ctx, fmt.Sprintf("successfully created a network interface: %s", id))

	if !plan.AttachedMachineId.IsNull() {
		attachedMachineIdPtr := plan.AttachedMachineId.ValueStringPointer()

		vmResponse, err := r.client.GetVirtualMachine(*attachedMachineIdPtr)
		if err != nil {
			addResourceError(
				&resp.Diagnostics,
				"failed to get a virtual machine",
				*attachedMachineIdPtr,
				err,
			)
			return
		}

		if vmResponse.Status != "idle" {
			resp.Diagnostics.AddError(
				"virtual machine is not idle",
				fmt.Sprintf(
					"virtual machine (%s) is not idle, status: %s (tip: remove the virtual machine allocation)",
					*attachedMachineIdPtr,
					vmResponse.Status,
				),
			)
			return
		}

		_, err = r.client.PatchNetworkInterface(id, nil, &attachedMachineIdPtr, nil)

		if err != nil {
			addResourceError(&resp.Diagnostics, "failed to patch network interface", id, err)
			return
		}

		tflog.Trace(
			ctx,
			fmt.Sprintf(
				"network interface (%s) attached to a virtual machine (%s)",
				id,
				*attachedMachineIdPtr,
			),
		)
	}

	getResponse, err := r.client.GetNetworkInterface(id)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to get a network interface", id, err)
		return
	}

	resp.Diagnostics.Append(
		resourceNetworkInterfaceGetResponseToNetworkInterfaceModel(ctx, getResponse, &state)...,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	if resp.Diagnostics.HasError() || getResponse.Status == "active" {
		return
	}

	_, diags := waitStatus(
		func() (*string, error) {
			getResponse, err := r.client.GetNetworkInterface(id)
			if err != nil {
				return nil, err
			}
			return &getResponse.Status, nil
		},
		[]string{"active"},
		10,
	)
	resp.Diagnostics.Append(diags...)
}

func (r *ResourceNetworkInterface) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state ResourceNetworkInterfaceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
	id := state.Id.ValueString()
	response, err := r.client.GetNetworkInterface(id)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to get a network interface", id, err)
		return
	}

	resp.Diagnostics.Append(
		resourceNetworkInterfaceGetResponseToNetworkInterfaceModel(ctx, response, &state)...,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourceNetworkInterface) Update(
	ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse,
) {
	var plan ResourceNetworkInterfaceModel
	var state ResourceNetworkInterfaceModel

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

	var attachedMachineIdPtr **string = nil
	if !plan.AttachedMachineId.Equal(state.AttachedMachineId) {
		if !state.AttachedMachineId.IsNull() && !plan.AttachedMachineId.IsNull() {
			var nilMachineIdPtr *string = nil

			_, err := r.client.PatchNetworkInterface(id, nil, &nilMachineIdPtr, nil)
			if err != nil {
				addResourceError(&resp.Diagnostics, "failed to patch a network interface", id, err)
				return
			}

			tflog.Info(
				ctx,
				fmt.Sprintf("network interface (%s) detached from a virtual machine", id),
			)
		}

		var attachedMachineId = plan.AttachedMachineId.ValueStringPointer()
		attachedMachineIdPtr = &attachedMachineId
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

	_, err := r.client.PatchNetworkInterface(id, namePtr, attachedMachineIdPtr, tagsPtr)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to patch a network interface", id, err)
		return
	}

	tflog.Info(ctx, fmt.Sprintf("successfully patched a network interface: %s", id))
	getResponse, err := r.client.GetNetworkInterface(id)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to patch network interface", id, err)
		return
	}

	resp.Diagnostics.Append(
		resourceNetworkInterfaceGetResponseToNetworkInterfaceModel(ctx, getResponse, &state)...,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourceNetworkInterface) Delete(
	ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse,
) {
	var state ResourceNetworkInterfaceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()

	if !state.AttachedMachineId.IsNull() {
		var attachedMachineIdPtr *string = nil
		_, err := r.client.PatchNetworkInterface(id, nil, &attachedMachineIdPtr, nil)

		if err != nil {
			addResourceError(&resp.Diagnostics, "failed to detach from a virtual machine", id, err)
			return
		}
	}

	publicIps, err := r.client.GetPublicIps(&id)
	if err != nil {
		addResourceError(
			&resp.Diagnostics,
			"failed to get list of public ips attached to a network interface",
			id,
			err,
		)
		return
	}

	var nilAttachedNetworkInterface *string = nil
	for _, publicIp := range publicIps {
		_, err = r.client.PatchPublicIp(publicIp.Id.String(), &nilAttachedNetworkInterface, nil)
		if err != nil {
			addResourceError(
				&resp.Diagnostics,
				"failed to patch a public ip to detach its network interface",
				publicIp.Id.String(),
				err,
			)
		}
	}

	_, err = r.client.DeleteNetworkInterface(id)
	successMessage, err := isResourceDeleted(err, "resource_network_interface", "deleted")

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to delete a network interface", id, err)
		return
	}

	tflog.Info(ctx, fmt.Sprintf("%s (network interface: %s)", successMessage, id))
}
