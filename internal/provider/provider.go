package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/shadeform/terraform-provider-shadeform/internal/datasources/instance_types"
	"github.com/shadeform/terraform-provider-shadeform/internal/provider/provider_shadeform"
	"github.com/shadeform/terraform-provider-shadeform/internal/resources/instance"
	"github.com/shadeform/terraform-provider-shadeform/internal/resources/volume"
)

var (
	_ provider.Provider = &ShadeformProvider{}
)

type ShadeformProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type ShadeformProviderModel struct {
	ApiKey types.String `tfsdk:"api_key"`
}

func (p *ShadeformProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "shadeform"
	resp.Version = p.version
}

func (p *ShadeformProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "API key for Shadeform. Can also be set via the SHADEFORM_API_KEY environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *ShadeformProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ShadeformProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client := provider_shadeform.NewClient(data.ApiKey.ValueString())
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ShadeformProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		instance.NewInstanceResource,
		volume.NewVolumeResource,
	}
}

func (p *ShadeformProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		instance_types.NewInstanceTypesDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ShadeformProvider{
			version: version,
		}
	}
}

func (p *ShadeformProvider) ValidateConfig(ctx context.Context, req provider.ValidateConfigRequest, resp *provider.ValidateConfigResponse) {
	var data ShadeformProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ApiKey.IsNull() || data.ApiKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing API Key",
			"The provider cannot create the Shadeform API client as there is a missing or empty value for the Shadeform API key. "+
				"Set the api_key value in the configuration or use the SHADEFORM_API_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}
}
