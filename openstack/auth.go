package openstack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	neturl "net/url"
	"path"
	"time"
)

// ApplicationCredential is env vars for .openrc for application credential
type ApplicationCredential struct {
	RegionName                  string `envconfig:"OS_REGION_NAME"`
	Interface                   string `envconfig:"OS_INTERFACE"`
	AuthURL                     string `envconfig:"OS_AUTH_URL"`
	ApplicationCredentialSecret string `envconfig:"OS_APPLICATION_CREDENTIAL_SECRET"`
	ApplicationCredentialID     string `envconfig:"OS_APPLICATION_CREDENTIAL_ID"`
	AuthType                    string `envconfig:"OS_AUTH_TYPE"`
	IdentityAPIVersion          string `envconfig:"OS_IDENTITY_API_VERSION"`
}

// TokenMetadata is the data returned in HTTP response body when obtaining token
type TokenMetadata struct {
	IsDomain bool     `json:"is_domain"`
	Methods  []string `json:"methods"`
	Roles    []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"roles"`
	ExpiresAt time.Time `json:"expires_at"`
	Project   struct {
		Domain struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"domain"`
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"project"`
	Catalog []struct {
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
	} `json:"catalog"`
	ApplicationCredentialRestricted bool `json:"application_credential_restricted"`
	User                            struct {
		PasswordExpiresAt interface{} `json:"password_expires_at"`
		Domain            struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"domain"`
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"user"`
	AuditIDs []string  `json:"audit_ids"`
	IssuedAt time.Time `json:"issued_at"`
}

// obtain a token using an application credential
// https://docs.openstack.org/api-ref/identity/v3/?expanded=authenticating-with-an-application-credential-detail#authenticating-with-an-application-credential
func obtainTokenWithAppCred(baseURL, appCredID, appCredSecret string) (string, TokenMetadata, error) {
	payload := map[string]interface{}{
		"auth": map[string]interface{}{
			"identity": map[string]interface{}{
				"methods": []string{
					"application_credential",
				},
				"application_credential": map[string]string{
					"id":     appCredID,
					"secret": appCredSecret,
				},
			},
		},
	}
	return makeAuthRequest(baseURL, payload)
}

func makeAuthRequest(baseURL string, payload map[string]interface{}) (token string, tokenMetadata TokenMetadata, err error) {
	var reqBody bytes.Buffer
	err = json.NewEncoder(&reqBody).Encode(payload)
	if err != nil {
		return "", TokenMetadata{}, err
	}
	url, err := neturl.Parse(baseURL)
	if err != nil {
		return "", TokenMetadata{}, err
	}
	url.Path = path.Join(url.Path, "/auth/tokens")
	req, err := http.NewRequest(http.MethodPost, url.String(), &reqBody)
	if err != nil {
		return "", TokenMetadata{}, err
	}
	resp, err := makeHTTPRequestWithRetry(req)
	if err != nil {
		return "", TokenMetadata{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("%s, %s", resp.Status, url)
		return "", TokenMetadata{}, err
	}
	token, err = extractTokenFromAuthResponse(resp)
	if err != nil {
		return "", TokenMetadata{}, err
	}
	tokenMetadata, err = extractTokenMetadataFromAuthResponse(resp)
	if err != nil {
		return "", TokenMetadata{}, err
	}
	return token, tokenMetadata, nil
}

func extractTokenFromAuthResponse(resp *http.Response) (string, error) {
	token := resp.Header.Get("X-Subject-Token")
	if token == "" {
		return "", fmt.Errorf("token not in header")
	}
	return token, nil
}

func extractTokenMetadataFromAuthResponse(resp *http.Response) (TokenMetadata, error) {
	var respBody struct {
		TokenMetadata TokenMetadata `json:"token"`
	}
	err := json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return TokenMetadata{}, err
	}
	resp.Body.Close()
	return respBody.TokenMetadata, nil
}
