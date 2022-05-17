package openstack

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"path"
	"time"
)

// Client is base client for OpenStack API
type Client struct {
	appCred        ApplicationCredential
	token          string
	tokenMetadata  TokenMetadata
	catalogEntries []CatalogEntry
}

// NewClient creates a new Client
func NewClient() Client {
	return Client{}
}

// Auth authenticate with OpenStack API using an application credential
func (c *Client) Auth(appCred ApplicationCredential) error {
	err := c.checkCred(appCred)
	if err != nil {
		return err
	}
	token, metadata, err := obtainTokenWithAppCred(appCred.AuthURL, appCred.ApplicationCredentialID, appCred.ApplicationCredentialSecret)
	if err != nil {
		return err
	}
	catalogEntries, err := c.getCatalog(appCred.AuthURL, token)
	if err != nil {
		return err
	}
	c.tokenMetadata = metadata
	c.catalogEntries = catalogEntries
	c.token = token
	c.appCred = appCred
	return nil
}

func (c *Client) checkCred(appCred ApplicationCredential) error {
	if appCred.AuthURL == "" {
		return fmt.Errorf("OS_AUTH_URL missing")
	}
	if appCred.ApplicationCredentialID == "" {
		return fmt.Errorf("OS_APPLICATION_CREDENTIAL_ID missing")
	}
	if appCred.ApplicationCredentialSecret == "" {
		return fmt.Errorf("OS_APPLICATION_CREDENTIAL_SECRET missing")
	}
	return nil
}

// get a list of catalogs
// https://docs.openstack.org/api-ref/identity/v3/?expanded=get-service-catalog-detail#get-service-catalog
func (c *Client) getCatalog(baseURL string, token string) ([]CatalogEntry, error) {
	url, err := neturl.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	url.Path = path.Join(url.Path, "/auth/catalog")
	resp, err := makeRequest(http.MethodGet, url.String(), token, nil)
	if err != nil {
		return nil, err
	}
	var respBody struct {
		Catalog []CatalogEntry `json:"catalog"`
		Links   struct {
			Self     string      `json:"self"`
			Previous interface{} `json:"previous"`
			Next     interface{} `json:"next"`
		} `json:"links"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	if len(respBody.Catalog) == 0 {
		return nil, fmt.Errorf("no catalog entries")
	}
	return respBody.Catalog, nil
}

// Network returns a NetworkClient
func (c Client) Network() (*NetworkClient, error) {
	if c.token == "" {
		return nil, fmt.Errorf("token not set")
	}
	networkEndpoint, err := findEndpoint(c.catalogEntries, "network", c.appCred.RegionName, c.appCred.Interface)
	if err != nil {
		return nil, err
	}
	return &NetworkClient{
		baseURL:       networkEndpoint.URL,
		token:         c.token,
		tokenMetadata: c.tokenMetadata,
	}, nil
}

// CurrentProject returns the current project (project of application credential used for authentication)
func (c Client) CurrentProject() (id string, name string) {
	return c.tokenMetadata.Project.ID, c.tokenMetadata.Project.Name
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

func makeRequest(httpMethod string, url string, token string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(httpMethod, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Auth-Token", token)
	resp, err := makeHTTPRequestWithRetry(req)
	if err != nil {
		return resp, err
	}
	if resp.StatusCode != 200 {
		return resp, fmt.Errorf("%s, %s", resp.Status, url)
	}
	return resp, nil
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
