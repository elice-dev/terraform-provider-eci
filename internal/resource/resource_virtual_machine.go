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

var _ resource.Resource = &ResourceVirtualMachine{}

type ResourceVirtualMachine struct {
	client *api.APIClient
}

func NewResourceVirtualMachine() resource.Resource {
	return &ResourceVirtualMachine{}
}

type ResourceVirtualMachineModel struct {
	Id             types.String `tfsdk:"id"`
	Tags           types.Map    `tfsdk:"tags"`
	Created        types.String `tfsdk:"created"`
	Modified       types.String `tfsdk:"modified"`
	ZoneId         types.String `tfsdk:"zone_id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	InstanceTypeId types.String `tfsdk:"instance_type_id"`

	AlwaysOn types.Bool   `tfsdk:"always_on"`
	Name     types.String `tfsdk:"name"`
	DR       types.Bool   `tfsdk:"dr"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`

	Allocated types.String `tfsdk:"allocated"`
	Deleted   types.String `tfsdk:"deleted"`
	Status    types.String `tfsdk:"status"`
}

func resourceVirtualMachineGetResponseToVirtualMachineModel(
	ctx context.Context,
	response *api.ResourceVirtualMachineGetResponse,
	data *ResourceVirtualMachineModel,
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
	data.InstanceTypeId = types.StringValue(response.InstanceTypeId.String())
	data.AlwaysOn = types.BoolValue(response.AlwaysOn)
	data.DR = types.BoolValue(response.DR)
	data.Allocated = StringOrNull(response.Allocated)
	data.Deleted = StringOrNull(response.Deleted)
	data.Status = types.StringValue(string(response.Status))
	data.Username = types.StringValue(response.Username)

	return diag.Diagnostics{}
}

func (r *ResourceVirtualMachine) Metadata(
	ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_virtual_machine"
}

func (r *ResourceVirtualMachine) Schema(
	ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Virtual Machine",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "unique identifier of the virtual machine",
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
				Description:   "time when the virtual machine is created",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"modified": schema.StringAttribute{
				Description: "last time when the virtual machine is modified",
				Computed:    true,
			},
			"zone_id": schema.StringAttribute{
				Description:   "id of zone that the virtual machine belongs to",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Description:   "id of organization that the virtual machine belongs to",
				Required:      false,
				Computed:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"instance_type_id": schema.StringAttribute{
				Description: "id of instance type that the virtual machine is created from",
				Required:    true,
				Computed:    false,
				Optional:    false,
			},
			"always_on": schema.BoolAttribute{
				Description: "whether to automatically restart the virtual machine when migrated to DR",
				Required:    true,
			},
			"dr": schema.BoolAttribute{
				Description:   "whether to enable DR support",
				Required:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			},
			"allocated": schema.StringAttribute{Computed: true, Required: false, Optional: false},
			"deleted": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"status": schema.StringAttribute{Computed: true, Required: false, Optional: false},
			"name": schema.StringAttribute{
				Description: "human-readable name of the virtual machine",
				Required:    true,
			},
			"username": schema.StringAttribute{
				Description:   "name of first user that the virtual machine will generate",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"password": schema.StringAttribute{
				Description:   "password of first user that the virtual machine will generate",
				Required:      true,
				Sensitive:     true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *ResourceVirtualMachine) Configure(
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

func (r *ResourceVirtualMachine) Create(
	ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse,
) {
	var plan ResourceVirtualMachineModel
	var state ResourceVirtualMachineModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tags := map[string]string{}
	resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.PostVirtualMachine(
		plan.InstanceTypeId.ValueString(),
		plan.Name.ValueString(),
		plan.AlwaysOn.ValueBool(),
		plan.DR.ValueBool(),
		plan.Username.ValueString(),
		plan.Password.ValueString(),
		tags,
	)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to create a virtual machine", "", err)
		return
	}

	id := response.Id.String()
	tflog.Trace(ctx, fmt.Sprintf("successfully created a virtual machine: %s", id))

	getResponse, err := r.client.GetVirtualMachine(id)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to get a virtual machine", id, err)
		return
	}

	resourceVirtualMachineGetResponseToVirtualMachineModel(ctx, getResponse, &state)
	state.Password = plan.Password
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourceVirtualMachine) Read(
	ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse,
) {
	var state ResourceVirtualMachineModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	response, err := r.client.GetVirtualMachine(id)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to get a virtual machine", id, err)
		return
	}

	resp.Diagnostics.Append(
		resourceVirtualMachineGetResponseToVirtualMachineModel(ctx, response, &state)...,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourceVirtualMachine) Update(
	ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse,
) {
	var plan ResourceVirtualMachineModel
	var state ResourceVirtualMachineModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var namePtr *string = nil
	if !plan.Name.Equal(state.Name) {
		namePtr = plan.Name.ValueStringPointer()
	}

	var alwaysOnPtr *bool = nil
	if !plan.AlwaysOn.Equal(state.AlwaysOn) {
		alwaysOnPtr = plan.AlwaysOn.ValueBoolPointer()
	}

	var instanceTypeIdPtr *string = nil
	if !plan.InstanceTypeId.Equal(state.InstanceTypeId) {
		instanceTypeIdPtr = plan.InstanceTypeId.ValueStringPointer()
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

	id := state.Id.ValueString()
	_, err := r.client.PatchVirtualMachine(
		id, instanceTypeIdPtr, namePtr, alwaysOnPtr, tagsPtr,
	)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to patch a virtual machine", id, err)
		return
	}

	tflog.Info(ctx, fmt.Sprintf("successfully patched a virtual machine: %s", id))

	getResponse, err := r.client.GetVirtualMachine(state.Id.ValueString())

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to get a virtual machine", id, err)
		return
	}

	resp.Diagnostics.Append(
		resourceVirtualMachineGetResponseToVirtualMachineModel(ctx, getResponse, &state)...,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Password = plan.Password
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourceVirtualMachine) Delete(
	ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse,
) {
	var state ResourceVirtualMachineModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()

	storages, err := r.client.GetBlockStorages(&id)
	if err != nil {
		addResourceError(
			&resp.Diagnostics,
			"failed to get list of block storages attached to virtual machine",
			id,
			err,
		)
		return
	}

	var attachedMachineIdPtr *string = nil
	for _, storage := range storages {
		_, err = r.client.PatchBlockStorage(storage.Id.String(), nil, &attachedMachineIdPtr, nil)
		if err != nil {
			addResourceError(
				&resp.Diagnostics,
				"failed to detach a virtual machine from a block stroage",
				storage.Id.String(),
				err,
			)
		}
	}

	networkInterfaces, err := r.client.GetNetworkInterfaces(&id)
	if err != nil {
		addResourceError(
			&resp.Diagnostics,
			"failed to get list of network interfaces attached to virtual machine",
			id,
			err,
		)
		return
	}

	for _, networkInterface := range networkInterfaces {
		_, err = r.client.PatchNetworkInterface(
			networkInterface.Id.String(),
			nil,
			&attachedMachineIdPtr,
			nil,
		)

		if err != nil {
			addResourceError(
				&resp.Diagnostics,
				"failed to detach virtual machine from network interface",
				networkInterface.Id.String(),
				err,
			)
		}
	}

	if state.AlwaysOn.ValueBool() {
		var falsePtr = false
		_, err = r.client.PatchVirtualMachine(id, nil, nil, &falsePtr, nil)
		if err != nil {
			addResourceError(
				&resp.Diagnostics,
				"virtual machine: failed to alter always_on to false",
				id,
				err,
			)
		}
	}

	allocations, err := r.client.GetVirtualMachineAllocations(&id, nil)

	if err != nil {
		addResourceError(
			&resp.Diagnostics,
			"failed to get allocations of a virtual machine",
			id,
			err,
		)
	}

	if len(allocations) > 0 {
		tflog.Info(
			ctx,
			fmt.Sprintf(
				"virtual machine (%s) has no virtual machine allocation; skipped deleting",
				id,
			),
		)
		allocation := allocations[0]
		_, err = r.client.DeleteVirtualMachineAllocation(allocation.Id.String())
		successMessage, err := isResourceDeleted(err, "resource_allocation", "terminated")

		if err != nil {
			addResourceError(
				&resp.Diagnostics,
				"failed to terminate virtual machine allocation",
				allocation.Id.String(),
				err,
			)
			return
		}

		tflog.Info(
			ctx,
			fmt.Sprintf(
				"%s (virtual machine allocation: %s)",
				successMessage,
				allocation.Id,
			),
		)

	}

	status, diags := waitStatus(
		func() (*string, error) {
			getResponse, err := r.client.GetVirtualMachine(id)
			if err != nil {
				return nil, err
			}
			return &getResponse.Status, nil
		},
		[]string{"deleted", "idle"},
		10,
	)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if *status == "deleted" {
		tflog.Info(
			ctx,
			fmt.Sprintf("a virtual machine is already deleted (virtual machine: %s)", id),
		)
		return
	}

	_, err = r.client.DeleteVirtualMachine(id)
	successMessage, err := isResourceDeleted(err, "resource_virtual_machine", "deleted")

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to delete a virtual machine", id, err)
		return
	}

	tflog.Info(ctx, fmt.Sprintf("%s (virtual machine: %s)", successMessage, id))
}
