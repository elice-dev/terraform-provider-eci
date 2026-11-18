package resource

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"terraform-provider-eci/internal/api"
	. "terraform-provider-eci/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// About 4 minutes with waitStatus backoff; activation is done asynchronously by the network agent.
const loadBalancerMaxRetry = 20

var regexpLeadingSlash = regexp.MustCompile(`^/`)

type LoadBalancerHealthCheckModel struct {
	Proto              types.String `tfsdk:"proto"`
	Port               types.Int64  `tfsdk:"port"`
	Interval           types.Int64  `tfsdk:"interval"`
	Timeout            types.Int64  `tfsdk:"timeout"`
	HealthyThreshold   types.Int64  `tfsdk:"healthy_threshold"`
	UnhealthyThreshold types.Int64  `tfsdk:"unhealthy_threshold"`
	Path               types.String `tfsdk:"path"`
}

type LoadBalancerRuleServerModel struct {
	Ip                 types.String `tfsdk:"ip"`
	NetworkInterfaceId types.String `tfsdk:"network_interface_id"`
	Port               types.Int64  `tfsdk:"port"`
}

type LoadBalancerRuleModel struct {
	Proto       types.String                  `tfsdk:"proto"`
	Port        types.Int64                   `tfsdk:"port"`
	Servers     []LoadBalancerRuleServerModel `tfsdk:"servers"`
	Healthcheck LoadBalancerHealthCheckModel  `tfsdk:"healthcheck"`
	Algorithm   types.String                  `tfsdk:"algorithm"`
}

type ResourceLoadBalancerModel struct {
	Id                  types.String            `tfsdk:"id"`
	Tags                types.Map               `tfsdk:"tags"`
	Created             types.String            `tfsdk:"created"`
	Modified            types.String            `tfsdk:"modified"`
	ZoneId              types.String            `tfsdk:"zone_id"`
	OrganizationId      types.String            `tfsdk:"organization_id"`
	Name                types.String            `tfsdk:"name"`
	ShortId             types.String            `tfsdk:"short_id"`
	AttachedNetworkId   types.String            `tfsdk:"attached_network_id"`
	PublicIpId          types.String            `tfsdk:"public_ip_id"`
	KubernetesClusterId types.String            `tfsdk:"kubernetes_cluster_id"`
	Rules               []LoadBalancerRuleModel `tfsdk:"rules"`
	PricingId           types.String            `tfsdk:"pricing_id"`
	PricingType         types.String            `tfsdk:"pricing_type"`
	Activated           types.String            `tfsdk:"activated"`
	Deleted             types.String            `tfsdk:"deleted"`
	Status              types.String            `tfsdk:"status"`
}

var _ resource.Resource = &ResourceLoadBalancer{}

type ResourceLoadBalancer struct {
	client *api.APIClient
}

func loadBalancerRulesToAPI(rules []LoadBalancerRuleModel) []api.LoadBalancerRule {
	result := make([]api.LoadBalancerRule, 0, len(rules))
	for _, rule := range rules {
		servers := make([]api.LoadBalancerRuleServer, 0, len(rule.Servers))
		for _, server := range rule.Servers {
			servers = append(servers, api.LoadBalancerRuleServer{
				Ip:                 server.Ip.ValueString(),
				NetworkInterfaceId: server.NetworkInterfaceId.ValueString(),
				Port:               server.Port.ValueInt64(),
			})
		}

		hc := rule.Healthcheck
		result = append(result, api.LoadBalancerRule{
			Proto:   rule.Proto.ValueString(),
			Port:    rule.Port.ValueInt64(),
			Servers: servers,
			Healthcheck: api.LoadBalancerHealthCheck{
				Proto:              hc.Proto.ValueString(),
				Port:               hc.Port.ValueInt64(),
				Interval:           hc.Interval.ValueInt64(),
				Timeout:            hc.Timeout.ValueInt64(),
				HealthyThreshold:   hc.HealthyThreshold.ValueInt64(),
				UnhealthyThreshold: hc.UnhealthyThreshold.ValueInt64(),
				Path:               hc.Path.ValueStringPointer(),
			},
			Algorithm: rule.Algorithm.ValueString(),
		})
	}
	return result
}

func loadBalancerRulesFromAPI(rules []api.LoadBalancerRule) []LoadBalancerRuleModel {
	result := make([]LoadBalancerRuleModel, 0, len(rules))
	for _, rule := range rules {
		servers := make([]LoadBalancerRuleServerModel, 0, len(rule.Servers))
		for _, server := range rule.Servers {
			servers = append(servers, LoadBalancerRuleServerModel{
				Ip:                 types.StringValue(server.Ip),
				NetworkInterfaceId: types.StringValue(server.NetworkInterfaceId),
				Port:               types.Int64Value(server.Port),
			})
		}

		hc := rule.Healthcheck
		result = append(result, LoadBalancerRuleModel{
			Proto:   types.StringValue(rule.Proto),
			Port:    types.Int64Value(rule.Port),
			Servers: servers,
			Healthcheck: LoadBalancerHealthCheckModel{
				Proto:              types.StringValue(hc.Proto),
				Port:               types.Int64Value(hc.Port),
				Interval:           types.Int64Value(hc.Interval),
				Timeout:            types.Int64Value(hc.Timeout),
				HealthyThreshold:   types.Int64Value(hc.HealthyThreshold),
				UnhealthyThreshold: types.Int64Value(hc.UnhealthyThreshold),
				Path:               StringValOrNull(hc.Path),
			},
			Algorithm: types.StringValue(rule.Algorithm),
		})
	}
	return result
}

// public_ip_id is not in the GET response, so it is kept from the plan/state.
func resourceLoadBalancerGetResponseToModel(
	ctx context.Context,
	response *api.ResourceLoadBalancerGetResponse,
	data *ResourceLoadBalancerModel,
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
	data.ShortId = types.StringValue(response.ShortId)
	data.AttachedNetworkId = types.StringValue(response.AttachedNetworkId.String())
	data.KubernetesClusterId = StringOrNull(response.KubernetesClusterId)
	data.Rules = loadBalancerRulesFromAPI(response.Rules)
	data.PricingId = types.StringValue(response.PricingId.String())
	data.PricingType = types.StringValue(response.PricingType)
	data.Activated = StringOrNull(response.Activated)
	data.Deleted = StringOrNull(response.Deleted)
	data.Status = types.StringValue(response.Status)

	return diag.Diagnostics{}
}

func NewResourceLoadBalancer() resource.Resource {
	return &ResourceLoadBalancer{}
}

func (r *ResourceLoadBalancer) Metadata(
	ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer"
}

func portValidators() []validator.Int64 {
	return []validator.Int64{int64validator.Between(1, 65535)}
}

func (r *ResourceLoadBalancer) Schema(
	ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse,
) {
	healthcheck := schema.SingleNestedAttribute{
		Description: "health check of the servers in the rule",
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"proto": schema.StringAttribute{
				Description: "health check protocol (TCP or HTTP)",
				Required:    true,
				Validators:  []validator.String{stringvalidator.OneOf("TCP", "HTTP")},
			},
			"port": schema.Int64Attribute{
				Description: "health check port",
				Required:    true,
				Validators:  portValidators(),
			},
			"interval": schema.Int64Attribute{
				Description: "seconds between health checks",
				Required:    true,
				Validators:  []validator.Int64{int64validator.AtLeast(1)},
			},
			"timeout": schema.Int64Attribute{
				Description: "seconds before a health check times out",
				Required:    true,
				Validators:  []validator.Int64{int64validator.AtLeast(1)},
			},
			"healthy_threshold": schema.Int64Attribute{
				Description: "consecutive successes to mark a server healthy",
				Required:    true,
				Validators:  []validator.Int64{int64validator.AtLeast(1)},
			},
			"unhealthy_threshold": schema.Int64Attribute{
				Description: "consecutive failures to mark a server unhealthy",
				Required:    true,
				Validators:  []validator.Int64{int64validator.AtLeast(1)},
			},
			"path": schema.StringAttribute{
				Description: "HTTP health check path, starting with /",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(256),
					stringvalidator.RegexMatches(regexpLeadingSlash, "must start with /"),
				},
			},
		},
	}

	server := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"ip": schema.StringAttribute{
				Description: "private ip of the backend server",
				Required:    true,
			},
			"network_interface_id": schema.StringAttribute{
				Description: "id of network interface of the backend server",
				Required:    true,
			},
			"port": schema.Int64Attribute{
				Description: "port of the backend server",
				Required:    true,
				Validators:  portValidators(),
			},
		},
	}

	rule := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"proto": schema.StringAttribute{
				Description: "protocol of the rule (TCP or UDP)",
				Required:    true,
				Validators:  []validator.String{stringvalidator.OneOf("TCP", "UDP")},
			},
			"port": schema.Int64Attribute{
				Description: "port that the load balancer listens on",
				Required:    true,
				Validators:  portValidators(),
			},
			"servers": schema.ListNestedAttribute{
				Description:  "backend servers of the rule",
				Required:     true,
				NestedObject: server,
				Validators:   []validator.List{listvalidator.SizeAtMost(256)},
			},
			"healthcheck": healthcheck,
			"algorithm": schema.StringAttribute{
				Description: "balancing algorithm (roundrobin, random or leastconn)",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("roundrobin", "random", "leastconn"),
				},
			},
		},
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Load Balancer",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "unique identifier of the load balancer",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tags": schema.MapAttribute{
				Description: "User-defined metadata of key-value pairs",
				ElementType: types.StringType,
				Required:    true,
			},
			"created": schema.StringAttribute{
				Description:   "the time when the load balancer is created",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"modified": schema.StringAttribute{
				Description: "the last time when the load balancer is modified",
				Computed:    true,
			},
			"zone_id": schema.StringAttribute{
				Description:   "id of zone that the load balancer belongs to",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Description:   "id of organization that the load balancer belongs to",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description: "name of the load balancer",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthBetween(1, 256)},
			},
			"short_id": schema.StringAttribute{
				Description:   "short identifier of the load balancer",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"attached_network_id": schema.StringAttribute{
				Description:   "id of virtual network that the load balancer attaches to",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"public_ip_id": schema.StringAttribute{
				Description:   "id of public ip assigned to the load balancer",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"kubernetes_cluster_id": schema.StringAttribute{
				Description: "id of kubernetes cluster that owns the load balancer, if any",
				Computed:    true,
			},
			"rules": schema.ListNestedAttribute{
				Description:  "forwarding rules of the load balancer",
				Required:     true,
				NestedObject: rule,
				Validators:   []validator.List{listvalidator.SizeBetween(1, 128)},
			},
			"pricing_id": schema.StringAttribute{
				Description:   "id of pricing plan for the load balancer",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"pricing_type": schema.StringAttribute{
				Description: "type of pricing plan (computed)",
				Computed:    true,
			},
			"activated": schema.StringAttribute{
				Description: "the time when the load balancer is activated",
				Computed:    true,
			},
			"deleted": schema.StringAttribute{
				Description: "the time when the load balancer is deleted",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "status of the load balancer",
				Computed:    true,
			},
		},
	}
}

func (r *ResourceLoadBalancer) Configure(
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

func (r *ResourceLoadBalancer) getStatus(id string) func() (*string, error) {
	return func() (*string, error) {
		response, err := r.client.GetLoadBalancer(id)
		if err != nil {
			return nil, err
		}
		return &response.Status, nil
	}
}

// refreshState reads the load balancer and stores it into state, keeping public_ip_id.
func (r *ResourceLoadBalancer) refreshState(
	ctx context.Context, id string, data *ResourceLoadBalancerModel, diags *diag.Diagnostics,
) {
	response, err := r.client.GetLoadBalancer(id)
	if err != nil {
		addResourceError(diags, "failed to get a load balancer", id, err)
		return
	}

	diags.Append(resourceLoadBalancerGetResponseToModel(ctx, response, data)...)
}

func (r *ResourceLoadBalancer) Create(
	ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse,
) {
	var plan ResourceLoadBalancerModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := map[string]string{}
	resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.PostLoadBalancer(
		plan.Name.ValueString(),
		plan.AttachedNetworkId.ValueString(),
		plan.PublicIpId.ValueString(),
		loadBalancerRulesToAPI(plan.Rules),
		plan.PricingId.ValueString(),
		tags,
	)
	if err != nil {
		addResourceError(
			&resp.Diagnostics,
			"failed to create a load balancer",
			"",
			fmt.Errorf("API Error: %s", err.Error()),
		)
		return
	}

	id := response.Id.String()
	tflog.Info(ctx, fmt.Sprintf("created a load balancer: %s", id))

	// Track the id first so a failure below leaves a tainted resource, not an untracked one.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)

	state := plan
	r.refreshState(ctx, id, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save before waiting so a timed-out activation is still tracked in state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(waitUntil(r.getStatus(id), "activated", loadBalancerMaxRetry)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.refreshState(ctx, id, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourceLoadBalancer) Read(
	ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse,
) {
	var state ResourceLoadBalancerModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	response, err := r.client.GetLoadBalancer(id)
	// Deleted outside Terraform: drop it so the next plan creates it again.
	if isNotFound(err) || (err == nil && response.Status == "deleted") {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to get a load balancer", id, err)
		return
	}

	resp.Diagnostics.Append(resourceLoadBalancerGetResponseToModel(ctx, response, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourceLoadBalancer) Update(
	ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse,
) {
	var plan ResourceLoadBalancerModel
	var state ResourceLoadBalancerModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	namePtr, rulesPtr, tagsPtr, diags := loadBalancerPatch(ctx, &plan, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.PatchLoadBalancer(id, namePtr, rulesPtr, tagsPtr); err != nil {
		addResourceError(&resp.Diagnostics, "failed to patch a load balancer", id, err)
		return
	}
	tflog.Info(ctx, fmt.Sprintf("successfully patched a load balancer: %s", id))

	r.refreshState(ctx, id, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// loadBalancerPatch returns only the fields that changed between state and plan.
func loadBalancerPatch(
	ctx context.Context, plan *ResourceLoadBalancerModel, state *ResourceLoadBalancerModel,
) (*string, *[]api.LoadBalancerRule, *map[string]string, diag.Diagnostics) {
	var diags diag.Diagnostics

	var namePtr *string
	if !plan.Name.Equal(state.Name) {
		namePtr = plan.Name.ValueStringPointer()
	}

	var rulesPtr *[]api.LoadBalancerRule
	planRules := loadBalancerRulesToAPI(plan.Rules)
	if !reflect.DeepEqual(planRules, loadBalancerRulesToAPI(state.Rules)) {
		rulesPtr = &planRules
	}

	var tagsPtr *map[string]string
	if !plan.Tags.Equal(state.Tags) {
		tags := map[string]string{}
		diags.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
		tagsPtr = &tags
	}

	return namePtr, rulesPtr, tagsPtr, diags
}

// loadBalancerDeleteResult also accepts deleting: eci-api rejects DELETE unless the load balancer
// is activated, so a destroy retried after a timed-out wait would otherwise fail until it is gone.
func loadBalancerDeleteResult(err error) (string, error) {
	successMessage, deletedErr := isResourceDeleted(err, "resource_load_balancer", "deleted")
	if deletedErr == nil {
		return successMessage, nil
	}
	if _, deletingErr := isResourceDeleted(
		err,
		"resource_load_balancer",
		"deleting",
	); deletingErr == nil {
		return "deletion is already in progress", nil
	}
	return "", deletedErr
}

func (r *ResourceLoadBalancer) Delete(
	ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse,
) {
	var state ResourceLoadBalancerModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()

	_, err := r.client.DeleteLoadBalancer(id)
	successMessage, err := loadBalancerDeleteResult(err)
	if err != nil {
		addResourceError(&resp.Diagnostics, "failed to delete a load balancer", id, err)
		return
	}

	// Wait so the attached public ip is released before Terraform deletes it.
	resp.Diagnostics.Append(waitDeleted(r.getStatus(id), loadBalancerMaxRetry)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("%s (load balancer: %s)", successMessage, id))
}
