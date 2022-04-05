package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kelseyhightower/envconfig"
	"gitlab.com/cyverse/openstack-auto-allocated-topology/openstack"
)

// New -
func New() *schema.Provider {
	return &schema.Provider{
		Schema:       map[string]*schema.Schema{},
		ResourcesMap: map[string]*schema.Resource{},
		DataSourcesMap: map[string]*schema.Resource{
			"openstack-auto-topology_auto_allocated_topology": dataSourceAutoAllocatedTopology(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	osClient := openstack.NewClient()
	appCred, err := loadApplicationCredentialFromEnv()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	err = osClient.Auth(appCred)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	return osClient, diags
}

func loadApplicationCredentialFromEnv() (openstack.ApplicationCredential, error) {
	var appCred openstack.ApplicationCredential
	err := envconfig.Process("", &appCred)
	if err != nil {
		return openstack.ApplicationCredential{}, fmt.Errorf("fail to load application credential from environment variables, %w", err)
	}
	return appCred, nil
}
