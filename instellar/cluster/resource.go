package cluster

import (
	"context"
	"time"

	// instellar client = instc
	instc "github.com/upmaru/instellar-go"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &clusterResource{}
	_ resource.ResourceWithConfigure   = &clusterResource{}
	_ resource.ResourceWithImportState = &clusterResource{}
)

func NewClusterResource() resource.Resource {
	return &clusterResource{}
}

type clusterResource struct {
	client *instc.Client
}

type clusterResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Slug          types.String `tfsdk:"slug"`
	CurrentState  types.String `tfsdk:"current_state"`
	ProviderName  types.String `tfsdk:"provider_name"`
	Region        types.String `tfsdk:"region"`
	Endpoint      types.String `tfsdk:"endpoint"`
	PasswordToken types.String `tfsdk:"password_token"`
	LastUpdated   types.String `tfsdk:"last_updated"`
}

func (r *clusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (r *clusterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Cluster management",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Cluster identifier",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name assigned by the user",
				Required:    true,
			},
			"slug": schema.StringAttribute{
				Description: "Unique slug for cluster",
				Computed:    true,
			},
			"current_state": schema.StringAttribute{
				Description: "Endpoint for the cluster",
				Computed:    true,
			},
			"provider_name": schema.StringAttribute{
				Description: "Provider of the infrastructure",
				Required:    true,
			},
			"region": schema.StringAttribute{
				Description: "Region of the cluster",
				Required:    true,
			},
			"endpoint": schema.StringAttribute{
				Description: "Endpoint for cluster",
				Required:    true,
			},
			"password_token": schema.StringAttribute{
				Description: "Password or Trust Token for cluster",
				Required:    true,
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the terraform update",
				Computed:    true,
			},
		},
	}
}

func (r *clusterResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*instc.Client)
}

func (r *clusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan clusterResourceModel
	diags := req.Plan.Get(ctx, &plan)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	clusterParams := instc.ClusterParams{
		Name:                           plan.Name.ValueString(),
		Provider:                       plan.ProviderName.ValueString(),
		Region:                         plan.Region.ValueString(),
		CredentialEndpoint:             plan.Endpoint.ValueString(),
		CredentialPassword:             plan.PasswordToken.ValueString(),
		CredentialPasswordConfirmation: plan.PasswordToken.ValueString(),
	}

	cluster, err := r.client.CreateCluster(clusterParams)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating instellar cluster",
			"Cloud not create order, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(cluster.Data.Attributes.Slug)
	plan.Name = types.StringValue(cluster.Data.Attributes.Name)
	plan.Slug = types.StringValue(cluster.Data.Attributes.Slug)
	plan.Region = types.StringValue(cluster.Data.Attributes.Region)
	plan.CurrentState = types.StringValue(cluster.Data.Attributes.CurrentState)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *clusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state clusterResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cluster, err := r.client.GetCluster(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading instellar cluster",
			"Cloud not read cluster id "+state.Slug.ValueString()+": "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(cluster.Data.Attributes.Name)
	state.Slug = types.StringValue(cluster.Data.Attributes.Slug)
	state.Endpoint = types.StringValue(cluster.Data.Attributes.Endpoint)
	state.ProviderName = types.StringValue(cluster.Data.Attributes.Provider)
	state.Region = types.StringValue(cluster.Data.Attributes.Region)
	state.CurrentState = types.StringValue(cluster.Data.Attributes.CurrentState)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *clusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan clusterResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.UpdateCluster(plan.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating instellar cluster",
			"Could not update cluster, unexpected error: "+err.Error(),
		)
		return
	}

	cluster, err := r.client.GetCluster(plan.Slug.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading instellar cluster",
			"Could not read instellar cluster ID "+plan.Slug.ValueString()+": "+err.Error(),
		)
	}

	plan.CurrentState = types.StringValue(cluster.Data.Attributes.CurrentState)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *clusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state clusterResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteCluster(state.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting cluster",
			"cloud not delete cluster, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *clusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
