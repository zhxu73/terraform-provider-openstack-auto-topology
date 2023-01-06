package openstack

import (
	"bytes"
	"fmt"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/catalog"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/users"
	"github.com/mitchellh/mapstructure"
	"io"
	"net/http"
	"time"
)

// Client is base client for OpenStack API
type Client struct {
	credEnv        CredentialEnv
	token          string
	tokenMetadata  TokenMetadata
	catalogEntries []CatalogEntry
	provider       *gophercloud.ProviderClient
}

// NewClient creates a new Client
func NewClient() Client {
	return Client{}
}

// Auth authenticate with OpenStack API using an application credential
func (c *Client) Auth(credEnv CredentialEnv) error {
	opts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return err
	}

	provider, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		return err
	}
	token, metadata, err := obtainToken(provider)
	if err != nil {
		return err
	}
	c.provider = provider
	c.tokenMetadata = metadata
	c.token = token
	c.credEnv = credEnv
	return nil
}

// Network returns a NetworkClient for a region.
// If regionName parameter is empty (""), then you will try to use OS_REGION_NAME from application credential.
func (c *Client) Network(regionName string) (*NetworkClient, error) {
	if c.token == "" {
		return nil, fmt.Errorf("token not set")
	}
	if regionName == "" {
		regionName = c.credEnv.RegionName
	}
	identityClient, err := openstack.NewIdentityV3(c.provider, gophercloud.EndpointOpts{})
	if err != nil {
		return nil, err
	}
	entry, err := findNeutronCatalogEntry(identityClient, regionName, c.credEnv.Interface)
	if err != nil {
		return nil, err
	}
	return &NetworkClient{
		baseURL:       entry.URL,
		token:         c.token,
		tokenMetadata: c.tokenMetadata,
	}, nil
}

// CurrentProject returns the current project (project of application credential used for authentication)
func (c *Client) CurrentProject() (id string, name string) {
	return c.tokenMetadata.Project.ID, c.tokenMetadata.Project.Name
}

// LookupProjectByName looks up the ID of a project by its name
// https://docs.openstack.org/api-ref/identity/v3/index.html?expanded=list-projects-for-user-detail#list-projects-for-user
func (c *Client) LookupProjectByName(projectName string) (id string, err error) {
	identityClient, err := openstack.NewIdentityV3(c.provider, gophercloud.EndpointOpts{})
	if err != nil {
		return "", err
	}
	list, err := users.ListProjects(identityClient, c.tokenMetadata.User.ID).AllPages()
	if err != nil {
		return "", err
	}
	if empty, err := list.IsEmpty(); err != nil || empty {
		return "", fmt.Errorf("%w", err)
	}

	var respBody struct {
		Projects []struct {
			ID          string                 `json:"id"`
			Name        string                 `json:"name"`
			DomainID    string                 `json:"domain_id"`
			Description string                 `json:"description"`
			Enabled     bool                   `json:"enabled"`
			ParentID    string                 `json:"parent_id"`
			IsDomain    bool                   `json:"is_domain"`
			Tags        []string               `json:"tags"`
			Options     map[string]interface{} `json:"options"`
			Links       struct {
				Self string `json:"self"`
			} `json:"links"`
		} `json:"projects"`
		Links struct {
			Next     interface{} `json:"next"`
			Self     string      `json:"self"`
			Previous interface{} `json:"previous"`
		} `json:"links"`
	}
	err = mapstructure.Decode(list.GetBody(), &respBody)
	if err != nil {
		return "", err
	}
	for _, project := range respBody.Projects {
		if project.Name == projectName {
			// return the first found
			return project.ID, nil
		}
	}
	return "", fmt.Errorf("project %s not found", projectName)
}

func findNeutronCatalogEntry(identityClient *gophercloud.ServiceClient, regionName, interfaceName string) (CatalogEndpoint, error) {
	catalogList := catalog.List(identityClient)
	page, err := catalogList.AllPages()
	if err != nil {
		return CatalogEndpoint{}, err
	}
	empty, err := page.IsEmpty()
	if err != nil {
		return CatalogEndpoint{}, err
	}
	if empty {
		return CatalogEndpoint{}, fmt.Errorf("catalog is empty")
	}
	var respBody struct {
		Catalog []CatalogEntry `json:"catalog"`
		Links   struct {
			Self     string      `json:"self"`
			Previous interface{} `json:"previous"`
			Next     interface{} `json:"next"`
		} `json:"links"`
	}
	err = mapstructure.Decode(page.GetBody(), &respBody)
	if err != nil {
		return CatalogEndpoint{}, err
	}
	if len(respBody.Catalog) == 0 {
		return CatalogEndpoint{}, fmt.Errorf("no catalog entries")
	}
	endpoint, err := findEndpoint(respBody.Catalog, "network", regionName, interfaceName)
	if err != nil {
		return CatalogEndpoint{}, err
	}
	return endpoint, nil
}

func findEndpoint(catalogEntries []CatalogEntry, serviceType, regionName, interfaceName string) (CatalogEndpoint, error) {
	var serviceEntry CatalogEntry
	for _, entry := range catalogEntries {
		if entry.Type == "network" {
			serviceEntry = entry
			break
		}
	}
	if serviceEntry.Type == "" {
		return CatalogEndpoint{}, fmt.Errorf("service type %s not found in catalog", serviceType)
	}
	var serviceEndpoint CatalogEndpoint
	for _, endpoint := range serviceEntry.Endpoints {
		if endpoint.Region == regionName && endpoint.Interface == interfaceName {
			serviceEndpoint = endpoint
			break
		}
	}
	if serviceEndpoint.Region == "" {
		return CatalogEndpoint{}, fmt.Errorf("service catalog for %s does not have endpoint for region %s", serviceEntry.Name, regionName)
	}
	return serviceEndpoint, nil
}

func makeRequest(httpMethod string, url string, token string, body io.Reader, successStatusCodes []int) (*http.Response, error) {
	req, err := http.NewRequest(httpMethod, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Auth-Token", token)
	resp, err := makeHTTPRequestWithRetry(req)
	if err != nil {
		return resp, err
	}
	for _, code := range successStatusCodes {
		if code == resp.StatusCode {
			return resp, nil
		}
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp, fmt.Errorf("%s, %s, %s", resp.Status, url, buf.String())
}

func makeHTTPRequestWithRetry(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	const maxRetryCount = 3

	client := getHTTPClient()
	for i := 0; i < maxRetryCount; i++ {
		resp, err = client.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode >= 400 {
			time.Sleep(time.Millisecond * 500 * time.Duration(i)) // exp backoff
			continue
		}
		break
	}
	return resp, err
}

func getHTTPClient() *http.Client {
	return &http.Client{Timeout: 5 * time.Second}
}

// CatalogEntry is an entry of Catalog, it contains metadata for an OpenStack service
type CatalogEntry struct {
	Endpoints []CatalogEndpoint `json:"endpoints"`
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Name      string            `json:"name"`
}

// CatalogEndpoint is endpoint for a CatalogEntry
type CatalogEndpoint struct {
	ID        string `json:"id"`
	Interface string `json:"interface"`
	Region    string `json:"region"`
	URL       string `json:"url"`
}
