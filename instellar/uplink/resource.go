package uplink

import (
	"context"
	"fmt"
	"strconv"
	"time"

	instc "github.com/upmaru/instellar-go"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &uplinkResource{}
	_ resource.ResourceWithConfigure   = &uplinkResource{}
	_ resource.ResourceWithImportState = &uplinkResource{}
)

func NewUplinkResource() resource.Resource {
	return &uplinkResource{}
}

type uplinkResource struct {
	client *instc.Client
}

type uplinkResourceModel struct {
	ID           types.String `tfsdk:"id"`
	ChannelSlug  types.String `tfsdk:"channel_slug"`
	CurrentState types.String `tfsdk:"current_state"`
	ClusterID    types.String `tfsdk:"cluster_id"`
	DatabaseURL  types.String `tfsdk:"database_url"`
	LastUpdated  types.String `tfsdk:"last_updated"`
}

func (r *uplinkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_uplink"
}

func (r *uplinkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Uplink provides, ingress management, deployment management and environment variable management on your cluster. It routes traffic using caddy and makes sure caddy's config is up-to-date.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Uplink identifier",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"channel_slug": schema.StringAttribute{
				Description: "Which distribution channel are we using?",
				Required:    true,
			},
			"current_state": schema.StringAttribute{
				Description: "The current state of uplink",
				Computed:    true,
			},
			"cluster_id": schema.StringAttribute{
				Description: "Which cluster does uplink belong to",
				Required:    true,
			},
			"database_url": schema.StringAttribute{
				Description: "Database URL to use with uplink, if supplied will setup uplink pro",
				Optional:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of terraform update",
				Computed:    true,
			},
		},
	}
}

func (r *uplinkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*instc.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf(
				"Expected *instc.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)
		return
	}

	r.client = client
}

func (r *uplinkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan uplinkResourceModel
	diags := req.Plan.Get(ctx, &plan)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	uplinkSetupParams := buildSetupParams(plan)

	uplink, err := r.client.CreateUplink(plan.ClusterID.ValueString(), uplinkSetupParams)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating uplink",
			"Could not create uplink, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(strconv.Itoa(uplink.Data.Attributes.ID))
	plan.CurrentState = types.StringValue(uplink.Data.Attributes.CurrentState)
	plan.ClusterID = types.StringValue(strconv.Itoa(uplink.Data.Attributes.ClusterID))
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	if uplink.Data.Attributes.DatabaseURL != nil {
		plan.DatabaseURL = types.StringValue(*uplink.Data.Attributes.DatabaseURL)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func buildSetupParams(plan uplinkResourceModel) instc.UplinkSetupParams {
	if plan.ClusterID.IsNull() {
		return instc.UplinkSetupParams{
			ChannelSlug: plan.ChannelSlug.ValueString(),
		}
	} else {
		return instc.UplinkSetupParams{
			ChannelSlug: plan.ChannelSlug.ValueString(),
			DatabaseURL: plan.DatabaseURL.ValueString(),
		}
	}
}

func (r *uplinkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state uplinkResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	uplink, err := r.client.GetUplink(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading uplink",
			"Could not read uplink id "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.ChannelSlug = types.StringValue(uplink.Data.Attributes.ChannelSlug)
	state.CurrentState = types.StringValue(uplink.Data.Attributes.CurrentState)
	state.ClusterID = types.StringValue(strconv.Itoa(uplink.Data.Attributes.ClusterID))

	if uplink.Data.Attributes.DatabaseURL != nil {
		state.DatabaseURL = types.StringValue(*uplink.Data.Attributes.DatabaseURL)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *uplinkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan uplinkResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.UpdateUplink(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating uplink",
			"Could not update uplink, unexpected error: "+err.Error(),
		)
		return
	}

	uplink, err := r.client.GetUplink(plan.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading uplink",
			"Could not read uplink ID "+plan.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	plan.ChannelSlug = types.StringValue(uplink.Data.Attributes.ChannelSlug)
	plan.CurrentState = types.StringValue(uplink.Data.Attributes.CurrentState)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	if uplink.Data.Attributes.DatabaseURL != nil {
		plan.DatabaseURL = types.StringValue(*uplink.Data.Attributes.DatabaseURL)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *uplinkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state uplinkResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteUplink(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting uplink",
			"Could not delete uplink, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *uplinkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
