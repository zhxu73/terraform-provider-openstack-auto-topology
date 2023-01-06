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
		Schema: map[string]*schema.Schema{},
		ResourcesMap: map[string]*schema.Resource{
			"openstack-auto-topology_auto_allocated_topology": resourceAutoAllocatedTopology(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"openstack-auto-topology_auto_allocated_topology": dataSourceAutoAllocatedTopology(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	osClient := openstack.NewClient()
	appCred, err := loadCredentialFromEnv()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	err = osClient.Auth(appCred)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	return osClient, diags
}

func loadCredentialFromEnv() (openstack.CredentialEnv, error) {
	var cred openstack.CredentialEnv
	err := envconfig.Process("", &cred)
	if err != nil {
		return openstack.CredentialEnv{}, fmt.Errorf("fail to load application credential from environment variables, %w", err)
	}
	return cred, nil
}
