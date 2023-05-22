package instellar

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
}
