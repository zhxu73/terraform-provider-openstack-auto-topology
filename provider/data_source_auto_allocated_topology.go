package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/cyverse/openstack-auto-allocated-topology/openstack"
)

func dataSourceAutoAllocatedTopology() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get the auto allocated topology of current project",
		ReadContext: dataSourceAutoAllocatedTopologyRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "network ID of the auto allocated topology",
			},
			"project_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "project ID of the auto allocated topology",
			},
		},
	}
}

func dataSourceAutoAllocatedTopologyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	// from return value of providerConfigure()
	osClient := m.(openstack.Client)

	networkClient, err := osClient.Network()
	if err != nil {
		return diag.FromErr(err)
	}
	projectID, _ := osClient.CurrentProject()
	topology, err := networkClient.GetAutoAllocatedTopology(projectID)
	if err != nil {
		return diag.FromErr(err)
	}
	if topology == nil {
		panic("topology is nil")
	}

	err = d.Set("id", topology.NetworkID)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("project_id", topology.ProjectID)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(topology.NetworkID)

	return diags
}
