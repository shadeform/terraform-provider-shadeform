# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of the Shadeform Terraform provider
- Support for GPU instance management (create, read, update, delete)
- Support for storage volume management (create, read, delete)
- Data source for querying available instance types
- Volume attachment to instances

### Resources
- `shadeform_instance` - Manage GPU instances
- `shadeform_volume` - Manage storage volumes

### Data Sources
- `shadeform_instance_types` - Query available instance types and availability

### Features
- Full CRUD operations for instances and volumes
- Availability checking before deployment
- Volume attachment and management
