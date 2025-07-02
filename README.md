# Terraform Provider for Shadeform

[![Terraform Registry](https://img.shields.io/badge/terraform-registry-blue.svg)](https://registry.terraform.io/providers/shadeform/shadeform)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Terraform provider for managing GPU instances and storage volumes on Shadeform, a unified platform for deploying and managing cloud GPUs across multiple cloud providers.

## Quick Start

### Installation

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    shadeform = {
      source  = "shadeform/shadeform"
      version = "~> 0.1.0"
    }
  }
}
```

### Configuration

Configure the provider with your Shadeform API key:

```hcl
provider "shadeform" {
  api_key = "YOU_API_KEY"
}
```

You can also set the API key via environment variable:

```bash
export SHADEFORM_API_KEY="YOU_API_KEY"
```

If you do that, make sure to set the `api_key` to `""`

### Basic Usage

Create a GPU instance:

```hcl
resource "shadeform_instance" "test-instance" {
  cloud              = "scaleway"
  region             = "paris-france-1"
  shade_instance_type = "H100"
  name               = "terraform-test-instance"
} 
```

Create a persistent storage volume:

```hcl
resource "shadeform_volume" "test-volume" {
  cloud      = "datacrunch"
  region     = "helsinki-finland-2"
  name       = "terraform-test-volume"
  size_in_gb = 101
}
```

> **_NOTE:_** Instances can take anywhere from 1 - 15 minutes on average to spin up with some evening taking upwards of 30-40 minutes.
The `terraform apply` command won't finish until the instances are active (or errored out)

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.0 |
| go | >= 1.21 |

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: [Shadeform API Documentation](https://docs.shadeform.ai)
- **Issues**: [GitHub Issues](https://github.com/shadeform/terraform-provider-shadeform/issues)
- **Email**: support@shadeform.ai

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a list of changes and version history. 