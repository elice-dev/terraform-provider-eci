package resource

import (
	"context"
	"fmt"
	"terraform-provider-eci/internal/api"
	. "terraform-provider-eci/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &ResourceBlockStorage{}

func NewResourceBlockStorage() resource.Resource {
	return &ResourceBlockStorage{}
}

type ResourceBlockStorage struct {
	client *api.APIClient
}

type ResourceBlockStorageModel struct {
	Id                 types.String `tfsdk:"id"`
	Tags               types.Map    `tfsdk:"tags"`
	Name               types.String `tfsdk:"name"`
	Created            types.String `tfsdk:"created"`
	Modified           types.String `tfsdk:"modified"`
	ZoneId             types.String `tfsdk:"zone_id"`
	OrganizationId     types.String `tfsdk:"organization_id"`
	AttachedMachineId  types.String `tfsdk:"attached_machine_id"`
	ImageId            types.String `tfsdk:"image_id"`
	SnapshotId         types.String `tfsdk:"snapshot_id"`
	SizeGib            types.Int64  `tfsdk:"size_gib"`
	DR                 types.Bool   `tfsdk:"dr"`
	LastSyncedSnapshot types.String `tfsdk:"last_synced_snapshot"`
	Assigned           types.String `tfsdk:"assigned"`
	Prepared           types.String `tfsdk:"prepared"`
	Deleting           types.String `tfsdk:"deleting"`
	Deleted            types.String `tfsdk:"deleted"`
	Status             types.String `tfsdk:"status"`
}

func resourceBlockStorageGetResponseToBlockStorageModel(
	ctx context.Context,
	response *api.ResourceBlockStorageGetResponse,
	data *ResourceBlockStorageModel,
) diag.Diagnostics {
	data.Id = types.StringValue(response.Id.String())
	data.Name = types.StringValue(response.Name)
	data.Created = types.StringValue(response.Created.String())

	tags, diags := types.MapValueFrom(ctx, types.StringType, response.Tags)

	if diags.HasError() {
		return diags
	}

	data.Tags = tags
	data.Modified = StringOrNull(response.Modified)
	data.ZoneId = types.StringValue(response.ZoneId.String())
	data.OrganizationId = types.StringValue(response.OrganizationId.String())
	data.AttachedMachineId = StringOrNull(response.AttachedMachineId)
	data.ImageId = StringOrNull(response.ImageId)
	data.SnapshotId = StringOrNull(response.SnapshotId)
	data.SizeGib = types.Int64Value(int64(response.SizeGib))
	data.DR = types.BoolValue(response.DR)
	data.LastSyncedSnapshot = StringValOrNull(response.LastSyncedSnapshot)
	data.Assigned = StringOrNull(response.Assigned)
	data.Prepared = StringOrNull(response.Prepared)
	data.Deleting = StringOrNull(response.Deleting)
	data.Deleted = StringOrNull(response.Deleted)
	data.Status = types.StringValue(string(response.Status))
	return diag.Diagnostics{}
}

func (r *ResourceBlockStorage) Metadata(
	ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_block_storage"
}

func (r *ResourceBlockStorage) Schema(
	ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Block Storage",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "unique identifier of the block storage",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tags": schema.MapAttribute{
				Description: "User-defined metadata of key-value pairs",
				ElementType: types.StringType,
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "name of the block storage",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.LengthAtMost(255),
				},
			},
			"created": schema.StringAttribute{
				Description:   "time when the block storage is created",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"modified": schema.StringAttribute{
				Description: "last time when the block storage is modified",
				Computed:    true,
			},
			"zone_id": schema.StringAttribute{
				Description:   "id of zone that the block storage belongs to",
				Required:      false,
				Computed:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Description:   "id of organization that the block storage belongs to",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"attached_machine_id": schema.StringAttribute{
				Description: "the id of the virtual machine this blocks storage will attach to",
				Required:    true,
			},
			"image_id": schema.StringAttribute{
				Description:   "id of image that the block storage will copy from",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"snapshot_id": schema.StringAttribute{
				Description:   "id of snapshot that the block storage will copy from",
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"size_gib": schema.Int64Attribute{
				Description:   "size of the block storage (GiB)",
				Required:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"dr": schema.BoolAttribute{
				Description:   "whether to enable DR support",
				Required:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			},
			"last_synced_snapshot": schema.StringAttribute{
				Description: "the last time when the block storage is synced with the DR zone",
				Computed:    true,
			},
			"assigned": schema.StringAttribute{
				Description:   "the time when the block storage enters `assigned` status",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"prepared": schema.StringAttribute{
				Description:   "the time when the block storage is prepared",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"deleting": schema.StringAttribute{
				Description:   "the time when the block storage enters `deleting` status",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"deleted": schema.StringAttribute{
				Description:   "the time when the block storage enters `deleted` status",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"status": schema.StringAttribute{
				Description: "status of the block storage",
				Computed:    true,
			},
		},
	}
}

func (r *ResourceBlockStorage) Configure(
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

func (r *ResourceBlockStorage) Create(
	ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse,
) {
	var plan ResourceBlockStorageModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tags := map[string]string{}
	resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var imageIdPtr *string = nil
	if !plan.ImageId.IsUnknown() {
		imageIdPtr = plan.ImageId.ValueStringPointer()
	}

	response, err := r.client.PostBlockStorage(
		plan.Name.ValueString(),
		imageIdPtr,
		plan.SnapshotId.ValueStringPointer(),
		int(plan.SizeGib.ValueInt64()),
		plan.DR.ValueBool(),
		tags,
	)

	if err != nil {
		addResourceError(&(resp.Diagnostics), "failed to create block stroage", "", err)
		return
	}

	id := response.Id.String()

	tflog.Trace(ctx, fmt.Sprintf("created a block storage: %s", id))

	if !plan.AttachedMachineId.IsNull() {
		var attachedMachineId = plan.AttachedMachineId.ValueStringPointer()
		_, err := r.client.PatchBlockStorage(id, nil, &attachedMachineId, nil)

		if err != nil {
			addResourceError(&resp.Diagnostics, "failed to patch block storage", id, err)
			return
		}

		tflog.Trace(
			ctx,
			fmt.Sprintf("block storage (%s) patched: attached to a virtual machine", id),
		)
	}

	getResponse, err := r.client.GetBlockStorage(id)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to get block storage", id, err)
		return
	}

	resourceBlockStorageGetResponseToBlockStorageModel(ctx, getResponse, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if getResponse.Status == "prepared" {
		tflog.Info(ctx, fmt.Sprintf("block storage (%s) is prepared", id))
		return
	}

	_, diags := waitStatus(
		func() (*string, error) {
			getResponse, err := r.client.GetBlockStorage(id)
			if err != nil {
				return nil, err
			}
			return &getResponse.Status, nil
		},
		[]string{"prepared"},
		10,
	)
	resp.Diagnostics.Append(diags...)
}

func (r *ResourceBlockStorage) Read(
	ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse,
) {
	var data ResourceBlockStorageModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.Id.ValueString()
	response, err := r.client.GetBlockStorage(id)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to create block stroage", id, err)
		return
	}

	resourceBlockStorageGetResponseToBlockStorageModel(ctx, response, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceBlockStorage) Update(
	ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse,
) {
	var plan ResourceBlockStorageModel
	var state ResourceBlockStorageModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()

	var attachedMachineIdPtr **string = nil
	if !state.AttachedMachineId.Equal(plan.AttachedMachineId) {
		if !state.AttachedMachineId.IsNull() && !plan.AttachedMachineId.IsNull() {
			var nilAttachedMachineId *string = nil

			_, err := r.client.PatchBlockStorage(id, nil, &nilAttachedMachineId, nil)
			if err != nil {
				addResourceError(&resp.Diagnostics, "failed to detach block stroage", id, err)
				return
			}

			tflog.Info(
				ctx,
				fmt.Sprintf("successfully detached from a virtual machine a block storage: %s", id),
			)
		}

		var attachedMachineId = plan.AttachedMachineId.ValueStringPointer()

		attachedMachineIdPtr = &attachedMachineId
	}

	var namePtr *string = nil
	if !state.Name.Equal(plan.Name) {
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

	_, err := r.client.PatchBlockStorage(id, namePtr, attachedMachineIdPtr, tagsPtr)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to patch block stroage", id, err)
		return
	}

	tflog.Info(ctx, fmt.Sprintf("successfully patched a block storage: %s", id))

	getResponse, err := r.client.GetBlockStorage(state.Id.ValueString())

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to get block storage", id, err)
		return
	}

	resourceBlockStorageGetResponseToBlockStorageModel(ctx, getResponse, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourceBlockStorage) Delete(
	ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse,
) {
	var plan ResourceBlockStorageModel

	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := plan.Id.ValueString()

	if !plan.AttachedMachineId.IsNull() {
		virtualMachine, err := r.client.GetVirtualMachine(plan.AttachedMachineId.ValueString())
		if err != nil {
			addResourceError(
				&resp.Diagnostics,
				"failed to get virtual machine",
				id,
				err,
			)
			return
		}

		if virtualMachine.Status != "idle" {
			resp.Diagnostics.AddError(
				"Invalid virtual machien status",
				"block storage is attached to a non-idle virtual machine. For safety, the practitioner has to kill the virtual machine allocation",
			)
			return
		}

		var nilAttachedMachineId *string = nil
		_, err = r.client.PatchBlockStorage(id, nil, &nilAttachedMachineId, nil)
		if err != nil {
			addResourceError(
				&resp.Diagnostics,
				"failed to patch block storage (while detaching from machine)",
				id,
				err,
			)
			return
		}

		tflog.Trace(ctx, fmt.Sprintf("block storage (%s) detached from a virtual machine", id))
	}

	_, err := r.client.DeleteBlockStorage(id)
	successMessage, err := isResourceDeleted(err, "resource_block_storage", "deleted")

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to delete a block storage", id, err)
		return
	}

	tflog.Info(ctx, fmt.Sprintf("%s (block storage: %s)", successMessage, id))
}
