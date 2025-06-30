package volume

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/shadeform/terraform-provider-shadeform/internal/provider/provider_shadeform"
)

var (
	_ resource.Resource                = &VolumeResource{}
	_ resource.ResourceWithConfigure   = &VolumeResource{}
	_ resource.ResourceWithImportState = &VolumeResource{}
)

type VolumeResource struct {
	client *provider_shadeform.Client
}

type VolumeResourceModel struct {
	Id                 types.String `tfsdk:"id"`
	Cloud              types.String `tfsdk:"cloud"`
	Region             types.String `tfsdk:"region"`
	Name               types.String `tfsdk:"name"`
	SizeInGb           types.Int64  `tfsdk:"size_in_gb"`
	FixedSize          types.Bool   `tfsdk:"fixed_size"`
	SupportsMultiMount types.Bool   `tfsdk:"supports_multi_mount"`
	CostEstimate       types.String `tfsdk:"cost_estimate"`
	MountedBy          types.String `tfsdk:"mounted_by"`
}

func NewVolumeResource() resource.Resource {
	return &VolumeResource{}
}

// Metadata returns the resource type name.
func (r *VolumeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume"
}

// Schema defines the schema for the resource.
func (r *VolumeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Shadeform storage volume.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the volume.",
				Computed:    true,
			},
			"cloud": schema.StringAttribute{
				Description: "The cloud provider.",
				Required:    true,
			},
			"region": schema.StringAttribute{
				Description: "The region where the volume will be created.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the volume.",
				Required:    true,
			},
			"size_in_gb": schema.Int64Attribute{
				Description: "The size of the volume in gigabytes.",
				Required:    true,
			},
			"fixed_size": schema.BoolAttribute{
				Description: "Whether the volume is fixed in size or elastically scaling.",
				Computed:    true,
			},
			"supports_multi_mount": schema.BoolAttribute{
				Description: "Whether the volume supports multiple instances mounting to it.",
				Computed:    true,
			},
			"cost_estimate": schema.StringAttribute{
				Description: "The cost estimate for the volume.",
				Computed:    true,
			},
			"mounted_by": schema.StringAttribute{
				Description: "The ID of the instance that is currently mounting the volume.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *VolumeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*provider_shadeform.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *provider_shadeform.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *VolumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan VolumeResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build request body
	requestBody := map[string]interface{}{
		"cloud":      plan.Cloud.ValueString(),
		"region":     plan.Region.ValueString(),
		"name":       plan.Name.ValueString(),
		"size_in_gb": plan.SizeInGb.ValueInt64(),
	}

	// Create volume
	result, err := r.client.CreateVolume(requestBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating volume",
			"Could not create volume, unexpected error: "+err.Error(),
		)
		return
	}

	// Extract volume ID from response
	volumeID, ok := result["id"].(string)
	if !ok {
		resp.Diagnostics.AddError(
			"Error creating volume",
			"Could not extract volume ID from response",
		)
		return
	}

	// Now fetch the full volume info to populate all computed fields
	volumeInfo, err := r.client.GetVolume(volumeID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading volume after create",
			"Could not read volume, unexpected error: "+err.Error(),
		)
		return
	}

	// Set all fields from the API response
	plan.Id = types.StringValue(volumeID)
	if cloud, ok := volumeInfo["cloud"].(string); ok {
		plan.Cloud = types.StringValue(cloud)
	}
	if region, ok := volumeInfo["region"].(string); ok {
		plan.Region = types.StringValue(region)
	}
	if name, ok := volumeInfo["name"].(string); ok {
		plan.Name = types.StringValue(name)
	}
	if sizeInGb, ok := volumeInfo["size_in_gb"].(float64); ok {
		plan.SizeInGb = types.Int64Value(int64(sizeInGb))
	}
	if fixedSize, ok := volumeInfo["fixed_size"].(bool); ok {
		plan.FixedSize = types.BoolValue(fixedSize)
	}
	if supportsMultiMount, ok := volumeInfo["supports_multi_mount"].(bool); ok {
		plan.SupportsMultiMount = types.BoolValue(supportsMultiMount)
	}
	if costEstimate, ok := volumeInfo["cost_estimate"].(string); ok {
		plan.CostEstimate = types.StringValue(costEstimate)
	}

	// Handle mounted_by - it can be null when not mounted
	if mountedBy, ok := volumeInfo["mounted_by"]; ok && mountedBy != nil {
		if mountedByStr, ok := mountedBy.(string); ok {
			plan.MountedBy = types.StringValue(mountedByStr)
		} else {
			plan.MountedBy = types.StringNull()
		}
	} else {
		plan.MountedBy = types.StringNull()
	}

	// Set state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *VolumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state VolumeResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get volume from API
	result, err := r.client.GetVolume(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading volume",
			"Could not read volume, unexpected error: "+err.Error(),
		)
		return
	}

	// Update state with API response
	if cloud, ok := result["cloud"].(string); ok {
		state.Cloud = types.StringValue(cloud)
	}
	if region, ok := result["region"].(string); ok {
		state.Region = types.StringValue(region)
	}
	if name, ok := result["name"].(string); ok {
		state.Name = types.StringValue(name)
	}
	if sizeInGb, ok := result["size_in_gb"].(float64); ok {
		state.SizeInGb = types.Int64Value(int64(sizeInGb))
	}
	if fixedSize, ok := result["fixed_size"].(bool); ok {
		state.FixedSize = types.BoolValue(fixedSize)
	}
	if supportsMultiMount, ok := result["supports_multi_mount"].(bool); ok {
		state.SupportsMultiMount = types.BoolValue(supportsMultiMount)
	}
	if costEstimate, ok := result["cost_estimate"].(string); ok {
		state.CostEstimate = types.StringValue(costEstimate)
	}

	// Handle mounted_by - it can be null when not mounted
	if mountedBy, ok := result["mounted_by"]; ok && mountedBy != nil {
		if mountedByStr, ok := mountedBy.(string); ok {
			state.MountedBy = types.StringValue(mountedByStr)
		} else {
			state.MountedBy = types.StringNull()
		}
	} else {
		state.MountedBy = types.StringNull()
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *VolumeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Volumes don't support updates in the API, so we'll just return the current state
	var plan VolumeResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state (no changes since volumes can't be updated)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *VolumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get state
	var state VolumeResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if volume is mounted before attempting to delete
	volumeInfo, err := r.client.GetVolume(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading volume before delete",
			"Could not read volume, unexpected error: "+err.Error(),
		)
		return
	}

	// Check if volume is mounted
	if mountedBy, ok := volumeInfo["mounted_by"]; ok && mountedBy != nil {
		if mountedByStr, ok := mountedBy.(string); ok && mountedByStr != "" {
			resp.Diagnostics.AddError(
				"Error deleting volume",
				fmt.Sprintf("Cannot delete volume %s because it is mounted by instance %s. Please delete the instance first.", state.Id.ValueString(), mountedByStr),
			)
			return
		}
	}

	// Delete volume
	err = r.client.DeleteVolume(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting volume",
			"Could not delete volume, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *VolumeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by volume ID
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
