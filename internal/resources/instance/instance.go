package instance

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/shadeform/terraform-provider-shadeform/internal/provider/provider_shadeform"
)

var (
	_ resource.Resource                = &InstanceResource{}
	_ resource.ResourceWithConfigure   = &InstanceResource{}
	_ resource.ResourceWithImportState = &InstanceResource{}
)

type InstanceResource struct {
	client *provider_shadeform.Client
}

type InstanceResourceModel struct {
	Id                types.String `tfsdk:"id"`
	Cloud             types.String `tfsdk:"cloud"`
	Region            types.String `tfsdk:"region"`
	ShadeInstanceType types.String `tfsdk:"shade_instance_type"`
	ShadeCloud        types.Bool   `tfsdk:"shade_cloud"`
	Name              types.String `tfsdk:"name"`
	Os                types.String `tfsdk:"os"`
	TemplateId        types.String `tfsdk:"template_id"`
	VolumeIds         types.List   `tfsdk:"volume_ids"`
}

func NewInstanceResource() resource.Resource {
	return &InstanceResource{}
}

func (r *InstanceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

func (r *InstanceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Shadeform instance.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the instance.",
				Computed:    true,
			},
			"cloud": schema.StringAttribute{
				Description: "The cloud provider.",
				Required:    true,
			},
			"region": schema.StringAttribute{
				Description: "The region where the instance will be deployed.",
				Required:    true,
			},
			"shade_instance_type": schema.StringAttribute{
				Description: "The Shadeform standardized instance type.",
				Required:    true,
			},
			"shade_cloud": schema.BoolAttribute{
				Description: "Whether to use Shade Cloud or linked cloud account. This is usually true.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the instance.",
				Required:    true,
			},
			"os": schema.StringAttribute{
				Description: "The operating system of the instance.",
				Optional:    true,
			},
			"template_id": schema.StringAttribute{
				Description: "The ID of the template to use for this instance.",
				Optional:    true,
			},
			"volume_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of volume IDs to be mounted. Currently only supports 1 volume at a time.",
				Optional:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *InstanceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *InstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan InstanceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build request body
	requestBody := map[string]interface{}{
		"cloud":               plan.Cloud.ValueString(),
		"region":              plan.Region.ValueString(),
		"shade_instance_type": plan.ShadeInstanceType.ValueString(),
		"shade_cloud":         plan.ShadeCloud.ValueBool(),
		"name":                plan.Name.ValueString(),
	}

	if !plan.Os.IsNull() && !plan.Os.IsUnknown() {
		requestBody["os"] = plan.Os.ValueString()
	}

	if !plan.TemplateId.IsNull() && !plan.TemplateId.IsUnknown() {
		requestBody["template_id"] = plan.TemplateId.ValueString()
	}

	// Add volume_ids if specified
	if !plan.VolumeIds.IsNull() && !plan.VolumeIds.IsUnknown() {
		var volumeIds []string
		diags := plan.VolumeIds.ElementsAs(ctx, &volumeIds, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		requestBody["volume_ids"] = volumeIds
	}

	// Create instance
	result, err := r.client.CreateInstance(requestBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating instance",
			"Could not create instance, unexpected error: "+err.Error(),
		)
		return
	}

	// Extract instance ID from response
	instanceID, ok := result["id"].(string)
	if !ok {
		resp.Diagnostics.AddError(
			"Error creating instance",
			"Could not extract instance ID from response",
		)
		return
	}

	// Now fetch the full instance info to populate all computed fields
	instanceInfo, err := r.client.GetInstance(instanceID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading instance after create",
			"Could not read instance, unexpected error: "+err.Error(),
		)
		return
	}

	// Set all fields from the API response
	plan.Id = types.StringValue(instanceID)
	if cloud, ok := instanceInfo["cloud"].(string); ok {
		plan.Cloud = types.StringValue(cloud)
	}
	if region, ok := instanceInfo["region"].(string); ok {
		plan.Region = types.StringValue(region)
	}
	if shadeInstanceType, ok := instanceInfo["shade_instance_type"].(string); ok {
		plan.ShadeInstanceType = types.StringValue(shadeInstanceType)
	}
	if shadeCloud, ok := instanceInfo["shade_cloud"].(bool); ok {
		plan.ShadeCloud = types.BoolValue(shadeCloud)
	}
	if name, ok := instanceInfo["name"].(string); ok {
		plan.Name = types.StringValue(name)
	}

	// Handle optional fields that might be returned by the API
	if os, ok := instanceInfo["os"].(string); ok {
		plan.Os = types.StringValue(os)
	} else {
		plan.Os = types.StringNull()
	}

	if templateId, ok := instanceInfo["template_id"].(string); ok {
		plan.TemplateId = types.StringValue(templateId)
	} else {
		plan.TemplateId = types.StringNull()
	}

	// Handle volume_ids - it's a list in the API response
	if volumeIdsRaw, ok := instanceInfo["volume_ids"]; ok && volumeIdsRaw != nil {
		if volumeIdsArray, ok := volumeIdsRaw.([]interface{}); ok {
			var volumeIds []attr.Value
			for _, v := range volumeIdsArray {
				if vStr, ok := v.(string); ok {
					volumeIds = append(volumeIds, types.StringValue(vStr))
				}
			}
			if len(volumeIds) > 0 {
				plan.VolumeIds = types.ListValueMust(types.StringType, volumeIds)
			} else {
				plan.VolumeIds = types.ListNull(types.StringType)
			}
		} else {
			plan.VolumeIds = types.ListNull(types.StringType)
		}
	} else {
		plan.VolumeIds = types.ListNull(types.StringType)
	}

	// Set state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *InstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state InstanceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get instance from API
	result, err := r.client.GetInstance(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading instance",
			"Could not read instance, unexpected error: "+err.Error(),
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
	if shadeInstanceType, ok := result["shade_instance_type"].(string); ok {
		state.ShadeInstanceType = types.StringValue(shadeInstanceType)
	}
	if shadeCloud, ok := result["shade_cloud"].(bool); ok {
		state.ShadeCloud = types.BoolValue(shadeCloud)
	}
	if name, ok := result["name"].(string); ok {
		state.Name = types.StringValue(name)
	}

	// Handle optional fields that might be returned by the API
	if os, ok := result["os"].(string); ok {
		state.Os = types.StringValue(os)
	} else {
		state.Os = types.StringNull()
	}

	if templateId, ok := result["template_id"].(string); ok {
		state.TemplateId = types.StringValue(templateId)
	} else {
		state.TemplateId = types.StringNull()
	}

	// Handle volume_ids - it's a list in the API response
	if volumeIdsRaw, ok := result["volume_ids"]; ok && volumeIdsRaw != nil {
		if volumeIdsArray, ok := volumeIdsRaw.([]interface{}); ok {
			var volumeIds []attr.Value
			for _, v := range volumeIdsArray {
				if vStr, ok := v.(string); ok {
					volumeIds = append(volumeIds, types.StringValue(vStr))
				}
			}
			if len(volumeIds) > 0 {
				state.VolumeIds = types.ListValueMust(types.StringType, volumeIds)
			} else {
				state.VolumeIds = types.ListNull(types.StringType)
			}
		} else {
			state.VolumeIds = types.ListNull(types.StringType)
		}
	} else {
		state.VolumeIds = types.ListNull(types.StringType)
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *InstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and state
	var plan InstanceResourceModel
	var state InstanceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build request body for update
	requestBody := make(map[string]interface{})

	if !plan.Name.Equal(state.Name) {
		requestBody["name"] = plan.Name.ValueString()
	}

	// Update instance
	err := r.client.UpdateInstance(state.Id.ValueString(), requestBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating instance",
			"Could not update instance, unexpected error: "+err.Error(),
		)
		return
	}

	// Fetch the updated instance data to ensure all computed fields are set
	instanceInfo, err := r.client.GetInstance(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading instance after update",
			"Could not read instance, unexpected error: "+err.Error(),
		)
		return
	}

	// Update the plan with the fetched data
	plan.Id = types.StringValue(state.Id.ValueString())
	if cloud, ok := instanceInfo["cloud"].(string); ok {
		plan.Cloud = types.StringValue(cloud)
	}
	if region, ok := instanceInfo["region"].(string); ok {
		plan.Region = types.StringValue(region)
	}
	if shadeInstanceType, ok := instanceInfo["shade_instance_type"].(string); ok {
		plan.ShadeInstanceType = types.StringValue(shadeInstanceType)
	}
	if shadeCloud, ok := instanceInfo["shade_cloud"].(bool); ok {
		plan.ShadeCloud = types.BoolValue(shadeCloud)
	}
	if name, ok := instanceInfo["name"].(string); ok {
		plan.Name = types.StringValue(name)
	}

	// Handle optional fields that might be returned by the API
	if os, ok := instanceInfo["os"].(string); ok {
		plan.Os = types.StringValue(os)
	} else {
		plan.Os = types.StringNull()
	}

	if templateId, ok := instanceInfo["template_id"].(string); ok {
		plan.TemplateId = types.StringValue(templateId)
	} else {
		plan.TemplateId = types.StringNull()
	}

	// Handle volume_ids - it's a list in the API response
	if volumeIdsRaw, ok := instanceInfo["volume_ids"]; ok && volumeIdsRaw != nil {
		if volumeIdsArray, ok := volumeIdsRaw.([]interface{}); ok {
			var volumeIds []attr.Value
			for _, v := range volumeIdsArray {
				if vStr, ok := v.(string); ok {
					volumeIds = append(volumeIds, types.StringValue(vStr))
				}
			}
			if len(volumeIds) > 0 {
				plan.VolumeIds = types.ListValueMust(types.StringType, volumeIds)
			} else {
				plan.VolumeIds = types.ListNull(types.StringType)
			}
		} else {
			plan.VolumeIds = types.ListNull(types.StringType)
		}
	} else {
		plan.VolumeIds = types.ListNull(types.StringType)
	}

	// Set state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *InstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get state
	var state InstanceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete instance
	err := r.client.DeleteInstance(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting instance",
			"Could not delete instance, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *InstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
