terraform {
  required_providers {
    openstack-auto-topology = {
      #source  = "terraform.cyverse.org/cyverse/openstack-auto-topology" # use this source if you install this provider locally using the Makefile
      source = "zhxu73/openstack-auto-topology"
      version = "0.1.0"
    }
    openstack = {
      source = "terraform-provider-openstack/openstack"
    }
  }
}

provider "openstack-auto-topology" {
    # auth_url = "https://cyverse.org"
}

data "openstack-auto-topology_auto_allocated_topology" "network" {
    # region_name = "MY_REGION" # you can override the region name from application credential
    # project_id = "MY_PROJECT_ID" # you can override the project ID, project ID takes priority over project name
    # project_name = "MY_PROJECT_NAME" # you can override the project name
}

output "network_id" {
  value = data.openstack-auto-topology_auto_allocated_topology.network.id
}

output "network_name" {
  value = data.openstack-auto-topology_auto_allocated_topology.network.name
}
