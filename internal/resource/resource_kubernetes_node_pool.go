package resource

import (
	"context"
	"fmt"
	"terraform-provider-eci/internal/api"
	. "terraform-provider-eci/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type NodePoolPricingIdsModel struct {
	VmAllocation types.String `tfsdk:"vm_allocation"`
	BlockStorage types.String `tfsdk:"block_storage"`
	PublicIp     types.String `tfsdk:"public_ip"`
}

type NodePoolSpecModel struct {
	ImageId        types.String            `tfsdk:"image_id"`
	StorageSizeGib types.Int64             `tfsdk:"storage_size_gib"`
	PricingIds     NodePoolPricingIdsModel `tfsdk:"pricing_ids"`
}

type ResourceKubernetesNodePoolModel struct {
	Id              types.String      `tfsdk:"id"`
	Tags            types.Map         `tfsdk:"tags"`
	Created         types.String      `tfsdk:"created"`
	Modified        types.String      `tfsdk:"modified"`
	ZoneId          types.String      `tfsdk:"zone_id"`
	OrganizationId  types.String      `tfsdk:"organization_id"`
	ClusterId       types.String      `tfsdk:"cluster_id"`
	Name            types.String      `tfsdk:"name"`
	TargetVersion   types.String      `tfsdk:"target_version"`
	CurrentVersion  types.String      `tfsdk:"current_version"`
	TargetReplicas  types.Int64       `tfsdk:"target_replicas"`
	CurrentReplicas types.Int64       `tfsdk:"current_replicas"`
	NodePoolSpec    NodePoolSpecModel `tfsdk:"node_pool_spec"`
	Username        types.String      `tfsdk:"username"`
	Password        types.String      `tfsdk:"password"`
	Ready           types.String      `tfsdk:"ready"`
	Deleted         types.String      `tfsdk:"deleted"`
	Status          types.String      `tfsdk:"status"`
}

var _ resource.Resource = &ResourceKubernetesNodePool{}

type ResourceKubernetesNodePool struct {
	client *api.APIClient
}

func nodePoolSpecToAPI(spec NodePoolSpecModel) api.NodePoolSpec {
	return api.NodePoolSpec{
		ImageId:        spec.ImageId.ValueString(),
		StorageSizeGib: spec.StorageSizeGib.ValueInt64(),
		PricingIds: api.NodePoolPricingIds{
			VmAllocation: spec.PricingIds.VmAllocation.ValueString(),
			BlockStorage: spec.PricingIds.BlockStorage.ValueString(),
			PublicIp:     spec.PricingIds.PublicIp.ValueString(),
		},
	}
}

func nodePoolSpecFromAPI(spec api.NodePoolSpec) NodePoolSpecModel {
	return NodePoolSpecModel{
		ImageId:        types.StringValue(spec.ImageId),
		StorageSizeGib: types.Int64Value(spec.StorageSizeGib),
		PricingIds: NodePoolPricingIdsModel{
			VmAllocation: types.StringValue(spec.PricingIds.VmAllocation),
			BlockStorage: types.StringValue(spec.PricingIds.BlockStorage),
			PublicIp:     types.StringValue(spec.PricingIds.PublicIp),
		},
	}
}

// password is not in the GET response, so it is kept from the plan/state.
func resourceKubernetesNodePoolGetResponseToModel(
	ctx context.Context,
	response *api.ResourceKubernetesNodePoolGetResponse,
	data *ResourceKubernetesNodePoolModel,
) diag.Diagnostics {
	tags, diags := types.MapValueFrom(ctx, types.StringType, response.Tags)
	if diags.HasError() {
		return diags
	}

	data.Id = types.StringValue(response.Id.String())
	data.Tags = tags
	data.Created = types.StringValue(response.Created.String())
	data.Modified = StringOrNull(response.Modified)
	data.ZoneId = types.StringValue(response.ZoneId.String())
	data.OrganizationId = types.StringValue(response.OrganizationId.String())
	data.ClusterId = types.StringValue(response.ClusterId.String())
	data.Name = types.StringValue(response.Name)
	data.TargetVersion = types.StringValue(response.TargetVersion)
	data.CurrentVersion = types.StringValue(response.CurrentVersion)
	data.TargetReplicas = types.Int64Value(response.TargetReplicas)
	data.CurrentReplicas = types.Int64Value(response.CurrentReplicas)
	data.NodePoolSpec = nodePoolSpecFromAPI(response.NodePoolSpec)
	data.Username = types.StringValue(response.Username)
	data.Ready = StringOrNull(response.Ready)
	data.Deleted = StringOrNull(response.Deleted)
	data.Status = types.StringValue(response.Status)

	return diag.Diagnostics{}
}

func NewResourceKubernetesNodePool() resource.Resource {
	return &ResourceKubernetesNodePool{}
}

func (r *ResourceKubernetesNodePool) Metadata(
	ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_node_pool"
}

func (r *ResourceKubernetesNodePool) Schema(
	ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Managed Kubernetes Node Pool",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "unique identifier of the node pool",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tags": schema.MapAttribute{
				Description: "User-defined metadata of key-value pairs",
				ElementType: types.StringType,
				Required:    true,
			},
			"created": schema.StringAttribute{
				Description:   "the time when the node pool is created",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"modified": schema.StringAttribute{
				Description: "the last time when the node pool is modified",
				Computed:    true,
			},
			"zone_id": schema.StringAttribute{
				Description:   "id of zone that the node pool belongs to",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Description:   "id of organization that the node pool belongs to",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"cluster_id": schema.StringAttribute{
				Description:   "id of kubernetes cluster that the node pool joins",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Description:   "name of the node pool",
				Required:      true,
				Validators:    kubernetesNameValidators(),
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"target_version": schema.StringAttribute{
				Description:   "kubernetes version of the nodes, following the cluster",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"current_version": schema.StringAttribute{
				Description: "kubernetes version that the nodes currently run",
				Computed:    true,
			},
			"target_replicas": schema.Int64Attribute{
				Description: "number of nodes (1 to 64)",
				Required:    true,
				Validators:  []validator.Int64{int64validator.Between(1, 64)},
			},
			"current_replicas": schema.Int64Attribute{
				Description:   "number of nodes currently running",
				Computed:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"node_pool_spec": schema.SingleNestedAttribute{
				Description:   "spec of the nodes",
				Required:      true,
				Attributes:    kubernetesSpecAttributes(false),
				PlanModifiers: []planmodifier.Object{objectplanmodifier.RequiresReplace()},
			},
			"username": schema.StringAttribute{
				Description:   "login username of the nodes",
				Required:      true,
				Validators:    kubernetesUsernameValidators(),
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"password": schema.StringAttribute{
				Description: "login password of the nodes. It cannot be changed after creation.",
				Required:    true,
				Sensitive:   true,
				Validators:  []validator.String{stringvalidator.LengthBetween(1, 256)},
				PlanModifiers: []planmodifier.String{
					errorOnChange("the password of a node pool cannot be changed"),
				},
			},
			"ready": schema.StringAttribute{
				Description: "the time when the node pool became ready",
				Computed:    true,
			},
			"deleted": schema.StringAttribute{
				Description: "the time when the node pool is deleted",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "status of the node pool",
				Computed:    true,
			},
		},
	}
}

func (r *ResourceKubernetesNodePool) Configure(
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

func (r *ResourceKubernetesNodePool) getStatus(id string) func() (*string, error) {
	return func() (*string, error) {
		response, err := r.client.GetKubernetesNodePool(id)
		if err != nil {
			return nil, err
		}
		return &response.Status, nil
	}
}

func (r *ResourceKubernetesNodePool) refreshState(
	ctx context.Context, id string, data *ResourceKubernetesNodePoolModel, diags *diag.Diagnostics,
) *api.ResourceKubernetesNodePoolGetResponse {
	response, err := r.client.GetKubernetesNodePool(id)
	if err != nil {
		addResourceError(diags, "failed to get a node pool", id, err)
		return nil
	}

	diags.Append(resourceKubernetesNodePoolGetResponseToModel(ctx, response, data)...)
	return response
}

func (r *ResourceKubernetesNodePool) Create(
	ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse,
) {
	var plan ResourceKubernetesNodePoolModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := map[string]string{}
	resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.PostKubernetesNodePool(
		plan.ClusterId.ValueString(),
		plan.Name.ValueString(),
		plan.TargetReplicas.ValueInt64(),
		nodePoolSpecToAPI(plan.NodePoolSpec),
		plan.Username.ValueString(),
		plan.Password.ValueString(),
		tags,
	)
	if err != nil {
		addResourceError(
			&resp.Diagnostics,
			"failed to create a node pool",
			"",
			fmt.Errorf("API Error: %s", err.Error()),
		)
		return
	}

	id := response.Id.String()
	tflog.Info(ctx, fmt.Sprintf("created a node pool: %s", id))

	state := plan
	getResponse := r.refreshState(ctx, id, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save before waiting so a timed-out provisioning is still tracked in state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() || getResponse.Status == "ready" {
		return
	}

	resp.Diagnostics.Append(waitUntil(r.getStatus(id), "ready", kubernetesMaxRetry)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.refreshState(ctx, id, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourceKubernetesNodePool) Read(
	ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse,
) {
	var state ResourceKubernetesNodePoolModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.refreshState(ctx, state.Id.ValueString(), &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourceKubernetesNodePool) Update(
	ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse,
) {
	var plan ResourceKubernetesNodePoolModel
	var state ResourceKubernetesNodePoolModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	replicasPtr, tagsPtr, diags := replicasAndTagsPatch(
		ctx, plan.TargetReplicas, state.TargetReplicas, plan.Tags, state.Tags,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.PatchKubernetesNodePool(id, replicasPtr, tagsPtr); err != nil {
		addResourceError(&resp.Diagnostics, "failed to patch a node pool", id, err)
		return
	}
	tflog.Info(ctx, fmt.Sprintf("successfully patched a node pool: %s", id))

	r.refreshState(ctx, id, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourceKubernetesNodePool) Delete(
	ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse,
) {
	var state ResourceKubernetesNodePoolModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()

	_, err := r.client.DeleteKubernetesNodePool(id)
	successMessage, err := isResourceDeleted(err, "resource_kubernetes_node_pool", "deleted")
	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to delete a node pool", id, err)
		return
	}

	// The cluster refuses deletion while any of its node pools is not deleted yet.
	resp.Diagnostics.Append(waitDeleted(r.getStatus(id), kubernetesMaxRetry)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("%s (node pool: %s)", successMessage, id))
}
