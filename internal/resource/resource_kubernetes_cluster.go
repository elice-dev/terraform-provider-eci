package resource

import (
	"context"
	"fmt"
	"regexp"
	"terraform-provider-eci/internal/api"
	. "terraform-provider-eci/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// About 32 minutes with waitStatus backoff, in line with eci-cli's 30 minute k8s WAIT_TIMEOUT.
const kubernetesMaxRetry = 130

var (
	regexpKubernetesName     = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	regexpKubernetesVersion  = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)
	regexpKubernetesUsername = regexp.MustCompile(`^[a-z_][a-z0-9_\-]*$`)
)

type ControlPlanePricingIdsModel struct {
	VmAllocation types.String `tfsdk:"vm_allocation"`
	BlockStorage types.String `tfsdk:"block_storage"`
	PublicIp     types.String `tfsdk:"public_ip"`
	LoadBalancer types.String `tfsdk:"load_balancer"`
}

type ControlPlaneSpecModel struct {
	ImageId        types.String                `tfsdk:"image_id"`
	StorageSizeGib types.Int64                 `tfsdk:"storage_size_gib"`
	PricingIds     ControlPlanePricingIdsModel `tfsdk:"pricing_ids"`
}

type ResourceKubernetesClusterModel struct {
	Id               types.String          `tfsdk:"id"`
	Tags             types.Map             `tfsdk:"tags"`
	Created          types.String          `tfsdk:"created"`
	Modified         types.String          `tfsdk:"modified"`
	ZoneId           types.String          `tfsdk:"zone_id"`
	OrganizationId   types.String          `tfsdk:"organization_id"`
	Name             types.String          `tfsdk:"name"`
	TargetVersion    types.String          `tfsdk:"target_version"`
	CurrentVersion   types.String          `tfsdk:"current_version"`
	AttachedSubnetId types.String          `tfsdk:"attached_subnet_id"`
	TargetReplicas   types.Int64           `tfsdk:"target_replicas"`
	CurrentReplicas  types.Int64           `tfsdk:"current_replicas"`
	ControlPlaneSpec ControlPlaneSpecModel `tfsdk:"control_plane_spec"`
	Username         types.String          `tfsdk:"username"`
	Password         types.String          `tfsdk:"password"`
	Endpoint         types.String          `tfsdk:"endpoint"`
	Ready            types.String          `tfsdk:"ready"`
	Deleted          types.String          `tfsdk:"deleted"`
	Status           types.String          `tfsdk:"status"`
}

var _ resource.Resource = &ResourceKubernetesCluster{}

type ResourceKubernetesCluster struct {
	client *api.APIClient
}

func controlPlaneSpecToAPI(spec ControlPlaneSpecModel) api.ControlPlaneSpec {
	return api.ControlPlaneSpec{
		ImageId:        spec.ImageId.ValueString(),
		StorageSizeGib: spec.StorageSizeGib.ValueInt64(),
		PricingIds: api.ControlPlanePricingIds{
			VmAllocation: spec.PricingIds.VmAllocation.ValueString(),
			BlockStorage: spec.PricingIds.BlockStorage.ValueString(),
			PublicIp:     spec.PricingIds.PublicIp.ValueString(),
			LoadBalancer: spec.PricingIds.LoadBalancer.ValueString(),
		},
	}
}

func controlPlaneSpecFromAPI(spec api.ControlPlaneSpec) ControlPlaneSpecModel {
	return ControlPlaneSpecModel{
		ImageId:        types.StringValue(spec.ImageId),
		StorageSizeGib: types.Int64Value(spec.StorageSizeGib),
		PricingIds: ControlPlanePricingIdsModel{
			VmAllocation: types.StringValue(spec.PricingIds.VmAllocation),
			BlockStorage: types.StringValue(spec.PricingIds.BlockStorage),
			PublicIp:     types.StringValue(spec.PricingIds.PublicIp),
			LoadBalancer: types.StringValue(spec.PricingIds.LoadBalancer),
		},
	}
}

// password is not in the GET response, so it is kept from the plan/state.
func resourceKubernetesClusterGetResponseToModel(
	ctx context.Context,
	response *api.ResourceKubernetesClusterGetResponse,
	data *ResourceKubernetesClusterModel,
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
	data.Name = types.StringValue(response.Name)
	data.TargetVersion = types.StringValue(response.TargetVersion)
	data.CurrentVersion = types.StringValue(response.CurrentVersion)
	data.AttachedSubnetId = types.StringValue(response.AttachedSubnetId.String())
	data.TargetReplicas = types.Int64Value(response.TargetReplicas)
	data.CurrentReplicas = types.Int64Value(response.CurrentReplicas)
	data.ControlPlaneSpec = controlPlaneSpecFromAPI(response.ControlPlaneSpec)
	data.Username = types.StringValue(response.Username)
	data.Endpoint = StringValOrNull(response.Endpoint)
	data.Ready = StringOrNull(response.Ready)
	data.Deleted = StringOrNull(response.Deleted)
	data.Status = types.StringValue(response.Status)

	return diag.Diagnostics{}
}

// replicasAndTagsPatch returns only the fields that changed between state and plan.
func replicasAndTagsPatch(
	ctx context.Context,
	planReplicas types.Int64,
	stateReplicas types.Int64,
	planTags types.Map,
	stateTags types.Map,
) (*int64, *map[string]string, diag.Diagnostics) {
	var diags diag.Diagnostics

	var replicasPtr *int64
	if !planReplicas.Equal(stateReplicas) {
		replicasPtr = planReplicas.ValueInt64Pointer()
	}

	var tagsPtr *map[string]string
	if !planTags.Equal(stateTags) {
		tags := map[string]string{}
		diags.Append(planTags.ElementsAs(ctx, &tags, false)...)
		tagsPtr = &tags
	}

	return replicasPtr, tagsPtr, diags
}

func NewResourceKubernetesCluster() resource.Resource {
	return &ResourceKubernetesCluster{}
}

func (r *ResourceKubernetesCluster) Metadata(
	ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_cluster"
}

func kubernetesNameValidators() []validator.String {
	return []validator.String{
		stringvalidator.LengthBetween(1, 32),
		stringvalidator.RegexMatches(
			regexpKubernetesName, "must consist of lower case letters, digits and '-'",
		),
	}
}

func kubernetesUsernameValidators() []validator.String {
	return []validator.String{
		stringvalidator.LengthBetween(1, 32),
		stringvalidator.RegexMatches(
			regexpKubernetesUsername,
			"must start with a lower case letter or '_' and contain only [a-z0-9_-]",
		),
	}
}

func kubernetesSpecAttributes(withLoadBalancer bool) map[string]schema.Attribute {
	pricingIds := map[string]schema.Attribute{
		"vm_allocation": schema.StringAttribute{
			Description: "id of pricing plan for the node VM allocations",
			Required:    true,
		},
		"block_storage": schema.StringAttribute{
			Description: "id of pricing plan for the node block storages",
			Required:    true,
		},
		"public_ip": schema.StringAttribute{
			Description: "id of pricing plan for the node public ips",
			Required:    true,
		},
	}
	if withLoadBalancer {
		pricingIds["load_balancer"] = schema.StringAttribute{
			Description: "id of pricing plan for the load balancer in front of the API server",
			Required:    true,
		}
	}

	return map[string]schema.Attribute{
		"image_id": schema.StringAttribute{
			Description: "id of kubernetes node image of the target version",
			Required:    true,
		},
		"storage_size_gib": schema.Int64Attribute{
			Description: "size of each node's block storage in GiB",
			Required:    true,
			Validators:  []validator.Int64{int64validator.Between(10, 32768)},
		},
		"pricing_ids": schema.SingleNestedAttribute{
			Description: "pricing plans of resources created for the nodes",
			Required:    true,
			Attributes:  pricingIds,
		},
	}
}

func (r *ResourceKubernetesCluster) Schema(
	ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Managed Kubernetes Cluster",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "unique identifier of the kubernetes cluster",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tags": schema.MapAttribute{
				Description: "User-defined metadata of key-value pairs",
				ElementType: types.StringType,
				Required:    true,
			},
			"created": schema.StringAttribute{
				Description:   "the time when the kubernetes cluster is created",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"modified": schema.StringAttribute{
				Description: "the last time when the kubernetes cluster is modified",
				Computed:    true,
			},
			"zone_id": schema.StringAttribute{
				Description:   "id of zone that the kubernetes cluster belongs to",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Description:   "id of organization that the kubernetes cluster belongs to",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description:   "name of the kubernetes cluster",
				Required:      true,
				Validators:    kubernetesNameValidators(),
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"target_version": schema.StringAttribute{
				Description: "kubernetes version of the control plane (e.g. v1.36.0). " +
					"Upgrades are not supported yet, so it cannot be changed after creation.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexpKubernetesVersion, "must look like v1.36.0"),
				},
				PlanModifiers: []planmodifier.String{
					errorOnChange("kubernetes version upgrades are not supported yet"),
				},
			},
			"current_version": schema.StringAttribute{
				Description: "kubernetes version that the control plane currently runs",
				Computed:    true,
			},
			"attached_subnet_id": schema.StringAttribute{
				Description:   "id of subnet that the cluster nodes attach to",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"target_replicas": schema.Int64Attribute{
				Description: "number of control plane nodes (1 or 3)",
				Required:    true,
				Validators:  []validator.Int64{int64validator.OneOf(1, 3)},
			},
			"current_replicas": schema.Int64Attribute{
				Description: "number of control plane nodes currently running",
				Computed:    true,
			},
			"control_plane_spec": schema.SingleNestedAttribute{
				Description:   "spec of the control plane nodes",
				Required:      true,
				Attributes:    kubernetesSpecAttributes(true),
				PlanModifiers: []planmodifier.Object{objectplanmodifier.RequiresReplace()},
			},
			"username": schema.StringAttribute{
				Description:   "login username of the control plane nodes",
				Required:      true,
				Validators:    kubernetesUsernameValidators(),
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"password": schema.StringAttribute{
				Description: "login password of the control plane nodes. " +
					"It cannot be changed after creation.",
				Required:   true,
				Sensitive:  true,
				Validators: []validator.String{stringvalidator.LengthBetween(1, 256)},
				PlanModifiers: []planmodifier.String{
					errorOnChange("the password of a kubernetes cluster cannot be changed"),
				},
			},
			"endpoint": schema.StringAttribute{
				Description: "kubernetes API server endpoint, set once the cluster is ready",
				Computed:    true,
			},
			"ready": schema.StringAttribute{
				Description: "the time when the kubernetes cluster became ready",
				Computed:    true,
			},
			"deleted": schema.StringAttribute{
				Description: "the time when the kubernetes cluster is deleted",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "status of the kubernetes cluster",
				Computed:    true,
			},
		},
	}
}

func (r *ResourceKubernetesCluster) Configure(
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

func (r *ResourceKubernetesCluster) getStatus(id string) func() (*string, error) {
	return func() (*string, error) {
		response, err := r.client.GetKubernetesCluster(id)
		if err != nil {
			return nil, err
		}
		return &response.Status, nil
	}
}

func (r *ResourceKubernetesCluster) refreshState(
	ctx context.Context, id string, data *ResourceKubernetesClusterModel, diags *diag.Diagnostics,
) {
	response, err := r.client.GetKubernetesCluster(id)
	if err != nil {
		addResourceError(diags, "failed to get a kubernetes cluster", id, err)
		return
	}

	diags.Append(resourceKubernetesClusterGetResponseToModel(ctx, response, data)...)
}

func (r *ResourceKubernetesCluster) Create(
	ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse,
) {
	var plan ResourceKubernetesClusterModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := map[string]string{}
	resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.PostKubernetesCluster(
		plan.Name.ValueString(),
		plan.TargetVersion.ValueString(),
		plan.AttachedSubnetId.ValueString(),
		plan.TargetReplicas.ValueInt64(),
		controlPlaneSpecToAPI(plan.ControlPlaneSpec),
		plan.Username.ValueString(),
		plan.Password.ValueString(),
		tags,
	)
	if err != nil {
		addResourceError(
			&resp.Diagnostics,
			"failed to create a kubernetes cluster",
			"",
			fmt.Errorf("API Error: %s", err.Error()),
		)
		return
	}

	id := response.Id.String()
	tflog.Info(ctx, fmt.Sprintf("created a kubernetes cluster: %s", id))

	state := plan
	r.refreshState(ctx, id, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save before waiting so a timed-out provisioning is still tracked in state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
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

func (r *ResourceKubernetesCluster) Read(
	ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse,
) {
	var state ResourceKubernetesClusterModel

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

func (r *ResourceKubernetesCluster) Update(
	ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse,
) {
	var plan ResourceKubernetesClusterModel
	var state ResourceKubernetesClusterModel

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

	if _, err := r.client.PatchKubernetesCluster(id, replicasPtr, tagsPtr); err != nil {
		addResourceError(&resp.Diagnostics, "failed to patch a kubernetes cluster", id, err)
		return
	}
	tflog.Info(ctx, fmt.Sprintf("successfully patched a kubernetes cluster: %s", id))

	r.refreshState(ctx, id, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourceKubernetesCluster) Delete(
	ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse,
) {
	var state ResourceKubernetesClusterModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()

	_, err := r.client.DeleteKubernetesCluster(id)
	successMessage, err := isResourceDeleted(err, "resource_kubernetes_cluster", "deleted")
	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to delete a kubernetes cluster", id, err)
		return
	}

	resp.Diagnostics.Append(waitDeleted(r.getStatus(id), kubernetesMaxRetry)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("%s (kubernetes cluster: %s)", successMessage, id))
}
