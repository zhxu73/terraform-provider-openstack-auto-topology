terraform {
  required_providers {
    openstack-auto-topology = {
      source  = "terraform.cyverse.org/cyverse/openstack-auto-topology"
    }
    openstack = {
      source = "terraform.cyverse.org/cyverse/openstack"
    }
  }
}

provider "openstack-auto-topology" {
    # auth_url = "https://cyverse.org"
}

data "openstack-auto-topology_auto_allocated_topology" "network" {
    # region_name = "MY_REGION" # you can override the region name from application credential
}

output "network_id" {
  value = data.openstack-auto-topology_auto_allocated_topology.network.id
}
