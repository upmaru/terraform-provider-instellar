package storage

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
	_ resource.Resource                = &storageResource{}
	_ resource.ResourceWithConfigure   = &storageResource{}
	_ resource.ResourceWithImportState = &storageResource{}
)

func NewStorageResource() resource.Resource {
	return &storageResource{}
}

type storageResource struct {
	client *instc.Client
}

type storageResourceModel struct {
	ID              types.String `tfsdk:"id"`
	CurrentState    types.String `tfsdk:"current_state"`
	Host            types.String `tfsdk:"host"`
	Bucket          types.String `tfsdk:"bucket"`
	Region          types.String `tfsdk:"region"`
	AccessKeyID     types.String `tfsdk:"access_key_id"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
	LastUpdated     types.String `tfsdk:"last_updated"`
}

func (r *storageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage"
}

func (r *storageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Storage is what instellar use to store all the build artifacts / ssl certificates / others. Basically anything that's needed to manage your deployments.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Storage Identifier",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"current_state": schema.StringAttribute{
				Description: "Current State",
				Computed:    true,
			},
			"host": schema.StringAttribute{
				Description: "Hostname of storage",
				Required:    true,
			},
			"bucket": schema.StringAttribute{
				Description: "Bucket name",
				Required:    true,
			},
			"region": schema.StringAttribute{
				Description: "Region of storage",
				Required:    true,
			},
			"access_key_id": schema.StringAttribute{
				Description: "Access key id",
				Required:    true,
			},
			"secret_access_key": schema.StringAttribute{
				Description: "Secret access key",
				Required:    true,
				Sensitive:   true,
			},
			"last_updated": schema.StringAttribute{
				Description: "Timesmap of teraform update",
				Computed:    true,
			},
		},
	}
}

func (r *storageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *storageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan storageResourceModel
	diags := req.Plan.Get(ctx, &plan)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	storageParams := instc.StorageParams{
		Host:                      plan.Host.ValueString(),
		Bucket:                    plan.Bucket.ValueString(),
		Region:                    plan.Region.ValueString(),
		CredentialAccessKeyID:     plan.AccessKeyID.ValueString(),
		CredentialSecretAccessKey: plan.SecretAccessKey.ValueString(),
	}

	storage, err := r.client.CreateStorage(storageParams)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating storage",
			fmt.Sprintf(
				"Error creating storage: %s",
				err.Error(),
			),
		)
		return
	}

	plan.ID = types.StringValue(strconv.Itoa(storage.Data.Attributes.ID))
	plan.CurrentState = types.StringValue(storage.Data.Attributes.CurrentState)
	plan.Host = types.StringValue(storage.Data.Attributes.Host)
	plan.Bucket = types.StringValue(storage.Data.Attributes.Bucket)
	plan.Region = types.StringValue(storage.Data.Attributes.Region)
	plan.AccessKeyID = types.StringValue(storage.Data.Attributes.CredentialAccessKeyID)
	plan.SecretAccessKey = types.StringValue(storage.Data.Attributes.CredentialSecretAccessKey)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *storageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state storageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	storage, err := r.client.GetStorage(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading storage",
			"cloud not read storage id "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.Host = types.StringValue(storage.Data.Attributes.Host)
	state.Bucket = types.StringValue(storage.Data.Attributes.Bucket)
	state.Region = types.StringValue(storage.Data.Attributes.Region)
	state.AccessKeyID = types.StringValue(storage.Data.Attributes.CredentialAccessKeyID)
	state.SecretAccessKey = types.StringValue(storage.Data.Attributes.CredentialSecretAccessKey)
	state.CurrentState = types.StringValue(storage.Data.Attributes.CurrentState)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *storageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan storageResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	storageParams := instc.StorageParams{
		Host:                      plan.Host.ValueString(),
		Bucket:                    plan.Bucket.ValueString(),
		Region:                    plan.Region.ValueString(),
		CredentialAccessKeyID:     plan.AccessKeyID.ValueString(),
		CredentialSecretAccessKey: plan.SecretAccessKey.ValueString(),
	}

	_, err := r.client.UpdateStorage(plan.ID.ValueString(), storageParams)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating storage",
			fmt.Sprintf(
				"Error updating storage: %s",
				err.Error(),
			),
		)
		return
	}

	storage, err := r.client.GetStorage(plan.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading storage",
			"Could not read storage ID "+plan.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	plan.Host = types.StringValue(storage.Data.Attributes.Host)
	plan.Bucket = types.StringValue(storage.Data.Attributes.Bucket)
	plan.Region = types.StringValue(storage.Data.Attributes.Region)
	plan.AccessKeyID = types.StringValue(storage.Data.Attributes.CredentialAccessKeyID)
	plan.SecretAccessKey = types.StringValue(storage.Data.Attributes.CredentialSecretAccessKey)
	plan.CurrentState = types.StringValue(storage.Data.Attributes.CurrentState)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *storageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state storageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteStorage(state.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting storage",
			fmt.Sprintf(
				"Error deleting storage: %s",
				err.Error(),
			),
		)
		return
	}
}

func (r *storageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
