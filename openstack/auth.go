package openstack

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	tokensv3 "github.com/gophercloud/gophercloud/openstack/identity/v3/tokens"
	"time"
)

// CredentialEnv is env vars for .openrc for openstack credential
type CredentialEnv struct {
	RegionName                  string `envconfig:"OS_REGION_NAME"`
	Interface                   string `envconfig:"OS_INTERFACE" default:"public"`
	AuthURL                     string `envconfig:"OS_AUTH_URL"`
	ApplicationCredentialSecret string `envconfig:"OS_APPLICATION_CREDENTIAL_SECRET"`
	ApplicationCredentialID     string `envconfig:"OS_APPLICATION_CREDENTIAL_ID"`
	AuthType                    string `envconfig:"OS_AUTH_TYPE"`
	IdentityAPIVersion          string `envconfig:"OS_IDENTITY_API_VERSION"`
}

// TokenMetadata is the data returned in HTTP response body when obtaining token
type TokenMetadata struct {
	IsDomain                        bool                   `json:"is_domain"`
	Methods                         []string               `json:"methods"`
	Roles                           []TokenMetadataRole    `json:"roles"`
	ExpiresAt                       time.Time              `json:"expires_at"`
	Project                         TokenMetadataProject   `json:"project"`
	Catalog                         []TokenMetadataCatalog `json:"catalog"`
	ApplicationCredentialRestricted bool                   `json:"application_credential_restricted"`
	User                            TokenMetadataUser      `json:"user"`
	AuditIDs                        []string               `json:"audit_ids"`
	IssuedAt                        time.Time              `json:"issued_at"`
}

type TokenMetadataRole struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TokenMetadataProject struct {
	Domain Domain `json:"domain"`
	ID     string `json:"id"`
	Name   string `json:"name"`
}

type TokenMetadataCatalog struct {
	Endpoints []struct {
		RegionID  string `json:"region_id"`
		URL       string `json:"url"`
		Region    string `json:"region"`
		Interface string `json:"interface"`
		ID        string `json:"id"`
	} `json:"endpoints"`
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TokenMetadataUser struct {
	PasswordExpiresAt interface{} `json:"password_expires_at"`
	Domain            Domain      `json:"domain"`
	ID                string      `json:"id"`
	Name              string      `json:"name"`
}

type Domain struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func obtainToken(provider *gophercloud.ProviderClient) (string, TokenMetadata, error) {
	identityClient, err := openstack.NewIdentityV3(provider, gophercloud.EndpointOpts{})
	if err != nil {
		return "", TokenMetadata{}, err
	}
	metadata, err := authResultToTokenMetadata(provider.GetAuthResult(), identityClient)
	if err != nil {
		return "", TokenMetadata{}, err
	}
	return provider.Token(), metadata, nil
}

func authResultToTokenMetadata(authResult gophercloud.AuthResult, client *gophercloud.ServiceClient) (TokenMetadata, error) {
	switch result := authResult.(type) {
	case tokensv3.CreateResult:
		return extractTokenMetadataFromAuthResult(result)
	case tokensv3.GetResult:
		return extractTokenMetadataFromAuthResult(result)
	default:
		res := tokensv3.Get(client, client.ProviderClient.TokenID)
		if res.Err != nil {
			return TokenMetadata{}, res.Err
		}
		return extractTokenMetadataFromAuthResult(res)
	}
}

// implement by tokensv3.CreateResult and tokensv3.GetResult
type iAuthResult interface {
	ExtractTokenID() (string, error)
	ExtractServiceCatalog() (*tokensv3.ServiceCatalog, error)
	ExtractUser() (*tokensv3.User, error)
	ExtractRoles() ([]tokensv3.Role, error)
	ExtractProject() (*tokensv3.Project, error)
	ExtractDomain() (*tokensv3.Domain, error)
}

func extractTokenMetadataFromAuthResult(result iAuthResult) (TokenMetadata, error) {
	user, err := result.ExtractUser()
	if err != nil {
		return TokenMetadata{}, err
	}
	project, err := result.ExtractProject()
	if err != nil {
		return TokenMetadata{}, err
	}
	return TokenMetadata{
		Roles: nil,
		Project: TokenMetadataProject{
			Domain: Domain{
				ID:   project.Domain.ID,
				Name: project.Domain.Name,
			},
			ID:   project.ID,
			Name: project.Name,
		},
		ApplicationCredentialRestricted: false,
		User: TokenMetadataUser{
			PasswordExpiresAt: nil,
			Domain: Domain{
				ID:   user.Domain.ID,
				Name: user.Domain.Name,
			},
			ID:   user.ID,
			Name: user.Name,
		},
	}, nil
}
