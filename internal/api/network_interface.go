package api

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ResourceNetworkInterfaceGetResponse struct {
	Id                uuid.UUID         `json:"id"`
	Tags              map[string]string `json:"tags"`
	Created           time.Time         `json:"created"`
	Modified          *time.Time        `json:"modified,omitempty"`
	ZoneId            uuid.UUID         `json:"zone_id"`
	OrganizationId    uuid.UUID         `json:"organization_id"`
	AttachedSubnetId  uuid.UUID         `json:"attached_subnet_id"`
	AttachedMachineId *uuid.UUID        `json:"attached_machine_id,omitempty"`
	DR                bool              `json:"bool"`
	Deleted           *time.Time        `json:"deleted,omitempty"`
	Status            string            `json:"status"`
	Name              string            `json:"name"`
	Ip                string            `json:"ip"`
	Mac               string            `json:"mac"`
}

type ResourceNetworkInterfacePostResponse struct {
	Id uuid.UUID `json:"id"`
}

type ResourceNetworkInterfacePatchResponse struct {
	Id uuid.UUID `json:"id"`
}

type ResourceNetworkInterfaceDeleteResponse struct {
	Id     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

func (api *APIClient) GetNetworkInterface(
	id string,
) (*ResourceNetworkInterfaceGetResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceNetworkInterfaceGetResponse{}).
		Get(fmt.Sprintf("%s/resource/network/network_interface/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceNetworkInterfaceGetResponse](resp, err)
}

func (api *APIClient) GetNetworkInterfaces(
	filterAttachedMachineIdPtr *string,
) ([]ResourceNetworkInterfaceGetResponse, error) {
	params := map[string]string{}
	setStrIfNotNil(params, "filter_attached_machine_id", filterAttachedMachineIdPtr)

	resp, err := api.restyClient.R().
		SetResult(&[]ResourceNetworkInterfaceGetResponse{}).
		SetQueryParams(params).
		Get(fmt.Sprintf("%s/resource/network/network_interface", api.pathPrefix))

	return handleListAPIResponse[ResourceNetworkInterfaceGetResponse](resp, err)
}

func (api *APIClient) PostNetworkInterface(
	name string,
	attachedSubnetId string,
	dr bool,
	ipPtr *string,
	macPtr *string,
	tags map[string]string,
) (*ResourceNetworkInterfacePostResponse, error) {
	params := map[string]interface{}{
		"zone_id":            api.zoneId,
		"organization_id":    api.OrganizationId,
		"name":               name,
		"attached_subnet_id": attachedSubnetId,
		"dr":                 dr,
		"tags":               tags,
	}
	setIfNotNil(params, "ip", ipPtr)
	setIfNotNil(params, "mac", macPtr)

	resp, err := api.restyClient.R().
		SetResult(&ResourceNetworkInterfacePostResponse{}).
		SetBody(params).
		Post(fmt.Sprintf("%s/resource/network/network_interface", api.pathPrefix))

	return handleAPIResponse[ResourceNetworkInterfacePostResponse](resp, err)
}

func (api *APIClient) PatchNetworkInterface(
	id string, namePtr *string, attachedMachineIdPtr **string, tagsPtr *map[string]string,
) (*ResourceNetworkInterfacePatchResponse, error) {
	params := map[string]interface{}{}
	setIfNotNil(params, "name", namePtr)
	setIfNotNil(params, "attached_machine_id", attachedMachineIdPtr)
	setIfNotNil(params, "tags", tagsPtr)

	resp, err := api.restyClient.R().
		SetResult(&ResourceNetworkInterfacePatchResponse{}).
		SetBody(params).
		Patch(fmt.Sprintf("%s/resource/network/network_interface/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceNetworkInterfacePatchResponse](resp, err)
}

func (api *APIClient) DeleteNetworkInterface(
	id string,
) (*ResourceNetworkInterfaceDeleteResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceNetworkInterfaceDeleteResponse{}).
		Delete(fmt.Sprintf("%s/resource/network/network_interface/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceNetworkInterfaceDeleteResponse](resp, err)
}
