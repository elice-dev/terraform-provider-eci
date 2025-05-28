package datasource

import (
	"context"
	"fmt"
	"terraform-provider-eci/internal/api"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &InstanceTypeDataSource{}
	_ datasource.DataSourceWithConfigure = &InstanceTypeDataSource{}
)

func NewInstanceTypeDataSource() datasource.DataSource {
	return &InstanceTypeDataSource{}
}

type InstanceTypeDataSource struct {
	client *api.APIClient
}

type InstanceTypeDataSourceModel struct {
	Id           types.String `tfsdk:"id"`
	Created      types.String `tfsdk:"created"`
	ZoneId       types.String `tfsdk:"zone_id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	CpuVcore     types.Int64  `tfsdk:"cpu_vcore"`
	MemoryGib    types.Int64  `tfsdk:"memory_gib"`
	Devices      types.List   `tfsdk:"devices"`
	PricePerHour types.String `tfsdk:"price_per_hour"`
	Activated    types.Bool   `tfsdk:"activated"`
}

func (d *InstanceTypeDataSource) Configure(
	_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.APIClient)

	if !ok {
		resp.Diagnostics.AddError(
			"unexpected resource configure type",
			fmt.Sprintf(
				`expected *api.APIClient, got: %T. 
				please report this issue to the provider developers.`,
				req.ProviderData,
			),
		)

		return
	}

	d.client = client
}

func (d *InstanceTypeDataSource) Metadata(
	_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_instance_type"
}

func (d *InstanceTypeDataSource) Schema(
	_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Instance Type",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "unique identifier of the instance type",
				Computed:    true,
				Optional:    false,
				Required:    false,
			},
			"created": schema.StringAttribute{
				Description: "the time when the instance type is created",
				Computed:    true,
				Optional:    false,
				Required:    false,
			},
			"zone_id": schema.StringAttribute{
				Description: "id of zone that the instance type belongs to",
				Required:    false,
				Computed:    true,
				Optional:    false,
			},
			"name": schema.StringAttribute{
				Description: "human-readable name of the instance type",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "description of the instance type",
				Computed:    true,
				Optional:    false,
				Required:    false,
			},
			"cpu_vcore": schema.Int64Attribute{
				Description: "number of CPU vCores that a virtual machine will acquire",
				Computed:    true,
				Optional:    false,
				Required:    false,
			},
			"memory_gib": schema.Int64Attribute{
				Description: "size of memory (GiB) that a virtual machine will acquire",
				Computed:    true,
			},
			"devices": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "list of devices that a virtual machine will acquire",
				Computed:    true,
			},
			"price_per_hour": schema.StringAttribute{
				Description: "price per hour of this instance type",
				Computed:    true,
			},
			"activated": schema.BoolAttribute{
				Description: "whether this instance type is activated;" +
					"not, it cannot be used to create a virtual machine",
				Computed: true,
			},
		},
	}
}

func InstanceTypeGetResponseToInstanceTypeModel(
	ctx context.Context,
	response *api.InfraInstanceTypeGetResponse,
	data *InstanceTypeDataSourceModel,
) diag.Diagnostics {
	data.Id = types.StringValue(response.Id.String())
	data.Created = types.StringValue(response.Created.String())
	data.ZoneId = types.StringValue(response.ZoneId.String())
	data.Name = types.StringValue(response.Name)
	data.Description = types.StringValue(response.Description)
	data.CpuVcore = types.Int64Value(int64(response.CpuVcore))
	data.MemoryGib = types.Int64Value(int64(response.MemoryGib))

	devices, diags := types.ListValueFrom(ctx, types.StringType, response.Devices)
	if diags.HasError() {
		return diags
	}

	data.Devices = devices
	data.PricePerHour = types.StringValue(response.PricePerHour)
	data.Activated = types.BoolValue(response.Activated)

	return diag.Diagnostics{}
}

func (d *InstanceTypeDataSource) Read(
	ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse,
) {
	var config InstanceTypeDataSourceModel
	var state InstanceTypeDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filterActivated := true

	instances, err := d.client.GetInstanceTypes(config.Name.ValueStringPointer(), &filterActivated, 0, 2)

	if err != nil {
		resp.Diagnostics.AddError(
			"error while fetching instance types",
			fmt.Sprintf("error: %v", err.Error()),
		)
		return
	}

	if len(instances) == 0 {
		resp.Diagnostics.AddError(
			"No such instance type",
			"Zero instance type is returned. Please check your instance type name",
		)
		return
	}

	if len(instances) > 1 {
		resp.Diagnostics.AddError(
			"Multiple instance type returned",
			"Multiple instance type is returned. Select instance type using id",
		)
		return
	}

	resp.Diagnostics.Append(
		InstanceTypeGetResponseToInstanceTypeModel(ctx, &instances[0], &state)...,
	)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
