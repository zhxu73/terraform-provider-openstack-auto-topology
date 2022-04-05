# terraform-provider-openstack-auto-topology

A simple Terraform provider that fetches the auto allocated topology for a OpenStack project

# Summary

This provider accepts openstack application credential via environment variables. The data source will fetches (or create if absent) the auto allocated topology for the current project (project of the application credential).

# Build
```bash
make build
```

or

```bash
go build
```

# Install to home directory
```bash
make install
```

# Example
see `example/main.tf` for example usage of the provider.

the `example/main.tf` assumes that you have the provider installed to your home directory.

