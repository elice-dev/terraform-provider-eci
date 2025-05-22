package api

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ResourcePublicIpGetResponse struct {
	Id                         uuid.UUID         `json:"id"`
	Tags                       map[string]string `json:"tags"`
	Created                    time.Time         `json:"created"`
	Modified                   *time.Time        `json:"modified,omitempty"`
	ZoneId                     uuid.UUID         `json:"zone_id"`
	OrganizationId             uuid.UUID         `json:"organization_id"`
	AttachedNetworkInterfaceId *uuid.UUID        `json:"attached_network_interface_id,omitempty"`
	DR                         bool              `json:"dr"`
	PoolId                     uuid.UUID         `json:"pool_id"`
	DrPoolId                   *uuid.UUID        `json:"dr_pool_id,omitempty"`
	Deleted                    *time.Time        `json:"deleted,omitempty"`
	Status                     string            `json:"status"`
	Ip                         string            `json:"ip"`
	DrIp                       *string           `json:"dr_ip,omitempty"`
}

type ResourcePublicIpPostResponse struct {
	Id uuid.UUID `json:"id"`
}

type ResourcePublicIpPatchResponse struct {
	Id uuid.UUID `json:"id"`
}

type ResourcePublicIpDeleteResponse struct {
	Id     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

func (api *APIClient) GetPublicIp(id string) (*ResourcePublicIpGetResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourcePublicIpGetResponse{}).
		Get(fmt.Sprintf("%s/resource/network/public_ip/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourcePublicIpGetResponse](resp, err)
}

func (api *APIClient) GetPublicIps(
	filterAttachedNetworkInterfaceIdPtr *string,
) ([]ResourcePublicIpGetResponse, error) {
	params := map[string]string{}
	setStrIfNotNil(
		params, "filter_attached_network_interface_id", filterAttachedNetworkInterfaceIdPtr,
	)

	resp, err := api.restyClient.R().
		SetResult(&[]ResourcePublicIpGetResponse{}).
		SetQueryParams(params).
		Get(fmt.Sprintf("%s/resource/network/public_ip", api.pathPrefix))

	return handleListAPIResponse[ResourcePublicIpGetResponse](resp, err)
}

func (api *APIClient) PostPublicIp(
	dr bool, tags map[string]string,
) (*ResourcePublicIpPostResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourcePublicIpPostResponse{}).
		SetBody(map[string]interface{}{
			"zone_id":         api.zoneId,
			"organization_id": api.OrganizationId,
			"dr":              dr,
			"tags":            tags,
		}).
		Post(fmt.Sprintf("%s/resource/network/public_ip", api.pathPrefix))

	return handleAPIResponse[ResourcePublicIpPostResponse](resp, err)
}

func (api *APIClient) PatchPublicIp(
	id string, attachedNetworkInterfaceIdPtr **string, tagsPtr *map[string]string,
) (*ResourcePublicIpPatchResponse, error) {
	params := map[string]interface{}{}
	setIfNotNil(params, "attached_network_interface_id", attachedNetworkInterfaceIdPtr)
	setIfNotNil(params, "tags", tagsPtr)

	resp, err := api.restyClient.R().
		SetResult(&ResourcePublicIpPatchResponse{}).
		SetBody(params).
		Patch(fmt.Sprintf("%s/resource/network/public_ip/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourcePublicIpPatchResponse](resp, err)
}

func (api *APIClient) DeletePublicIp(id string) (*ResourcePublicIpDeleteResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourcePublicIpDeleteResponse{}).
		Delete(fmt.Sprintf("%s/resource/network/public_ip/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourcePublicIpDeleteResponse](resp, err)
}
