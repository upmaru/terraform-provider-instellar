package instellar

import (
	"context"
	"os"

	// instellar client = instc
	instc "github.com/upmaru/instellar-go"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/upmaru/terraform-provider-instellar/instellar/cluster"
)

var (
	_ provider.Provider = &instellarProvider{}
)

func New() provider.Provider {
	return &instellarProvider{}
}

type instellarProvider struct{}

type instellarProviderModel struct {
	Host      types.String `tfsdk:"host"`
	AuthToken types.String `tfsdk:"auth_token"`
}

func (p *instellarProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "instellar"
}

func (p *instellarProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provision instellar resources.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "Host for instellar API. May also be provided via INSTELLAR_HOST env variable.",
				Optional:    true,
			},
			"auth_token": schema.StringAttribute{
				Description: "Authentication token for instellar, May also be provided via INSTELLAR_AUTH_TOKEN env variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *instellarProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Instellar client")

	var config instellarProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if config.AuthToken.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("auth_token"),
			"Unknown Instellar Auth Token",
			"The provider cannot create Instellar API client as there is an unknown configuration value for the Instellar API auth token."+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the INSTELLAR_AUTH_TOKEN env variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	host := os.Getenv("INSTELLAR_HOST")
	auth_token := os.Getenv("INSTELLAR_AUTH_TOKEN")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.AuthToken.IsNull() {
		auth_token = config.AuthToken.ValueString()
	}

	if host == "" {
		host = "https://web.instellar.app"
	}

	if auth_token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("auth_token"),
			"Missing Instellar API Auth Token",
			"The provider cannot create Instellar API client as there is a missing or empty value for the Instellar API auth token. "+
				"Set the auth_token value in teh configuration or use the INSTELLAR_AUTH_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "instellar_host", host)
	ctx = tflog.SetField(ctx, "instellar_auth_token", auth_token)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "instellar_auth_token")

	tflog.Debug(ctx, "Creating Instellar client")

	client, err := instc.NewClient(&host, &auth_token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Instellar API Client",
			"An unexected error occured when creating Instellar API client. "+
				"If the error is not clear, please contact provider developers.\n\n"+
				"Instellar Client Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured Instellar client", map[string]any{"success": true})
}

func (p *instellarProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *instellarProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		cluster.NewClusterResource,
	}
}
