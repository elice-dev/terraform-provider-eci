package api

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ResourceVirtualMachineAllocationGetResponse struct {
	Id                 uuid.UUID         `json:"id"`
	Tags               map[string]string `json:"tags"`
	Created            time.Time         `json:"created"`
	Modified           *time.Time        `json:"modified,omitempty"`
	ZoneId             uuid.UUID         `json:"zone_id"`
	OrganizationId     uuid.UUID         `json:"organization_id"`
	MachineId          uuid.UUID         `json:"machine_id"`
	RequestedCpuVcore  int               `json:"requested_cpu_vcore"`
	RequestedMemoryGib int               `json:"requested_memory_gib"`
	RequestedDevices   []string          `json:"requested_devices"`
	LastHeartbeat      *time.Time        `json:"last_heartbeat,omitempty"`
	Assigned           *time.Time        `json:"assigned,omitempty"`
	Taken              *time.Time        `json:"taken,omitempty"`
	Started            *time.Time        `json:"started,omitempty"`
	Terminating        *time.Time        `json:"terminating,omitempty"`
	Terminated         *time.Time        `json:"terminated,omitempty"`
	Status             string            `json:"status"`
}

type ResourceVirtualMachineAllocationPostResponse struct {
	Id uuid.UUID `json:"id"`
}

type ResourceVirtualMachineAllocationPatchResponse struct {
	Id uuid.UUID `json:"id"`
}

type ResourceVirtualMachineAllocationDeleteResponse struct {
	Id     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

func (api *APIClient) GetVirtualMachineAllocation(
	id string,
) (*ResourceVirtualMachineAllocationGetResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceVirtualMachineAllocationGetResponse{}).
		Get(fmt.Sprintf("%s/resource/compute/virtual_machine_allocation/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceVirtualMachineAllocationGetResponse](resp, err)
}

func (api *APIClient) GetVirtualMachineAllocations(
	filterMachineIdPtr *string, filterStatusPtr *string,
) ([]ResourceVirtualMachineAllocationGetResponse, error) {
	params := map[string]string{}
	setStrIfNotNil(params, "filter_machine_id", filterMachineIdPtr)
	setStrIfNotNil(params, "filter_status", filterStatusPtr)

	resp, err := api.restyClient.R().
		SetResult(&[]ResourceVirtualMachineAllocationGetResponse{}).
		SetQueryParams(params).
		Get(fmt.Sprintf("%s/resource/compute/virtual_machine_allocation", api.pathPrefix))

	return handleListAPIResponse[ResourceVirtualMachineAllocationGetResponse](resp, err)
}

func (api *APIClient) PostVirtualMachineAllocation(
	machineId string, tags map[string]string,
) (*ResourceVirtualMachineAllocationPostResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceVirtualMachineAllocationPostResponse{}).
		SetBody(map[string]interface{}{
			"zone_id":         api.zoneId,
			"organization_id": api.OrganizationId,
			"machine_id":      machineId,
			"tags":            tags,
		}).
		Post(fmt.Sprintf("%s/resource/compute/virtual_machine_allocation", api.pathPrefix))

	return handleAPIResponse[ResourceVirtualMachineAllocationPostResponse](resp, err)
}

func (api *APIClient) DeleteVirtualMachineAllocation(id string,
) (*ResourceVirtualMachineAllocationDeleteResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceVirtualMachineAllocationDeleteResponse{}).
		Delete(fmt.Sprintf("%s/resource/compute/virtual_machine_allocation/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceVirtualMachineAllocationDeleteResponse](resp, err)
}
