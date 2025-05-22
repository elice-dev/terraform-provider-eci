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
	_ datasource.DataSource              = &RegionDataSource{}
	_ datasource.DataSourceWithConfigure = &RegionDataSource{}
)

func NewRegionDataSource() datasource.DataSource {
	return &RegionDataSource{}
}

type RegionDataSource struct {
	client *api.APIClient
}

type RegionDataSourceModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *RegionDataSource) Configure(
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

func (d *RegionDataSource) Metadata(
	_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_region"
}

func (d *RegionDataSource) Schema(
	_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Region",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "unique identifier of the region",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "human-readable name of the region",
				Required:    true,
			},
		},
	}
}

func RegionGetResponseToRegionModel(
	ctx context.Context,
	response *api.RegionGetResponse,
	data *RegionDataSourceModel,
) diag.Diagnostics {
	data.Id = types.StringValue(response.Id.String())
	data.Name = types.StringValue(response.Name)

	return diag.Diagnostics{}
}

func (d *RegionDataSource) Read(
	ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse,
) {
	var config RegionDataSourceModel
	var state RegionDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	regions, err := d.client.GetRegions(config.Name.ValueStringPointer(), 0, 2)

	if err != nil {
		resp.Diagnostics.AddError(
			"error while fetching regions",
			fmt.Sprintf("error: %v", err.Error()),
		)
		return
	}

	if len(regions) == 0 {
		resp.Diagnostics.AddError(
			"No such region",
			"Zero zone is returned. Please check the name of zone",
		)
		return
	}

	if len(regions) > 1 {
		resp.Diagnostics.AddError(
			"Multiple zone returned",
			"Multiple zone is returned. Select zone using id",
		)
		return
	}

	resp.Diagnostics.Append(RegionGetResponseToRegionModel(ctx, &regions[0], &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
