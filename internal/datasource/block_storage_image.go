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
	_ datasource.DataSource              = &BlockStorageImageDataSource{}
	_ datasource.DataSourceWithConfigure = &BlockStorageImageDataSource{}
)

func NewBlockStorageImageDataSource() datasource.DataSource {
	return &BlockStorageImageDataSource{}
}

type BlockStorageImageDataSource struct {
	client *api.APIClient
}

type BlockStorageImageDataSourceModel struct {
	Id          types.String `tfsdk:"id"`
	Created     types.String `tfsdk:"created"`
	ZoneId      types.String `tfsdk:"zone_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Keywords    types.List   `tfsdk:"keywords"`
	SizeGib     types.Int64  `tfsdk:"size_gib"`
	Status      types.String `tfsdk:"status"`
}

func (d *BlockStorageImageDataSource) Configure(
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

func (d *BlockStorageImageDataSource) Metadata(
	_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_block_storage_image"
}

func (d *BlockStorageImageDataSource) Schema(
	_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Block Storage Image",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "unique identifier of the block storage image",
				Computed:    true,
			},
			"created": schema.StringAttribute{
				Description: "the time when the block storage image is created",
				Computed:    true,
			},
			"zone_id": schema.StringAttribute{
				Description: "id of zone that the block storage image belongs to",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "human-readable name of the block storage image",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "description of the block storage image",
				Computed:    true,
			},
			"keywords": schema.ListAttribute{
				Description: "keywords describing the block storage image",
				ElementType: types.StringType,
				Computed:    true,
			},
			"size_gib": schema.Int64Attribute{
				Description: "size of the block storage image (GiB)",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "status of the block storage image",
				Computed:    true,
			},
		},
	}
}

func BlockStorageImageGetResponseToBlockStorageImageModel(
	ctx context.Context,
	response *api.ResourceBlockStorageImageGetResponse,
	data *BlockStorageImageDataSourceModel,
) diag.Diagnostics {
	data.Id = types.StringValue(response.Id.String())
	data.Created = types.StringValue(response.Created.String())
	data.ZoneId = types.StringValue(response.ZoneId.String())
	data.Name = types.StringValue(response.Name)
	data.Description = types.StringValue(response.Description)

	keywords, diags := types.ListValueFrom(ctx, types.StringType, response.Keywords)
	if diags.HasError() {
		return diags
	}

	data.Keywords = keywords
	data.SizeGib = types.Int64Value(int64(response.SizeGib))
	data.Status = types.StringValue(response.Status)

	return diag.Diagnostics{}
}

func (d *BlockStorageImageDataSource) Read(
	ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse,
) {
	var config BlockStorageImageDataSourceModel
	var state BlockStorageImageDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	images, err := d.client.GetBlockStorageImages(config.Name.ValueStringPointer(), 0, 2)

	if err != nil {
		resp.Diagnostics.AddError(
			"error while fetching block storage images",
			fmt.Sprintf("error: %v", err.Error()),
		)
		return
	}

	if len(images) == 0 {
		resp.Diagnostics.AddError(
			"No such block storage image",
			"Block storage with the given name does not exist",
		)
		return
	}

	if len(images) > 1 {
		resp.Diagnostics.AddError(
			"Multiple block storage images returned",
			"Multiple block storage is returned. Select block storage image using its id",
		)
		return
	}

	resp.Diagnostics.Append(
		BlockStorageImageGetResponseToBlockStorageImageModel(ctx, &images[0], &state)...,
	)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
