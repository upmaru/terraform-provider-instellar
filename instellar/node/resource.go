package node

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
	_ resource.Resource                = &nodeResource{}
	_ resource.ResourceWithConfigure   = &nodeResource{}
	_ resource.ResourceWithImportState = &nodeResource{}
)

func NewNodeResource() resource.Resource {
	return &nodeResource{}
}

type nodeResource struct {
	client *instc.Client
}

type nodeResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Slug         types.String `tfsdk:"slug"`
	ClusterID    types.String `tfsdk:"cluster_id"`
	PublicIP     types.String `tfsdk:"public_ip"`
	CurrentState types.String `tfsdk:"current_state"`
	LastUpdated  types.String `tfsdk:"last_updated"`
}

func (r *nodeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node"
}

func (r *nodeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Node management",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Node identifier",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"slug": schema.StringAttribute{
				Description: "Node slug",
				Required:    true,
			},
			"current_state": schema.StringAttribute{
				Description: "Current state",
				Computed:    true,
			},
			"public_ip": schema.StringAttribute{
				Description: "Public IP of the node",
				Required:    true,
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of terraform update",
				Computed:    true,
			},
		},
	}
}

func (r *nodeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *nodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan nodeResourceModel
	diags := req.Plan.Get(ctx, &plan)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	nodeParams := instc.NodeParams{
		PublicIP: plan.PublicIP.ValueString(),
	}

	node, err := r.client.CreateNode(plan.ClusterID.ValueString(), plan.Slug.ValueString(), nodeParams)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating node",
			"Could not create node, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(strconv.Itoa(node.Data.Attributes.ID))
	plan.CurrentState = types.StringValue(node.Data.Attributes.CurrentState)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *nodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state nodeResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	node, err := r.client.GetNode(state.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading node",
			"Could not read node id "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.Slug = types.StringValue(node.Data.Attributes.Slug)
	state.ClusterID = types.StringValue(strconv.Itoa(node.Data.Attributes.ClusterID))
	state.PublicIP = types.StringValue(node.Data.Attributes.PublicIP)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *nodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan nodeResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	nodeParams := instc.NodeParams{
		PublicIP: plan.PublicIP.ValueString(),
	}

	_, err := r.client.UpdateNode(plan.ClusterID.ValueString(), plan.Slug.ValueString(), nodeParams)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating node",
			"Could not update node, unexpected error: "+err.Error(),
		)
		return
	}

	node, err := r.client.GetNode(plan.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading node",
			"Could not read node ID "+plan.ID.ValueString()+": "+err.Error(),
		)
	}

	plan.PublicIP = types.StringValue(node.Data.Attributes.PublicIP)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *nodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state nodeResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteNode(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting node",
			"Could not delete node, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *nodeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
