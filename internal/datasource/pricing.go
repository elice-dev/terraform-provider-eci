package datasource

import (
	"context"
	"fmt"
	"terraform-provider-eci/internal/api"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &PricingDataSource{}
	_ datasource.DataSourceWithConfigure = &PricingDataSource{}
)

func NewPricingDataSource() datasource.DataSource {
	return &PricingDataSource{}
}

type PricingDataSource struct {
	client *api.APIClient
}

type PricingDataSourceModel struct {
	Id                 types.String `tfsdk:"id"`
	Tags               types.Map    `tfsdk:"tags"`
	Created            types.String `tfsdk:"created"`
	Modified           types.String `tfsdk:"modified"`
	OrganizationId     types.String `tfsdk:"organization_id"`
	ZoneId             types.String `tfsdk:"zone_id"`
	ResourceKind       types.String `tfsdk:"resource_kind"`
	ResourceId         types.String `tfsdk:"resource_id"`
	Name               types.String `tfsdk:"name"`
	PricingType        types.String `tfsdk:"pricing_type"`
	PricePerHour       types.String `tfsdk:"price_per_hour"`
	ListedPricePerHour types.String `tfsdk:"listed_price_per_hour"`
	Amount             types.Map    `tfsdk:"amount"`
	Start              types.String `tfsdk:"start"`
	End                types.String `tfsdk:"end"`
	Activated          types.Bool   `tfsdk:"activated"`
	Quota              types.Int64  `tfsdk:"quota"`
}

func (d *PricingDataSource) Configure(
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

func (d *PricingDataSource) Metadata(
	_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_pricing"
}

func (d *PricingDataSource) Schema(
	_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Pricing",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "unique identifier of the pricing",
				Computed:    true,
			},
			"tags": schema.MapAttribute{
				ElementType: types.StringType,
				Description: "user-defined metadata of key-value pairs",
				Computed:    true,
			},
			"created": schema.StringAttribute{
				Description: "time when the pricing is created",
				Computed:    true,
			},
			"modified": schema.StringAttribute{
				Description: "last time when the pricing is modified",
				Computed:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "id of organization that the pricing belongs to (null for global pricing)",
				Computed:    true,
			},
			"zone_id": schema.StringAttribute{
				Description: "id of zone that the pricing belongs to",
				Computed:    true,
			},
			"resource_kind": schema.StringAttribute{
				Description: "kind of resource this pricing applies to (vm_allocation, block_storage, public_ip, etc.)",
				Optional:    true,
			},
			"resource_id": schema.StringAttribute{
				Description: "id of specific resource instance (null for generic pricing)",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "human-readable name of the pricing",
				Required:    true,
			},
			"pricing_type": schema.StringAttribute{
				Description: "type of pricing (ondemand, reserved)",
				Required:    true,
			},
			"price_per_hour": schema.StringAttribute{
				Description: "actual price per hour",
				Computed:    true,
			},
			"listed_price_per_hour": schema.StringAttribute{
				Description: "listed price per hour",
				Computed:    true,
			},
			"amount": schema.MapNestedAttribute{
				Description: "price per hour by currency (krw, usd); usd is in cents",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"per_hour": schema.StringAttribute{
							Description: "actual price per hour",
							Computed:    true,
						},
						"listed_per_hour": schema.StringAttribute{
							Description: "listed price per hour",
							Computed:    true,
						},
					},
				},
			},
			"start": schema.StringAttribute{
				Description: "start time of this pricing validity period",
				Computed:    true,
			},
			"end": schema.StringAttribute{
				Description: "end time of this pricing validity period",
				Computed:    true,
			},
			"activated": schema.BoolAttribute{
				Description: "whether this pricing is activated",
				Computed:    true,
			},
			"quota": schema.Int64Attribute{
				Description: "quota limit for this pricing (null for unlimited)",
				Computed:    true,
			},
		},
	}
}

func PricingGetResponseToPricingModel(
	ctx context.Context,
	response *api.PricingGetResponse,
	data *PricingDataSourceModel,
) diag.Diagnostics {
	data.Id = types.StringValue(response.Id.String())

	tags, diags := types.MapValueFrom(ctx, types.StringType, response.Tags)
	if diags.HasError() {
		return diags
	}
	data.Tags = tags

	data.Created = types.StringValue(response.Created.String())

	if response.Modified != nil {
		data.Modified = types.StringValue(response.Modified.String())
	} else {
		data.Modified = types.StringNull()
	}

	if response.OrganizationId != nil {
		data.OrganizationId = types.StringValue(response.OrganizationId.String())
	} else {
		data.OrganizationId = types.StringNull()
	}

	data.ZoneId = types.StringValue(response.ZoneId.String())
	data.ResourceKind = types.StringValue(response.ResourceKind)

	if response.ResourceId != nil {
		data.ResourceId = types.StringValue(response.ResourceId.String())
	} else {
		data.ResourceId = types.StringNull()
	}

	data.Name = types.StringValue(response.Name)
	data.PricingType = types.StringValue(response.PricingType)
	data.PricePerHour = types.StringValue(response.PricePerHour)
	data.ListedPricePerHour = types.StringValue(response.ListedPricePerHour)

	amountType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"per_hour":        types.StringType,
			"listed_per_hour": types.StringType,
		},
	}

	if response.Amount != nil {
		amount, diags := types.MapValueFrom(ctx, amountType, response.Amount)
		if diags.HasError() {
			return diags
		}
		data.Amount = amount
	} else {
		data.Amount = types.MapNull(amountType)
	}

	if response.Start != nil {
		data.Start = types.StringValue(response.Start.String())
	} else {
		data.Start = types.StringNull()
	}

	if response.End != nil {
		data.End = types.StringValue(response.End.String())
	} else {
		data.End = types.StringNull()
	}

	data.Activated = types.BoolValue(response.Activated)

	if response.Quota != nil {
		data.Quota = types.Int64Value(int64(*response.Quota))
	} else {
		data.Quota = types.Int64Null()
	}

	return diag.Diagnostics{}
}

func (d *PricingDataSource) Read(
	ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse,
) {
	var config PricingDataSourceModel
	var state PricingDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filterActivated := true

	var filterResourceKind *string = nil
	if !config.ResourceKind.IsNull() {
		filterResourceKind = config.ResourceKind.ValueStringPointer()
	}

	pricings, err := d.client.GetPricings(
		config.Name.ValueStringPointer(),
		filterResourceKind,
		config.PricingType.ValueStringPointer(),
		&filterActivated,
		0,
		2,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"error while fetching pricings",
			fmt.Sprintf("error: %v", err.Error()),
		)
		return
	}

	if len(pricings) == 0 {
		resp.Diagnostics.AddError(
			"No such pricing",
			"Zero pricing is returned. Please check your pricing name",
		)
		return
	}

	if len(pricings) > 1 {
		resp.Diagnostics.AddError(
			"Multiple pricings returned",
			"Multiple pricing is returned. Select instance type using id",
		)
		return
	}

	resp.Diagnostics.Append(
		PricingGetResponseToPricingModel(ctx, &pricings[0], &state)...,
	)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
