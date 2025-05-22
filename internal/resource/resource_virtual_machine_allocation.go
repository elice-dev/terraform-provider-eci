package resource

import (
	"context"
	"fmt"
	"terraform-provider-eci/internal/api"
	. "terraform-provider-eci/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &ResourceVirtualMachineAllocation{}

type ResourceVirtualMachineAllocation struct {
	client *api.APIClient
}

func NewResourceVirtualMachineAllocation() resource.Resource {
	return &ResourceVirtualMachineAllocation{}
}

type ResourceVirtualMachineAllocationModel struct {
	Id   types.String `tfsdk:"id"`
	Tags types.Map    `tfsdk:"tags"`

	Created        types.String `tfsdk:"created"`
	Modified       types.String `tfsdk:"modified"`
	ZoneId         types.String `tfsdk:"zone_id"`
	OrganizationId types.String `tfsdk:"organization_id"`

	MachineId          types.String `tfsdk:"machine_id"`
	RequestedCpuVcore  types.Int64  `tfsdk:"requested_cpu_vcore"`
	RequestedMemoryGib types.Int64  `tfsdk:"requested_memory_gib"`
	RequestedDevices   types.List   `tfsdk:"requested_devices"`
	LastHeartbeat      types.String `tfsdk:"last_heartbeat"`
	Assigned           types.String `tfsdk:"assigned"`
	Taken              types.String `tfsdk:"taken"`
	Started            types.String `tfsdk:"started"`
	Terminating        types.String `tfsdk:"terminating"`
	Terminated         types.String `tfsdk:"terminated"`
	Status             types.String `tfsdk:"status"`
}

func resourceVirtualMachineAllocationGetResponseToVirtualMachineAllocationModel(
	ctx context.Context,
	response *api.ResourceVirtualMachineAllocationGetResponse,
	data *ResourceVirtualMachineAllocationModel,
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
	data.MachineId = types.StringValue(response.MachineId.String())
	data.RequestedCpuVcore = types.Int64Value(int64(response.RequestedCpuVcore))
	data.RequestedMemoryGib = types.Int64Value(int64(response.RequestedMemoryGib))

	requestedDevices, diags := types.ListValueFrom(ctx, types.StringType, response.RequestedDevices)
	if diags.HasError() {
		return diags
	}
	data.RequestedDevices = requestedDevices
	data.LastHeartbeat = StringOrNull(response.LastHeartbeat)
	data.Assigned = StringOrNull(response.Assigned)
	data.Taken = StringOrNull(response.Taken)
	data.Started = StringOrNull(response.Started)
	data.Terminating = StringOrNull(response.Terminating)
	data.Terminated = StringOrNull(response.Terminated)
	data.Status = types.StringValue(response.Status)

	return diag.Diagnostics{}
}

func (r *ResourceVirtualMachineAllocation) Metadata(
	ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_virtual_machine_allocation"
}

func (r *ResourceVirtualMachineAllocation) Schema(
	ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Virtual Machine Allocation",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "unique identifier of the virtual machine allocation",
				Computed:      true,
				Optional:      false,
				Required:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tags": schema.MapAttribute{
				ElementType:   types.StringType,
				Required:      true,
				PlanModifiers: []planmodifier.Map{mapplanmodifier.RequiresReplace()},
			},
			"created": schema.StringAttribute{
				Description:   "time when the virtual machine allocation is created",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"modified": schema.StringAttribute{
				Description: "last time when the virtual machine allocation is modified",
				Computed:    true,
			},
			"zone_id": schema.StringAttribute{
				Description:   "id of zone that the virtual machine allocation belongs to",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Description:   "id of zone that the organization allocation belongs to",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"machine_id": schema.StringAttribute{
				Description:   "id of virtual machine that this allocation is instantiated from",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"requested_cpu_vcore": schema.Int64Attribute{
				Description:   "number of CPU vCores assigned to the virtual machine",
				Computed:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"requested_memory_gib": schema.Int64Attribute{
				Description:   "size of memory (GiB) assigned to the virtual machine",
				Computed:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"requested_devices": schema.ListAttribute{
				ElementType:   types.StringType,
				Description:   "devices assigned to the virtual machine",
				Computed:      true,
				PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			},
			"last_heartbeat": schema.StringAttribute{
				Description:   "last time when a heartbeat from the virtual machine allocation is received",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"assigned": schema.StringAttribute{
				Description:   "the time when the virtual machine allocation is assigned to a host machine",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"taken": schema.StringAttribute{
				Description:   "the time when the virtual machine allocation is taken by a host machine",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"started": schema.StringAttribute{
				Description:   "the time when the virtual machine allocation is started",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"terminating": schema.StringAttribute{
				Description:   "the time when the virtual machine allocation enters `terminating` state",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"terminated": schema.StringAttribute{
				Description:   "the time when the virtual machine allocation is terminated",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"status": schema.StringAttribute{
				Description: "status of the virtual machine allocation",
				Computed:    true,
			},
		},
	}
}

func (r *ResourceVirtualMachineAllocation) Configure(
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

func (r *ResourceVirtualMachineAllocation) Create(
	ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse,
) {
	var plan ResourceVirtualMachineAllocationModel
	var state ResourceVirtualMachineAllocationModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	machineId := plan.MachineId.ValueString()
	machine, err := r.client.GetVirtualMachine(machineId)
	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to get virtual machine", machineId, err)
		return
	}

	if machine.AlwaysOn {
		resp.Diagnostics.AddError(
			"Virtual machine has invalid configuration",
			"VM allocation cannot be defined in terraform for VM with `always_on` enabled",
		)
		return
	}

	tagsPtr := map[string]string{}
	resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tagsPtr, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.PostVirtualMachineAllocation(machineId, tagsPtr)

	if err != nil {
		addResourceError(
			&resp.Diagnostics, "failed to create a virtual machine allocation", "", err,
		)
		return
	}

	id := response.Id.String()
	tflog.Trace(ctx, fmt.Sprintf("successfully created a virtual machine allocation: %s", id))

	getResponse, err := r.client.GetVirtualMachineAllocation(id)

	if err != nil {
		addResourceError(
			&resp.Diagnostics, "failed to get a virtual machine allocation", id, err,
		)
		return
	}

	resp.Diagnostics.Append(
		resourceVirtualMachineAllocationGetResponseToVirtualMachineAllocationModel(
			ctx, getResponse, &state,
		)...,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourceVirtualMachineAllocation) Read(
	ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse,
) {
	var state ResourceVirtualMachineAllocationModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	allocation, err := r.client.GetVirtualMachineAllocation(id)

	if err != nil {
		addResourceError(
			&resp.Diagnostics, "failed to get a virtual machine allocation", id, err,
		)
		return
	}

	resp.Diagnostics.Append(
		resourceVirtualMachineAllocationGetResponseToVirtualMachineAllocationModel(
			ctx, allocation, &state,
		)...,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourceVirtualMachineAllocation) Update(
	ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse,
) {
	resp.Diagnostics.AddError("update not supported", "")
}

func (r *ResourceVirtualMachineAllocation) Delete(
	ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse,
) {
	var state ResourceVirtualMachineAllocationModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	deleteResponse, err := r.client.DeleteVirtualMachineAllocation(id)
	successMessage, err := isResourceDeleted(err, "resource_allocation", "terminated")

	if err != nil {
		addResourceError(
			&resp.Diagnostics, "failed to delete a virtual machine allocation", id, err,
		)
		return
	}

	if deleteResponse.Status == "terminated" {
		tflog.Info(ctx, fmt.Sprintf("%s (id: %s)", successMessage, id))
		return
	}

	_, diags := waitStatus(
		func() (*string, error) {
			getResponse, err := r.client.GetVirtualMachineAllocation(id)
			if err != nil {
				return nil, err
			}
			return &getResponse.Status, nil
		},
		[]string{"terminated"},
		10,
	)

	resp.Diagnostics.Append(diags...)
}
