package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/cyverse/openstack-auto-allocated-topology/openstack"
)

func resourceAutoAllocatedTopology() *schema.Resource {
	return &schema.Resource{
		Description:   "Use this data source to get the auto allocated topology of current project",
		CreateContext: resourceAutoAllocatedTopologyCreate,
		ReadContext:   resourceAutoAllocatedTopologyRead,
		UpdateContext: resourceAutoAllocatedTopologyUpdate,
		DeleteContext: resourceAutoAllocatedTopologyDelete,
		SchemaVersion: 1,
		Schema:        autoAllocatedTopologySchema,
	}
}

func resourceAutoAllocatedTopologyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return dataSourceAutoAllocatedTopologyRead(ctx, d, m)
}

func resourceAutoAllocatedTopologyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return dataSourceAutoAllocatedTopologyRead(ctx, d, m)
}

func resourceAutoAllocatedTopologyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return dataSourceAutoAllocatedTopologyRead(ctx, d, m)
}

func resourceAutoAllocatedTopologyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// from return value of providerConfigure()
	osClient := m.(openstack.Client)

	networkClient, err := osClient.Network(getRegionNameFromResourceData(d))
	if err != nil {
		return addErrorDiagnostic(diags, err)
	}
	projectID, err := getProjectID(d, &osClient)
	if err != nil {
		return addErrorDiagnostic(diags, err)
	}
	if projectID == "" {
		return addErrorDiagnostic(diags, fmt.Errorf("cannot obtain project ID"))
	}
	err = networkClient.DeleteAutoAllocatedTopology(projectID)
	if err != nil {
		return addErrorDiagnostic(diags, err)
	}

	return diags
}
