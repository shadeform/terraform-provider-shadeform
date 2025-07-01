package instance

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
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
	Id                types.String   `tfsdk:"id"`
	Cloud             types.String   `tfsdk:"cloud"`
	Region            types.String   `tfsdk:"region"`
	ShadeInstanceType types.String   `tfsdk:"shade_instance_type"`
	ShadeCloud        types.Bool     `tfsdk:"shade_cloud"`
	Name              types.String   `tfsdk:"name"`
	Os                types.String   `tfsdk:"os"`
	SshKeyId          types.String   `tfsdk:"ssh_key_id"`
	TemplateId        types.String   `tfsdk:"template_id"`
	VolumeIds         types.List     `tfsdk:"volume_ids"`
	CloudInstanceType types.String   `tfsdk:"cloud_instance_type"`
	CloudAssignedID   types.String   `tfsdk:"cloud_assigned_id"`
	IP                types.String   `tfsdk:"ip"`
	SshUser           types.String   `tfsdk:"ssh_user"`
	SshPort           types.Int64    `tfsdk:"ssh_port"`
	Status            types.String   `tfsdk:"status"`
	CostEstimate      types.String   `tfsdk:"cost_estimate"`
	HourlyPrice       types.String   `tfsdk:"hourly_price"`
	CreatedAt         types.String   `tfsdk:"created_at"`
	Timeouts          timeouts.Value `tfsdk:"timeouts"`
}

func NewInstanceResource() resource.Resource {
	return &InstanceResource{}
}

func (r *InstanceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

func (r *InstanceResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"ssh_key_id": schema.StringAttribute{
				Description: "The ID of the SSH key to use for this instance.",
				Optional:    true,
				Computed:    true,
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
			"cloud_instance_type": schema.StringAttribute{
				Description: "The type of the instance in the cloud provider.",
				Computed:    true,
			},
			"cloud_assigned_id": schema.StringAttribute{
				Description: "The ID of the instance in the cloud provider.",
				Computed:    true,
			},
			"ip": schema.StringAttribute{
				Description: "The IP address of the instance.",
				Computed:    true,
			},
			"ssh_user": schema.StringAttribute{
				Description: "The user to use for SSH access to the instance.",
				Computed:    true,
			},
			"ssh_port": schema.Int64Attribute{
				Description: "The port to use for SSH access to the instance.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The status of the instance.",
				Computed:    true,
			},
			"cost_estimate": schema.StringAttribute{
				Description: "The cost estimate so far for the instance.",
				Computed:    true,
			},
			"hourly_price": schema.StringAttribute{
				Description: "The hourly price of the instance.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The date and time the instance was created.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
			}),
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

	if !plan.SshKeyId.IsNull() && !plan.SshKeyId.IsUnknown() {
		requestBody["ssh_key_id"] = plan.SshKeyId.ValueString()
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

	const defaultCreateTimeout = 60 * time.Minute

	createTimeout, diags := plan.Timeouts.Create(ctx, defaultCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	if err := pollInstanceStatus(ctx, r.client, plan.Name.ValueString(), instanceID, 15*time.Second); err != nil {
		resp.Diagnostics.AddError(
			"Instance not ready",
			fmt.Sprintf("timed out waiting for %s to become active: %s", instanceID, err),
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

	fmt.Println("instanceInfo")
	fmt.Println(instanceInfo)

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

	// For ssh_key_id, preserve the original plan value if it was set
	// Only override if it wasn't in the original plan
	if !plan.SshKeyId.IsNull() && !plan.SshKeyId.IsUnknown() {
		// Keep the original plan value
	} else if sshKeyId, ok := instanceInfo["ssh_key_id"].(string); ok {
		plan.SshKeyId = types.StringValue(sshKeyId)
	} else {
		plan.SshKeyId = types.StringNull()
	}
	if cloudInstanceType, ok := instanceInfo["cloud_instance_type"].(string); ok {
		plan.CloudInstanceType = types.StringValue(cloudInstanceType)
	} else {
		plan.CloudInstanceType = types.StringNull()
	}
	if cloudAssignedId, ok := instanceInfo["cloud_assigned_id"].(string); ok {
		plan.CloudAssignedID = types.StringValue(cloudAssignedId)
	} else {
		plan.CloudAssignedID = types.StringNull()
	}
	if ip, ok := instanceInfo["ip"].(string); ok {
		plan.IP = types.StringValue(ip)
	} else {
		plan.IP = types.StringNull()
	}
	if sshUser, ok := instanceInfo["ssh_user"].(string); ok {
		plan.SshUser = types.StringValue(sshUser)
	} else {
		plan.SshUser = types.StringNull()
	}
	if sshPort, ok := instanceInfo["ssh_port"].(int64); ok {
		plan.SshPort = types.Int64Value(sshPort)
	} else {
		plan.SshPort = types.Int64Null()
	}
	if status, ok := instanceInfo["status"].(string); ok {
		plan.Status = types.StringValue(status)
	} else {
		plan.Status = types.StringNull()
	}
	if costEstimate, ok := instanceInfo["cost_estimate"].(string); ok {
		plan.CostEstimate = types.StringValue(costEstimate)
	} else {
		plan.CostEstimate = types.StringNull()
	}
	if hourlyPrice, ok := instanceInfo["hourly_price"].(string); ok {
		plan.HourlyPrice = types.StringValue(hourlyPrice)
	} else {
		plan.HourlyPrice = types.StringNull()
	}
	if createdAt, ok := instanceInfo["created_at"].(string); ok {
		plan.CreatedAt = types.StringValue(createdAt)
	} else {
		plan.CreatedAt = types.StringNull()
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

	if sshKeyId, ok := result["ssh_key_id"].(string); ok {
		state.SshKeyId = types.StringValue(sshKeyId)
	} else {
		state.SshKeyId = types.StringNull()
	}
	if cloudInstanceType, ok := result["cloud_instance_type"].(string); ok {
		state.CloudInstanceType = types.StringValue(cloudInstanceType)
	} else {
		state.CloudInstanceType = types.StringNull()
	}
	if cloudAssignedId, ok := result["cloud_assigned_id"].(string); ok {
		state.CloudAssignedID = types.StringValue(cloudAssignedId)
	} else {
		state.CloudAssignedID = types.StringNull()
	}
	if ip, ok := result["ip"].(string); ok {
		state.IP = types.StringValue(ip)
	} else {
		state.IP = types.StringNull()
	}
	if sshUser, ok := result["ssh_user"].(string); ok {
		state.SshUser = types.StringValue(sshUser)
	} else {
		state.SshUser = types.StringNull()
	}
	if sshPort, ok := result["ssh_port"].(int64); ok {
		state.SshPort = types.Int64Value(sshPort)
	} else {
		state.SshPort = types.Int64Null()
	}
	if status, ok := result["status"].(string); ok {
		state.Status = types.StringValue(status)
	} else {
		state.Status = types.StringNull()
	}
	if costEstimate, ok := result["cost_estimate"].(string); ok {
		state.CostEstimate = types.StringValue(costEstimate)
	} else {
		state.CostEstimate = types.StringNull()
	}
	if hourlyPrice, ok := result["hourly_price"].(string); ok {
		state.HourlyPrice = types.StringValue(hourlyPrice)
	} else {
		state.HourlyPrice = types.StringNull()
	}
	if createdAt, ok := result["created_at"].(string); ok {
		state.CreatedAt = types.StringValue(createdAt)
	} else {
		state.CreatedAt = types.StringNull()
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

// pollInstanceStatus blocks until the instance reaches the wanted status or
// the ctx deadline is hit.
func pollInstanceStatus(
	ctx context.Context,
	c *provider_shadeform.Client,
	name string,
	id string,
	interval time.Duration,
) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err() // timeout or user ^C
		case <-ticker.C:
			info, err := c.GetInstance(id)
			if err != nil {
				return err // API error – abort
			}

			status, _ := info["status"].(string)
			tflog.Debug(ctx, fmt.Sprintf("instance [name: %s, id: %s] status=%s", name, id, status))

			if status == "active" {
				return nil // success
			} else if status == "error" {
				return fmt.Errorf("instance %s is in error state", id)
			}
		}
	}
}
