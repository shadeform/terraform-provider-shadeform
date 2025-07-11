---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "shadeform_volume Resource - terraform-provider-shadeform"
subcategory: ""
description: |-
  Manages a Shadeform storage volume.
---

# shadeform_volume (Resource)

Manages a Shadeform storage volume. Volumes provide persistent block storage that can be attached to instances for storing data independently of the instance lifecycle.
For more information, see our [public docs](https://docs.shadeform.ai/guides/attachvolume).

## Example Usage

```terraform
terraform {
  required_providers {
    shadeform = {
      source = "shadeform/shadeform"
    }
  }
}

provider "shadeform" {
  api_key = "YOUR_API_KEY"
}

# Create a volume before you create an instance
resource "shadeform_volume" "test-volume" {
  cloud      = "datacrunch"
  region     = "helsinki-finland-2"
  name       = "terraform-test-volume"
  size_in_gb = 101
}

# Now you can attach the volume an instance during its creation
resource "shadeform_instance" "test-instance" {
  cloud = "datacrunch"
  region = "helsinki-finland-2"
  shade_instance_type = "H200"
  name = "terraform-test-instance"
  volume_ids = [shadeform_volume.test-volume.id]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cloud` (String) The cloud provider.
- `name` (String) The name of the volume.
- `region` (String) The region where the volume will be created.
- `size_in_gb` (Number) The size of the volume in gigabytes.

### Read-Only

- `cost_estimate` (String) The cost estimate for the volume.
- `fixed_size` (Boolean) Whether the volume is fixed in size or elastically scaling.
- `id` (String) The unique identifier for the volume.
- `mounted_by` (String) The ID of the instance that is currently mounting the volume.
- `supports_multi_mount` (Boolean) Whether the volume supports multiple instances mounting to it.
