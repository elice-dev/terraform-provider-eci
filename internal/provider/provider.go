package provider

import (
	"context"
	"fmt"
	"net/url"

	"terraform-provider-eci/internal/api"
	ds "terraform-provider-eci/internal/datasource"
	res "terraform-provider-eci/internal/resource"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ provider.Provider = &EliceCloudProvider{}
)

func New(version string, debug bool) func() provider.Provider {
	return func() provider.Provider {
		return &EliceCloudProvider{
			version: version,
			debug:   debug,
		}
	}
}

type EliceCloudProvider struct {
	version string
	debug   bool
}

type EliceCloudProviderModel struct {
	ApiEndpoint    types.String `tfsdk:"api_endpoint"`
	ApiAccessToken types.String `tfsdk:"api_access_token"`
	ZoneId         types.String `tfsdk:"zone_id"`
}

func (p *EliceCloudProvider) Metadata(
	_ context.Context,
	_ provider.MetadataRequest,
	resp *provider.MetadataResponse,
) {
	resp.TypeName = "eci"
	resp.Version = p.version
}

func (p *EliceCloudProvider) Schema(
	_ context.Context,
	_ provider.SchemaRequest,
	resp *provider.SchemaResponse,
) {

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_access_token": schema.StringAttribute{
				Description: "API access token",
				Required:    true,
				Sensitive:   true, // Prevents logging the token
			},
			"api_endpoint": schema.StringAttribute{
				Description: "API endpoint URL (e.g., https://admin.elice.cloud/api/docs)",
				Required:    true},
			"zone_id": schema.StringAttribute{
				Description: "ID of the zone (UUID) that you will manage resources in",
				Required:    true,
			},
		},
	}
}

func (p *EliceCloudProvider) Configure(
	ctx context.Context,
	req provider.ConfigureRequest,
	resp *provider.ConfigureResponse,
) {
	var data EliceCloudProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ApiEndpoint.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_endpoint"),
			"missing api_endpoint",
			"missing api_endpoint",
		)
	}

	if data.ApiAccessToken.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_access_token"),
			"missing api_access_token",
			"missing api_access_token",
		)
	}

	if data.ZoneId.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("zone_id"),
			"missing zone_id",
			"missing zone_id",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	baseURL := data.ApiEndpoint.ValueString()
	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		resp.Diagnostics.AddError(
			"failed to parse API Endpoint",
			fmt.Sprintf("url: %s (reason: %s)", baseURL, err.Error()),
		)
		return
	}

	pathPrefix := parsedBaseURL.Path
	parsedBaseURL.Path = ""
	parsedBaseURL.RawPath = ""

	tflog.Info(
		ctx,
		fmt.Sprintf("zone_id: %s", data.ZoneId.ValueString()),
	)

	client, err := api.NewAPIClient(
		data.ApiAccessToken.ValueString(),
		parsedBaseURL.String(),
		pathPrefix,
		data.ZoneId.ValueString(),
		p.debug,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"failed to create api client", fmt.Sprintf("Error: %v", err.Error()),
		)
		return
	}

	tflog.Info(
		ctx,
		fmt.Sprintf("organizeion_id: %s", client.OrganizationId),
	)

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *EliceCloudProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		func() datasource.DataSource {
			return ds.NewBlockStorageImageDataSource()
		},
		func() datasource.DataSource {
			return ds.NewInstanceTypeDataSource()
		},
		func() datasource.DataSource {
			return ds.NewRegionDataSource()
		},
		func() datasource.DataSource {
			return ds.NewZoneDataSource()
		},
	}
}

func (p *EliceCloudProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource {
			return res.NewResourceBlockStorage()
		},
		func() resource.Resource {
			return res.NewResourceVirtualMachine()
		},
		func() resource.Resource {
			return res.NewResourceVirtualMachineAllocation()
		},
		func() resource.Resource {
			return res.NewResourceVirtualNetwork()
		},
		func() resource.Resource {
			return res.NewResourceSubnet()
		},
		func() resource.Resource {
			return res.NewResourceNetworkInterface()
		},
		func() resource.Resource {
			return res.NewResourcePublicIp()
		},
	}
}
