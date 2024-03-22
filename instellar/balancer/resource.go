package balancer

import (
	"context"
	"fmt"
	"strconv"
	"time"

	// instellar client = instc.
	instc "github.com/upmaru/instellar-go"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &balancerResource{}
	_ resource.ResourceWithConfigure   = &balancerResource{}
	_ resource.ResourceWithImportState = &balancerResource{}
)

func NewBalancerResource() resource.Resource {
	return &balancerResource{}
}

type balancerResource struct {
	client *instc.Client
}

type balancerResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Address      types.String `tfsdk:"address"`
	CurrentState types.String `tfsdk:"current_state"`
	ClusterID    types.String `tfsdk:"cluster_id"`
	LastUpdated  types.String `tfsdk:"last_updated"`
}

func (r *balancerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_balancer"
}

func (r *balancerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Balancer registers the load balancer address from your infrastructure load balancer and tells OpsMaru to use the load balancer address for communicating with the cluster.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Balancer identifier",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Balancer Name",
				Required:    true,
			},
			"address": schema.StringAttribute{
				Description: "Balancer Address",
				Required:    true,
			},
			"current_state": schema.StringAttribute{
				Description: "Balancer Current State",
				Computed:    true,
			},
			"cluster_id": schema.StringAttribute{
				Description: "Which cluster does balancer belong to",
				Required:    true,
			},
			"last_updated": schema.StringAttribute{
				Description: "Balancer Last Updated",
				Computed:    true,
			},
		},
	}
}

func (r *balancerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *balancerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan balancerResourceModel
	diags := req.Plan.Get(ctx, &plan)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	balancerParams := instc.BalancerParams{
		Name:    plan.Name.ValueString(),
		Address: plan.Address.ValueString(),
	}

	balancer, err := r.client.CreateBalancer(plan.ClusterID.ValueString(), balancerParams)

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create balancer",
			"Could not create balancer, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(strconv.Itoa(balancer.Data.Attributes.ID))
	plan.CurrentState = types.StringValue(balancer.Data.Attributes.CurrentState)
	plan.ClusterID = types.StringValue(strconv.Itoa(balancer.Data.Attributes.ClusterID))
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *balancerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state balancerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	balancer, err := r.client.GetBalancer(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading uplink",
			"Could not read uplink id "+state.ID.ValueString()+": "+err.Error(),
		)
	}

	state.Name = types.StringValue(balancer.Data.Attributes.Name)
	state.Address = types.StringValue(balancer.Data.Attributes.Address)
	state.CurrentState = types.StringValue(balancer.Data.Attributes.CurrentState)
	state.ClusterID = types.StringValue(strconv.Itoa(balancer.Data.Attributes.ClusterID))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *balancerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan balancerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	balancerParams := instc.BalancerParams{
		Name:    plan.Name.ValueString(),
		Address: plan.Address.ValueString(),
	}

	_, err := r.client.UpdateBalancer(plan.ID.ValueString(), balancerParams)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating balancer",
			"Could not update balancer, unexpected error: "+err.Error(),
		)
		return
	}

	balancer, err := r.client.GetBalancer(plan.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading balancer",
			"Could not read balancer ID "+plan.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	plan.Name = types.StringValue(balancer.Data.Attributes.Name)
	plan.Address = types.StringValue(balancer.Data.Attributes.Address)
	plan.CurrentState = types.StringValue(balancer.Data.Attributes.CurrentState)
	plan.ClusterID = types.StringValue(strconv.Itoa(balancer.Data.Attributes.ClusterID))
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *balancerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *balancerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteBalancer(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting balancer",
			"Cloud not delete balancer, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *balancerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
