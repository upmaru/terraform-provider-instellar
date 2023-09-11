package component

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	instc "github.com/upmaru/instellar-go"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
)

var (
	_ resource.Resource                = &componentResource{}
	_ resource.ResourceWithConfigure   = &componentResource{}
	_ resource.ResourceWithImportState = &componentResource{}
)

func NewComponentResource() resource.Resource {
	return &componentResource{}
}

type componentResource struct {
	client *instc.Client
}

type componentResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Slug                types.String `tfsdk:"slug"`
	DriverVersion       types.String `tfsdk:"driver_version"`
	CurrentState        types.String `tfsdk:"current_state"`
	ProviderName        types.String `tfsdk:"provider_name"`
	Driver              types.String `tfsdk:"driver"`
	ClusterIDS          types.List   `tfsdk:"cluster_ids"`
	Channels            types.List   `tfsdk:"channels"`
	Credential          types.Object `tfsdk:"credential"`
	InsterraComponentID types.Int64  `tfsdk:"insterra_component_id"`
	LastUpdated         types.String `tfsdk:"last_updated"`
}

type componentCredentialResourceModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Resource types.String `tfsdk:"resource"`
	Host     types.String `tfsdk:"host"`
	Port     types.Int64  `tfsdk:"port"`
	Secure   types.Bool   `tfsdk:"secure"`
}

func (r *componentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_component"
}

func (r *componentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Components enable you to add things like PostgreSQL, MySQL, Redis or any other 'components' and associate them to a given cluster.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Component identifier",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the component assigned by the user",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 64),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9\-]+$`),
						"must contain only lowercase alphanumeric characters",
					),
				},
			},
			"slug": schema.StringAttribute{
				Description: "Unique slug for component",
				Computed:    true,
			},
			"current_state": schema.StringAttribute{
				Description: "Current state for the component",
				Computed:    true,
			},
			"driver_version": schema.StringAttribute{
				Description: "Version of the driver",
				Required:    true,
			},
			"provider_name": schema.StringAttribute{
				Description: "Provider of the infrastructure",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"aws", "hcloud", "digitalocean", "google", "azurerm"}...),
				},
			},
			"driver": schema.StringAttribute{
				Description: "Driver of the component",
				Required:    true,
			},
			"cluster_ids": schema.ListAttribute{
				Description: "Cluster ids to attach component",
				Required:    true,
				ElementType: types.NumberType,
			},
			"channels": schema.ListAttribute{
				Description: "Channels to restrict component availability",
				Required:    true,
				ElementType: types.StringType,
			},
			"insterra_component_id": schema.Int64Attribute{
				Description: "Reference to insterra component",
				Optional:    true,
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of terraform update",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"credential": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						Required:    true,
						Description: "Username for the component.",
					},
					"password": schema.StringAttribute{
						Required:    true,
						Sensitive:   true,
						Description: "Password for the component, this field is sensitive.",
					},
					"resource": schema.StringAttribute{
						Required:    true,
						Description: "Resource for the component, this can be the database name or the region name.",
					},
					"host": schema.StringAttribute{
						Required:    true,
						Description: "Host for the component.",
					},
					"port": schema.Int64Attribute{
						Required:    true,
						Description: "Port for the component",
					},
					"secure": schema.BoolAttribute{
						Optional:    true,
						Description: "SSL configuration for the component",
					},
				},
			},
		},
	}
}

func (r *componentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *componentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan componentResourceModel
	diags := req.Plan.Get(ctx, &plan)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var ClusterIDS []int
	var Channels []string
	var Credential componentCredentialResourceModel

	diags = plan.ClusterIDS.ElementsAs(ctx, &ClusterIDS, false)

	resp.Diagnostics.Append(diags...)

	diags = plan.Channels.ElementsAs(ctx, &Channels, false)

	resp.Diagnostics.Append(diags...)

	diags = plan.Credential.As(ctx, &Credential, basetypes.ObjectAsOptions{})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var credentialParams = instc.ComponentCredentialParams{
		Username: Credential.Username.ValueString(),
		Password: Credential.Password.ValueString(),
		Resource: Credential.Resource.ValueString(),
		Host:     Credential.Host.ValueString(),
		Port:     int(Credential.Port.ValueInt64()),
		Secure:   Credential.Secure.ValueBool(),
	}

	componentParams := instc.ComponentParams{
		Name:                plan.Name.ValueString(),
		Provider:            plan.ProviderName.ValueString(),
		Version:             plan.DriverVersion.ValueString(),
		Driver:              plan.Driver.ValueString(),
		ClusterIDS:          ClusterIDS,
		Channels:            Channels,
		InsterraComponentID: int(plan.InsterraComponentID.ValueInt64()),
		Credential:          &credentialParams,
	}

	component, err := r.client.CreateComponent(componentParams)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating instellar component",
			"Cloud not create component, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(strconv.Itoa(component.Data.Attributes.ID))
	plan.Slug = types.StringValue(component.Data.Attributes.Slug)
	plan.CurrentState = types.StringValue(component.Data.Attributes.CurrentState)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *componentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state componentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	component, err := r.client.GetComponent(state.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading instellar component",
			"Cloud not read component id "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(component.Data.Attributes.Slug)
	state.Slug = types.StringValue(component.Data.Attributes.Slug)
	state.CurrentState = types.StringValue(component.Data.Attributes.CurrentState)
	state.Driver = types.StringValue(component.Data.Attributes.Driver)
	state.ProviderName = types.StringValue(component.Data.Attributes.Provider)
	state.DriverVersion = types.StringValue(component.Data.Attributes.Version)

	ClusterIDS, d := types.ListValueFrom(ctx, types.NumberType, component.Data.Attributes.ClusterIDS)

	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ClusterIDS = ClusterIDS

	Channels, d := types.ListValueFrom(ctx, types.StringType, component.Data.Attributes.Channels)

	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Channels = Channels

	credentialData := componentCredentialResourceModel{
		Username: types.StringValue(component.Data.Attributes.Credential.Username),
		Password: types.StringValue(component.Data.Attributes.Credential.Password),
		Resource: types.StringValue(component.Data.Attributes.Credential.Resource),
		Host:     types.StringValue(component.Data.Attributes.Credential.Host),
		Port:     types.Int64Value(int64(component.Data.Attributes.Credential.Port)),
		Secure:   types.BoolValue(component.Data.Attributes.Credential.Secure),
	}

	Credential, d := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"username": types.StringType,
		"password": types.StringType,
		"resource": types.StringType,
		"host":     types.StringType,
		"port":     types.Int64Type,
		"secure":   types.BoolType,
	}, credentialData)

	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Credential = Credential

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *componentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan componentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var ClusterIDS []int
	var Channels []string

	diags = plan.ClusterIDS.ElementsAs(ctx, &ClusterIDS, false)
	resp.Diagnostics.Append(diags...)

	diags = plan.Channels.ElementsAs(ctx, &Channels, false)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	componentParams := instc.ComponentParams{
		ClusterIDS: ClusterIDS,
		Channels:   Channels,
	}

	_, err := r.client.UpdateComponent(plan.ID.ValueString(), componentParams)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating instellar component",
			"Could not update component, unexpected error: "+err.Error(),
		)
		return
	}

	component, err := r.client.GetComponent(plan.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading instellar component",
			"Could not read instellar component ID "+plan.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	resultClusterIDS, d := types.ListValueFrom(ctx, types.NumberType, component.Data.Attributes.ClusterIDS)
	resp.Diagnostics.Append(d...)

	resultChannels, d := types.ListValueFrom(ctx, types.StringType, component.Data.Attributes.Channels)
	resp.Diagnostics.Append(d...)

	if resp.Diagnostics.HasError() {
		return
	}

	plan.Name = types.StringValue(component.Data.Attributes.Slug)
	plan.Slug = types.StringValue(component.Data.Attributes.Slug)
	plan.CurrentState = types.StringValue(component.Data.Attributes.CurrentState)
	plan.ClusterIDS = resultClusterIDS
	plan.Channels = resultChannels
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *componentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state componentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteComponent(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting component",
			"cloud not delete component, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *componentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
