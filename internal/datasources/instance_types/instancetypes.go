package instance_types

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/shadeform/terraform-provider-shadeform/internal/provider/provider_shadeform"
)

var (
	_ datasource.DataSource = &InstanceTypesDataSource{}
)

type InstanceTypesDataSource struct {
	client *provider_shadeform.Client
}

type InstanceTypesDataSourceModel struct {
	Cloud             types.String `tfsdk:"cloud"`
	Region            types.String `tfsdk:"region"`
	NumGpus           types.String `tfsdk:"num_gpus"`
	GpuType           types.String `tfsdk:"gpu_type"`
	ShadeInstanceType types.String `tfsdk:"shade_instance_type"`
	Available         types.Bool   `tfsdk:"available"`
	Sort              types.String `tfsdk:"sort"`
	InstanceTypes     types.List   `tfsdk:"instance_types"`
}

type InstanceTypeModel struct {
	Cloud             types.String `tfsdk:"cloud"`
	Region            types.String `tfsdk:"region"`
	ShadeInstanceType types.String `tfsdk:"shade_instance_type"`
	CloudInstanceType types.String `tfsdk:"cloud_instance_type"`
	HourlyPrice       types.Int64  `tfsdk:"hourly_price"`
	DeploymentType    types.String `tfsdk:"deployment_type"`
	OsOptions         types.List   `tfsdk:"os_options"`
	Availability      types.List   `tfsdk:"availability"`
	BootTime          types.Object `tfsdk:"boot_time"`
}

type AvailabilityModel struct {
	Region      types.String `tfsdk:"region"`
	Available   types.Bool   `tfsdk:"available"`
	DisplayName types.String `tfsdk:"display_name"`
}

type BootTimeModel struct {
	MinBootInSec types.Int64 `tfsdk:"min_boot_in_sec"`
	MaxBootInSec types.Int64 `tfsdk:"max_boot_in_sec"`
}

func NewInstanceTypesDataSource() datasource.DataSource {
	return &InstanceTypesDataSource{}
}

// Metadata returns the data source type name.
func (d *InstanceTypesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance_types"
}

// Schema defines the schema for the data source.
func (d *InstanceTypesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieve available instance types from Shadeform.",
		Attributes: map[string]schema.Attribute{
			"cloud": schema.StringAttribute{
				Description: "Filter the instance type results by cloud.",
				Optional:    true,
			},
			"region": schema.StringAttribute{
				Description: "Filter the instance type results by region.",
				Optional:    true,
			},
			"num_gpus": schema.StringAttribute{
				Description: "Filter the instance type results by the number of gpus.",
				Optional:    true,
			},
			"gpu_type": schema.StringAttribute{
				Description: "Filter the instance type results by gpu type.",
				Optional:    true,
			},
			"shade_instance_type": schema.StringAttribute{
				Description: "Filter the instance type results by the shade instance type.",
				Optional:    true,
			},
			"available": schema.BoolAttribute{
				Description: "Filter the instance type results by availability.",
				Optional:    true,
			},
			"sort": schema.StringAttribute{
				Description: "Sort the order of the instance type results.",
				Optional:    true,
			},
			"instance_types": schema.ListAttribute{
				Description: "List of available instance types.",
				Computed:    true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"cloud":               types.StringType,
						"region":              types.StringType,
						"shade_instance_type": types.StringType,
						"cloud_instance_type": types.StringType,
						"hourly_price":        types.Int64Type,
						"deployment_type":     types.StringType,
						"os_options":          types.ListType{ElemType: types.StringType},
						"availability": types.ListType{ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"region":       types.StringType,
								"available":    types.BoolType,
								"display_name": types.StringType,
							},
						}},
						"boot_time": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"min_boot_in_sec": types.Int64Type,
								"max_boot_in_sec": types.Int64Type,
							},
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *InstanceTypesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*provider_shadeform.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *provider_shadeform.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

// Read refreshes the Terraform state with the latest data.
func (d *InstanceTypesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data InstanceTypesDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Build query parameters
	params := make(map[string]string)
	if !data.Cloud.IsNull() && !data.Cloud.IsUnknown() {
		params["cloud"] = data.Cloud.ValueString()
	}
	if !data.Region.IsNull() && !data.Region.IsUnknown() {
		params["region"] = data.Region.ValueString()
	}
	if !data.NumGpus.IsNull() && !data.NumGpus.IsUnknown() {
		params["num_gpus"] = data.NumGpus.ValueString()
	}
	if !data.GpuType.IsNull() && !data.GpuType.IsUnknown() {
		params["gpu_type"] = data.GpuType.ValueString()
	}
	if !data.ShadeInstanceType.IsNull() && !data.ShadeInstanceType.IsUnknown() {
		params["shade_instance_type"] = data.ShadeInstanceType.ValueString()
	}
	if !data.Available.IsNull() && !data.Available.IsUnknown() {
		params["available"] = fmt.Sprintf("%t", data.Available.ValueBool())
	}
	if !data.Sort.IsNull() && !data.Sort.IsUnknown() {
		params["sort"] = data.Sort.ValueString()
	}

	// Get instance types from API
	result, err := d.client.GetInstanceTypes(params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading instance types",
			"Could not read instance types, unexpected error: "+err.Error(),
		)
		return
	}

	// Parse the response
	instanceTypesRaw, ok := result["instance_types"]
	if !ok {
		resp.Diagnostics.AddError(
			"Error reading instance types",
			"Response does not contain instance_types field",
		)
		return
	}

	instanceTypesArray, ok := instanceTypesRaw.([]interface{})
	if !ok {
		resp.Diagnostics.AddError(
			"Error reading instance types",
			"instance_types field is not an array",
		)
		return
	}

	// Convert to Terraform types
	var instanceTypes []attr.Value
	for _, instanceTypeRaw := range instanceTypesArray {
		instanceTypeMap, ok := instanceTypeRaw.(map[string]interface{})
		if !ok {
			continue
		}

		instanceType := InstanceTypeModel{}

		// Parse basic fields
		if cloud, ok := instanceTypeMap["cloud"].(string); ok {
			instanceType.Cloud = types.StringValue(cloud)
		}
		if region, ok := instanceTypeMap["region"].(string); ok {
			instanceType.Region = types.StringValue(region)
		}
		if shadeInstanceType, ok := instanceTypeMap["shade_instance_type"].(string); ok {
			instanceType.ShadeInstanceType = types.StringValue(shadeInstanceType)
		}
		if cloudInstanceType, ok := instanceTypeMap["cloud_instance_type"].(string); ok {
			instanceType.CloudInstanceType = types.StringValue(cloudInstanceType)
		}
		if hourlyPrice, ok := instanceTypeMap["hourly_price"].(float64); ok {
			instanceType.HourlyPrice = types.Int64Value(int64(hourlyPrice))
		}
		if deploymentType, ok := instanceTypeMap["deployment_type"].(string); ok {
			instanceType.DeploymentType = types.StringValue(deploymentType)
		}

		// Parse OS options
		if config, ok := instanceTypeMap["configuration"].(map[string]interface{}); ok {
			if osOptionsRaw, ok := config["os_options"].([]interface{}); ok {
				var osOptions []attr.Value
				for _, osOption := range osOptionsRaw {
					if osOptionStr, ok := osOption.(string); ok {
						osOptions = append(osOptions, types.StringValue(osOptionStr))
					}
				}
				if len(osOptions) > 0 {
					instanceType.OsOptions = types.ListValueMust(types.StringType, osOptions)
				} else {
					instanceType.OsOptions = types.ListNull(types.StringType)
				}
			}
		}

		// Parse availability
		if availabilityRaw, ok := instanceTypeMap["availability"].([]interface{}); ok {
			var availability []attr.Value
			for _, availRaw := range availabilityRaw {
				if availMap, ok := availRaw.(map[string]interface{}); ok {
					avail := AvailabilityModel{}
					if region, ok := availMap["region"].(string); ok {
						avail.Region = types.StringValue(region)
					}
					if available, ok := availMap["available"].(bool); ok {
						avail.Available = types.BoolValue(available)
					}
					if displayName, ok := availMap["display_name"].(string); ok {
						avail.DisplayName = types.StringValue(displayName)
					}
					availability = append(availability, types.ObjectValueMust(
						map[string]attr.Type{
							"region":       types.StringType,
							"available":    types.BoolType,
							"display_name": types.StringType,
						},
						map[string]attr.Value{
							"region":       avail.Region,
							"available":    avail.Available,
							"display_name": avail.DisplayName,
						},
					))
				}
			}
			if len(availability) > 0 {
				instanceType.Availability = types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"region":       types.StringType,
							"available":    types.BoolType,
							"display_name": types.StringType,
						},
					},
					availability,
				)
			} else {
				instanceType.Availability = types.ListNull(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"region":       types.StringType,
							"available":    types.BoolType,
							"display_name": types.StringType,
						},
					},
				)
			}
		}

		// Parse boot time
		if bootTimeRaw, ok := instanceTypeMap["boot_time"].(map[string]interface{}); ok {
			bootTime := BootTimeModel{}
			if minBoot, ok := bootTimeRaw["min_boot_in_sec"].(float64); ok {
				bootTime.MinBootInSec = types.Int64Value(int64(minBoot))
			}
			if maxBoot, ok := bootTimeRaw["max_boot_in_sec"].(float64); ok {
				bootTime.MaxBootInSec = types.Int64Value(int64(maxBoot))
			}
			instanceType.BootTime = types.ObjectValueMust(
				map[string]attr.Type{
					"min_boot_in_sec": types.Int64Type,
					"max_boot_in_sec": types.Int64Type,
				},
				map[string]attr.Value{
					"min_boot_in_sec": bootTime.MinBootInSec,
					"max_boot_in_sec": bootTime.MaxBootInSec,
				},
			)
		}

		// Convert to ObjectValue
		instanceTypeObj := types.ObjectValueMust(
			map[string]attr.Type{
				"cloud":               types.StringType,
				"region":              types.StringType,
				"shade_instance_type": types.StringType,
				"cloud_instance_type": types.StringType,
				"hourly_price":        types.Int64Type,
				"deployment_type":     types.StringType,
				"os_options":          types.ListType{ElemType: types.StringType},
				"availability": types.ListType{ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"region":       types.StringType,
						"available":    types.BoolType,
						"display_name": types.StringType,
					},
				}},
				"boot_time": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"min_boot_in_sec": types.Int64Type,
						"max_boot_in_sec": types.Int64Type,
					},
				},
			},
			map[string]attr.Value{
				"cloud":               instanceType.Cloud,
				"region":              instanceType.Region,
				"shade_instance_type": instanceType.ShadeInstanceType,
				"cloud_instance_type": instanceType.CloudInstanceType,
				"hourly_price":        instanceType.HourlyPrice,
				"deployment_type":     instanceType.DeploymentType,
				"os_options":          instanceType.OsOptions,
				"availability":        instanceType.Availability,
				"boot_time":           instanceType.BootTime,
			},
		)

		instanceTypes = append(instanceTypes, instanceTypeObj)
	}

	// Set the instance types
	data.InstanceTypes = types.ListValueMust(
		types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"cloud":               types.StringType,
				"region":              types.StringType,
				"shade_instance_type": types.StringType,
				"cloud_instance_type": types.StringType,
				"hourly_price":        types.Int64Type,
				"deployment_type":     types.StringType,
				"os_options":          types.ListType{ElemType: types.StringType},
				"availability": types.ListType{ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"region":       types.StringType,
						"available":    types.BoolType,
						"display_name": types.StringType,
					},
				}},
				"boot_time": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"min_boot_in_sec": types.Int64Type,
						"max_boot_in_sec": types.Int64Type,
					},
				},
			},
		},
		instanceTypes,
	)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
