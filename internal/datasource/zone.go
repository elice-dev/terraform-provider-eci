package datasource

import (
	"context"
	"fmt"
	"terraform-provider-eci/internal/api"
	. "terraform-provider-eci/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &ZoneDataSource{}
	_ datasource.DataSourceWithConfigure = &ZoneDataSource{}
)

func NewZoneDataSource() datasource.DataSource {
	return &ZoneDataSource{}
}

type ZoneDataSource struct {
	client *api.APIClient
}

type ZoneDataSourceModel struct {
	Id              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	RegionId        types.String `tfsdk:"region_id"`
	SecondaryZoneId types.String `tfsdk:"secondary_zone_id"`
}

func (d *ZoneDataSource) Configure(
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

func (d *ZoneDataSource) Metadata(
	_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_zone"
}

func (d *ZoneDataSource) Schema(
	_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Zone",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "unique identifier of the zone",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "human-readable name of the zone",
				Required:    true,
			},
			"region_id": schema.StringAttribute{
				Description: "id of the region that the zone belongs to",
				Required:    true,
			},
			"secondary_zone_id": schema.StringAttribute{
				Description: "id of the secondary zone that this zone will fail over when DR",
				Computed:    true,
			},
		},
	}
}

func ZoneGetResponseToZoneModel(
	ctx context.Context,
	response *api.InfraZoneGetResponse,
	data *ZoneDataSourceModel,
) diag.Diagnostics {
	data.Id = types.StringValue(response.Id.String())
	data.Name = types.StringValue(response.Name)
	data.RegionId = types.StringValue(response.RegionId.String())

	data.SecondaryZoneId = StringOrNull(response.SecondaryZoneId)

	return diag.Diagnostics{}
}

func (d *ZoneDataSource) Read(
	ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse,
) {
	var config ZoneDataSourceModel
	var state ZoneDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zones, err := d.client.GetZones(
		config.RegionId.ValueStringPointer(), config.Name.ValueStringPointer(), 0, 2)

	if err != nil {
		resp.Diagnostics.AddError(
			"error while fetching zones",
			fmt.Sprintf("error: %v", err.Error()),
		)
		return
	}

	if len(zones) == 0 {
		resp.Diagnostics.AddError(
			"No such zone",
			"Zero zone is returned. Please check the name of zone",
		)
		return
	}

	if len(zones) > 1 {
		resp.Diagnostics.AddError(
			"Multiple zone returned",
			"Multiple zone is returned. Select zone using id",
		)
		return
	}

	resp.Diagnostics.Append(ZoneGetResponseToZoneModel(ctx, &zones[0], &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
