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

var _ resource.Resource = &ResourceBlockStorageSnapshot{}

func NewResourceBlockStorageSnapshot() resource.Resource {
	return &ResourceBlockStorageSnapshot{}
}

type ResourceBlockStorageSnapshot struct {
	client *api.APIClient
}

type ResourceBlockStorageSnapshotModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Tags           types.Map    `tfsdk:"tags"`
	Created        types.String `tfsdk:"created"`
	Modified       types.String `tfsdk:"modified"`
	ZoneId         types.String `tfsdk:"zone_id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	BlockStorageId types.String `tfsdk:"block_storage_id"`
	ImageId        types.String `tfsdk:"image_id"`
	SizeGib        types.Int64  `tfsdk:"size_gib"`
	Assigned       types.String `tfsdk:"assigned"`
	Prepared       types.String `tfsdk:"prepared"`
	Deleting       types.String `tfsdk:"deleting"`
	Deleted        types.String `tfsdk:"deleted"`
	DR             types.Bool   `tfsdk:"dr"`
	Status         types.String `tfsdk:"status"`
}

func resourceBlockStorageSnapshotGetResponseToBlockStorageSnapshotModel(
	ctx context.Context,
	response *api.ResourceBlockStorageSnapshotGetResponse,
	data *ResourceBlockStorageSnapshotModel,
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
	data.BlockStorageId = types.StringValue(response.BlockStorageId.String())
	data.ImageId = StringOrNull(response.ImageId)
	data.SizeGib = types.Int64Value(int64(response.SizeGib))
	data.Assigned = StringOrNull(response.Assigned)
	data.Prepared = StringOrNull(response.Prepared)
	data.Deleting = StringOrNull(response.Deleting)
	data.Deleted = StringOrNull(response.Deleted)
	data.DR = types.BoolValue(response.DR)
	data.Status = types.StringValue(string(response.Status))
	return diag.Diagnostics{}
}

func (r *ResourceBlockStorageSnapshot) Metadata(
	ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_block_storag_snapshot "
}

func (r *ResourceBlockStorageSnapshot) Schema(
	ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Block Storage Snapshot",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "unique identifier of the block storage snapshot",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tags": schema.MapAttribute{
				Description: "User-defined metadata of key-value pairs",
				ElementType: types.StringType,
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "name of the block storage snapshot",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.LengthAtMost(256),
				},
			},
			"created": schema.StringAttribute{
				Description:   "time when the block storage snapshot is created",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"modified": schema.StringAttribute{
				Description: "last time when the block storage snapshot is modified",
				Computed:    true,
			},
			"zone_id": schema.StringAttribute{
				Description:   "id of zone that the block storage snapshot belongs to",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Description:   "id of organization that the block storage snapshot belongs to",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"block_storage_id": schema.StringAttribute{
				Description: "id of the block storage this blocks storage snapshot was taken from",
				Required:    true,
			},
			"image_id": schema.StringAttribute{
				Description:   "id of the image that the block storage of this snapshot was created from",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"size_gib": schema.Int64Attribute{
				Description:   "size of the block storage snapshot (GiB)",
				Computed:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"assigned": schema.StringAttribute{
				Description:   "the time when the block storage snapshot enters `assigned` status",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"prepared": schema.StringAttribute{
				Description:   "the time when the block storage snapshot is prepared",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"deleting": schema.StringAttribute{
				Description:   "the time when the block storage snapshot enters `deleting` status",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"deleted": schema.StringAttribute{
				Description:   "the time when the block storage snapshot enters `deleted` status",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"dr": schema.BoolAttribute{
				Description:   "whether to enable DR support",
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			},
			"status": schema.StringAttribute{
				Description: "status of the block storage snapshot",
				Computed:    true,
			},
		},
	}
}

func (r *ResourceBlockStorageSnapshot) Configure(
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

func (r *ResourceBlockStorageSnapshot) Create(
	ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse,
) {
	var plan ResourceBlockStorageSnapshotModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tags := map[string]string{}
	resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.PostBlockStorageSnapshot(
		plan.Name.ValueString(),
		plan.BlockStorageId.String(),
		tags,
	)

	if err != nil {
		addResourceError(&(resp.Diagnostics), "failed to create block stroage snapshot", "", err)
		return
	}

	id := response.Id.String()

	tflog.Trace(ctx, fmt.Sprintf("created a block storage snapshot: %s", id))

	getResponse, err := r.client.GetBlockStorageSnapshot(id)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to get block storage snapshot", id, err)
		return
	}

	resourceBlockStorageSnapshotGetResponseToBlockStorageSnapshotModel(ctx, getResponse, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if getResponse.Status == "prepared" {
		tflog.Info(ctx, fmt.Sprintf("block storage snapshot (%s) is prepared", id))
		return
	}

	_, diags := waitStatus(
		func() (*string, error) {
			getResponse, err := r.client.GetBlockStorageSnapshot(id)
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

func (r *ResourceBlockStorageSnapshot) Read(
	ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse,
) {
	var data ResourceBlockStorageSnapshotModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.Id.ValueString()
	response, err := r.client.GetBlockStorageSnapshot(id)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to create block stroage snapshot", id, err)
		return
	}

	resourceBlockStorageSnapshotGetResponseToBlockStorageSnapshotModel(ctx, response, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceBlockStorageSnapshot) Update(
	ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse,
) {
	var plan ResourceBlockStorageSnapshotModel
	var state ResourceBlockStorageSnapshotModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()

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

	_, err := r.client.PatchBlockStorageSnapshot(id, namePtr, tagsPtr)

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to patch block stroage snapshot", id, err)
		return
	}

	tflog.Info(ctx, fmt.Sprintf("successfully patched a block storage snapshot: %s", id))

	getResponse, err := r.client.GetBlockStorageSnapshot(state.Id.ValueString())

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to get block storage snapshot", id, err)
		return
	}

	resourceBlockStorageSnapshotGetResponseToBlockStorageSnapshotModel(ctx, getResponse, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourceBlockStorageSnapshot) Delete(
	ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse,
) {
	var plan ResourceBlockStorageSnapshotModel

	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := plan.Id.ValueString()

	_, err := r.client.DeleteBlockStorageSnapshot(id)
	successMessage, err := isResourceDeleted(err, "resource_block_storage_snapshot", "deleted")

	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to delete a block storage snapshot", id, err)
		return
	}

	tflog.Info(ctx, fmt.Sprintf("%s (block storage snapshot: %s)", successMessage, id))
}
