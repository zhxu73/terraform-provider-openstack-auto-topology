package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/cyverse/openstack-auto-allocated-topology/openstack"
)

const (
	topologyIDAttribute  = "id"
	projectIDAttribute   = "project_id"
	projectNameAttribute = "project_name"
	regionNameAttribute  = "region_name"
)

func dataSourceAutoAllocatedTopology() *schema.Resource {
	return &schema.Resource{
		Description:   "Use this data source to get the auto allocated topology of current project",
		ReadContext:   dataSourceAutoAllocatedTopologyRead,
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			topologyIDAttribute: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "network ID of the auto allocated topology",
			},
			projectIDAttribute: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "project ID of the auto allocated topology",
			},
			projectNameAttribute: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "project name of the auto allocated topology",
			},
			regionNameAttribute: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "region name of the auto allocated topology",
			},
		},
	}
}

func dataSourceAutoAllocatedTopologyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	// from return value of providerConfigure()
	osClient := m.(openstack.Client)

	networkClient, err := osClient.Network(getRegionNameFromResourceData(d))
	if err != nil {
		return addErrorDiagnostic(diags, err)
	}
	projectID := getProjectID(d, &osClient)
	if projectID == "" {
		return addErrorDiagnostic(diags, fmt.Errorf("cannot obtain project ID"))
	}
	topology, err := networkClient.GetAutoAllocatedTopology(projectID)
	if err != nil {
		return addErrorDiagnostic(diags, err)
	}
	if topology == nil {
		return addErrorDiagnostic(diags, fmt.Errorf("topology is nil"))
	}

	err = d.Set(topologyIDAttribute, topology.NetworkID)
	if err != nil {
		return addErrorDiagnostic(diags, err)
	}
	err = d.Set(projectIDAttribute, topology.ProjectID)
	if err != nil {
		return addErrorDiagnostic(diags, err)
	}
	d.SetId(topology.NetworkID)

	return diags
}

// Look up project ID use the following hierarchy:
// - project_id if user specified it
// - project_name if user specified it
// - current project associated with the credential, which may not exists (e.g. unscoped credential)
func getProjectID(d *schema.ResourceData, osClient *openstack.Client) string {
	projectID := getProjectIDFromResourceData(d)
	if projectID != "" {
		return projectID
	}
	projectName := getProjectNameFromResourceData(d)
	if projectName != "" {
		projectID = osClient.LookupProjectByName(projectName)
		if projectID != "" {
			return projectID
		}
	}
	projectID, _ = osClient.CurrentProject()
	return projectID
}

func getProjectIDFromResourceData(d *schema.ResourceData) string {
	raw := d.Get(projectIDAttribute)
	projectID, ok := raw.(string)
	if !ok {
		return ""
	}
	return projectID
}

func getProjectNameFromResourceData(d *schema.ResourceData) string {
	raw := d.Get(projectNameAttribute)
	projectName, ok := raw.(string)
	if !ok {
		return ""
	}
	return projectName
}

func getRegionNameFromResourceData(d *schema.ResourceData) string {
	regionNameRaw := d.Get(regionNameAttribute)
	regionName, ok := regionNameRaw.(string)
	if !ok {
		return ""
	}
	return regionName
}

func addDiagnostic(diags diag.Diagnostics, diag2 diag.Diagnostic) diag.Diagnostics {
	return append(diags, diag2)
}

func addErrorDiagnostic(diags diag.Diagnostics, err error) diag.Diagnostics {
	return append(diags, diag.FromErr(err)...)
}
