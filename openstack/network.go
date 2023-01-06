package openstack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// NetworkClient is client for OpenStack Network (Neutron) API
type NetworkClient struct {
	baseURL       string // base URL for network API
	token         string
	tokenMetadata TokenMetadata
}

// GetAutoAllocatedTopology get (or create if not exists) the auto allocated topology of a project.
// https://docs.openstack.org/api-ref/network/v2/?expanded=show-auto-allocated-topology-details-detail#show-auto-allocated-topology-details
func (c NetworkClient) GetAutoAllocatedTopology(projectID string) (*AutoAllocatedTopology, error) {

	url := fmt.Sprintf("%s/v2.0/auto-allocated-topology/%s", c.baseURL, projectID)
	resp, err := makeRequest(http.MethodGet, url, c.token, nil, []int{200})
	if err != nil {
		return nil, err
	}
	var respBody struct {
		Topology struct {
			ID        string `json:"id"`
			TenantID  string `json:"tenant_id"`
			ProjectID string `json:"project_id"`
		} `json:"auto_allocated_topology"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	return &AutoAllocatedTopology{
		NetworkID: respBody.Topology.ID,
		ProjectID: projectID,
	}, nil
}

// DeleteAutoAllocatedTopology deletes the auto allocated topology for a project.
// https://docs.openstack.org/api-ref/network/v2/?expanded=delete-the-auto-allocated-topology-detail#show-auto-allocated-topology-details
func (c NetworkClient) DeleteAutoAllocatedTopology(projectID string) error {

	url := fmt.Sprintf("%s/v2.0/auto-allocated-topology/%s", c.baseURL, projectID)
	resp, err := makeRequest(http.MethodDelete, url, c.token, nil, []int{200, 204})
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// AutoAllocatedTopology is network (and related entities) that created by openstack via the auto allocated topology extension
type AutoAllocatedTopology struct {
	NetworkID string `json:"network_id"`
	ProjectID string `json:"project_id"`
}
