package api

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ResourceSubnetGetResponse struct {
	Id                uuid.UUID         `json:"id"`
	Tags              map[string]string `json:"tags"`
	Created           time.Time         `json:"created"`
	Modified          *time.Time        `json:"modified,omitempty"`
	ZoneId            uuid.UUID         `json:"zone_id"`
	OrganizationId    uuid.UUID         `json:"organization_id"`
	AttachedNetworkId uuid.UUID         `json:"attached_network_id"`
	Activated         *time.Time        `json:"activated,omitempty"`
	Deleted           *time.Time        `json:"deleted,omitempty"`
	Status            string            `json:"status"`
	Name              string            `json:"name"`
	Purpose           string            `json:"purpose"`
	NetworkGw         string            `json:"network_gw"`
}

type ResourceSubnetPostResponse struct {
	Id uuid.UUID `json:"id"`
}

type ResourceSubnetPatchResponse struct {
	Id uuid.UUID `json:"id"`
}

type ResourceSubnetDeleteResponse struct {
	Id     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

func (api *APIClient) PostSubnet(
	name string, attachedNetworkId string, purpose string, networkGw string, tags map[string]string,
) (*ResourceSubnetPostResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceSubnetPostResponse{}).
		SetBody(map[string]interface{}{
			"zone_id":             api.ZoneId,
			"organization_id":     api.OrganizationId,
			"name":                name,
			"attached_network_id": attachedNetworkId,
			"purpose":             purpose,
			"network_gw":          networkGw,
			"tags":                tags,
		}).
		Post(fmt.Sprintf("%s/resource/network/subnet", api.pathPrefix))

	return handleAPIResponse[ResourceSubnetPostResponse](resp, err)
}

func (api *APIClient) GetSubnet(id string) (*ResourceSubnetGetResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceSubnetGetResponse{}).
		Get(fmt.Sprintf("%s/resource/network/subnet/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceSubnetGetResponse](resp, err)
}

func (api *APIClient) PatchSubnet(
	id string, namePtr *string, tagsPtr *map[string]string,
) (*ResourceSubnetPatchResponse, error) {
	params := map[string]interface{}{}
	setIfNotNil(params, "name", namePtr)
	setIfNotNil(params, "tags", tagsPtr)

	resp, err := api.restyClient.R().
		SetResult(&ResourceSubnetPatchResponse{}).
		SetBody(params).
		Patch(fmt.Sprintf("%s/resource/network/subnet/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceSubnetPatchResponse](resp, err)
}

func (api *APIClient) DeleteSubnet(id string) (*ResourceSubnetDeleteResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceSubnetDeleteResponse{}).
		Delete(fmt.Sprintf("%s/resource/network/subnet/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceSubnetDeleteResponse](resp, err)
}
