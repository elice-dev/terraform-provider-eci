package api

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ResourceVirtualMachineGetResponse struct {
	Id             uuid.UUID         `json:"id"`
	Tags           map[string]string `json:"tags"`
	Created        time.Time         `json:"created"`
	Modified       *time.Time        `json:"modified,omitempty"`
	ZoneId         uuid.UUID         `json:"zone_id"`
	OrganizationId uuid.UUID         `json:"organization_id"`
	InstanceTypeId uuid.UUID         `json:"instance_type_id"`
	CpuVcore       int               `json:"cpu_vcore"`
	MemoryGib      int               `json:"memory_gib"`
	AlwaysOn       bool              `json:"always_on"`
	DR             bool              `json:"dr"`
	Allocated      *time.Time        `json:"allocated,omitempty"`
	Deleted        *time.Time        `json:"deleted,omitempty"`
	Status         string            `json:"status"`
	Name           string            `json:"name"`
	Username       string            `json:"username"`
	OnInitScript   string            `json:"on_init_script"`
}

type ResourceVirtualMachinePostResponse struct {
	Id uuid.UUID `json:"id"`
}

type ResourceVirtualMachinePatchResponse struct {
	Id uuid.UUID `json:"id"`
}

type ResourceVirtualMachineDeleteResponse struct {
	Id     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

func (api *APIClient) GetVirtualMachine(id string) (*ResourceVirtualMachineGetResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceVirtualMachineGetResponse{}).
		Get(fmt.Sprintf("%s/resource/compute/virtual_machine/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceVirtualMachineGetResponse](resp, err)
}

func (api *APIClient) PostVirtualMachine(
	instanceTypeId string,
	name string,
	alwaysOn bool,
	DR bool,
	username string,
	password string,
	onInitScript string,
	tags map[string]string,
) (*ResourceVirtualMachinePostResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceVirtualMachinePostResponse{}).
		SetBody(map[string]interface{}{
			"zone_id":          api.ZoneId,
			"organization_id":  api.OrganizationId,
			"instance_type_id": instanceTypeId,
			"name":             name,
			"always_on":        alwaysOn,
			"dr":               DR,
			"username":         username,
			"password":         password,
			"on_init_script":   onInitScript,
			"tags":             tags,
		}).
		Post(fmt.Sprintf("%s/resource/compute/virtual_machine", api.pathPrefix))

	return handleAPIResponse[ResourceVirtualMachinePostResponse](resp, err)
}

func (api *APIClient) PatchVirtualMachine(
	id string,
	instanceTypeIdPtr *string,
	namePtr *string,
	alwaysOnPtr *bool,
	tagsPtr *map[string]string,
) (*ResourceVirtualMachinePatchResponse, error) {
	params := map[string]interface{}{}
	setIfNotNil(params, "name", namePtr)
	setIfNotNil(params, "instance_type_id", instanceTypeIdPtr)
	setIfNotNil(params, "always_on", alwaysOnPtr)
	setIfNotNil(params, "tags", tagsPtr)

	resp, err := api.restyClient.R().
		SetResult(&ResourceVirtualMachinePatchResponse{}).
		SetBody(params).
		Patch(fmt.Sprintf("%s/resource/compute/virtual_machine/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceVirtualMachinePatchResponse](resp, err)
}

func (api *APIClient) DeleteVirtualMachine(
	id string,
) (*ResourceVirtualMachineDeleteResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceVirtualMachineDeleteResponse{}).
		Delete(fmt.Sprintf("%s/resource/compute/virtual_machine/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceVirtualMachineDeleteResponse](resp, err)
}
