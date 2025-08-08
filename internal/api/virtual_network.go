package api

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type NetworkFirewallRule struct {
	Proto       string `json:"proto"              tfsdk:"proto"`
	Source      string `json:"source"             tfsdk:"source"`
	Destination string `json:"destination"        tfsdk:"destination"`
	Port        *int   `json:"port,omitempty"     tfsdk:"port"`
	PortEnd     *int   `json:"port_end,omitempty" tfsdk:"port_end"`
	Action      string `json:"action"             tfsdk:"action"`
	Comment     string `json:"comment"            tfsdk:"comment"`
}

type ResourceVirtualNetworkGetResponse struct {
	Id             uuid.UUID             `json:"id"`
	Tags           map[string]string     `json:"tags"`
	Created        time.Time             `json:"created"`
	Modified       *time.Time            `json:"modified,omitempty"`
	ZoneId         uuid.UUID             `json:"zone_id"`
	OrganizationId uuid.UUID             `json:"organization_id"`
	Deleted        *time.Time            `json:"deleted,omitempty"`
	Status         string                `json:"status"`
	Name           string                `json:"name"`
	NetworkCidr    string                `json:"network_cidr"`
	FirewallRules  []NetworkFirewallRule `json:"firewall_rules"`
}

type ResourceVirtualNetworkPostResponse struct {
	Id uuid.UUID `json:"id"`
}

type ResourceVirtualNetworkPatchResponse struct {
	Id uuid.UUID `json:"id"`
}

type ResourceVirtualNetworkDeleteResponse struct {
	Id     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

func (api *APIClient) GetVirtualNetwork(id string) (*ResourceVirtualNetworkGetResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceVirtualNetworkGetResponse{}).
		Get(fmt.Sprintf("%s/user/resource/network/virtual_network/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceVirtualNetworkGetResponse](resp, err)
}

func (api *APIClient) PostVirtualNetwork(
	name string, networkCidr string, tags map[string]string,
) (*ResourceVirtualNetworkPostResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceVirtualNetworkPostResponse{}).
		SetBody(map[string]interface{}{
			"zone_id":         api.ZoneId,
			"organization_id": api.OrganizationId,
			"name":            name,
			"network_cidr":    networkCidr,
			"tags":            tags,
		}).
		Post(fmt.Sprintf("%s/user/resource/network/virtual_network", api.pathPrefix))

	return handleAPIResponse[ResourceVirtualNetworkPostResponse](resp, err)
}

func (api *APIClient) PatchVirtualNetwork(
	id string, namePtr *string, firewallRulesPtr *[]NetworkFirewallRule, tags *map[string]string,
) (*ResourceVirtualNetworkPatchResponse, error) {
	params := map[string]interface{}{}
	setIfNotNil(params, "name", namePtr)
	setIfNotNil(params, "firewall_rules", firewallRulesPtr)
	setIfNotNil(params, "tags", tags)

	resp, err := api.restyClient.R().
		SetResult(&ResourceVirtualNetworkPatchResponse{}).
		SetBody(params).
		Patch(fmt.Sprintf("%s/user/resource/network/virtual_network/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceVirtualNetworkPatchResponse](resp, err)
}

func (api *APIClient) DeleteVirtualNetwork(
	id string,
) (*ResourceVirtualNetworkDeleteResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceVirtualNetworkDeleteResponse{}).
		Delete(fmt.Sprintf("%s/user/resource/network/virtual_network/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceVirtualNetworkDeleteResponse](resp, err)
}
