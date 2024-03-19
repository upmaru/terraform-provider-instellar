package uplink

import (
	"context"
	"fmt"

	instc "github.com/upmaru/instellar-go"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &uplinkDataSource{}
	_ datasource.DataSourceWithConfigure = &uplinkDataSource{}
)

func NewUplinkDataSource() datasource.DataSource {
	return &uplinkDataSource{}
}

type uplinkDataSource struct {
	client *instc.Client
}

type uplinkDataSourceModel struct {
	ID    types.String `tfsdk:"id"`
	Nodes types.List   `tfsdk:"nodes"`
}

func (d *uplinkDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_uplink"
}

func (d *uplinkDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Uplink provides a list of nodes running uplink",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Uplink id",
				Required:    true,
			},
			"nodes": schema.ListAttribute{
				Description: "List of nodes",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *uplinkDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*instc.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *instc.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *uplinkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state uplinkDataSourceModel

	diags := req.Config.Get(ctx, &state)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	uplink, err := d.client.GetUplink(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading uplink",
			"Could not read uplink id "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	Nodes, di := types.ListValueFrom(ctx, types.StringType, uplink.Data.Attributes.Nodes)

	resp.Diagnostics.Append(di...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Nodes = Nodes

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
