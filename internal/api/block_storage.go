package api

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ResourceBlockStorageGetResponse struct {
	Id                 uuid.UUID         `json:"id"`
	Name               string            `json:"name"`
	Tags               map[string]string `json:"tags"`
	Created            time.Time         `json:"created"`
	Modified           *time.Time        `json:"modified,omitempty"`
	ZoneId             uuid.UUID         `json:"zone_id"`
	OrganizationId     uuid.UUID         `json:"organization_id"`
	AttachedMachineId  *uuid.UUID        `json:"attached_machine_id,omitempty"`
	ImageId            *uuid.UUID        `json:"image_id,omitempty"`
	SnapshotId         *uuid.UUID        `json:"snapshot_id,omitempty"`
	SizeGib            int               `json:"size_gib"`
	DR                 bool              `json:"dr"`
	LastSyncedSnapshot *string           `json:"last_synced_snapshot,omitempty"`
	Assigned           *time.Time        `json:"assigned,omitempty"`
	Prepared           *time.Time        `json:"prepared,omitempty"`
	Deleting           *time.Time        `json:"deleting,omitempty"`
	Deleted            *time.Time        `json:"deleted,omitempty"`
	Status             string            `json:"status"`
}

type ResourceBlockStoragePostResponse struct {
	Id uuid.UUID `json:"id"`
}

type ResourceBlockStoragePatchResponse struct {
	Id uuid.UUID `json:"id"`
}
type ResourceBlockStorageDeleteResponse struct {
	Id     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

func (api *APIClient) PostBlockStorage(
	name string,
	imageId *string,
	snapshotId *string,
	sizeGiB int,
	dr bool,
	tags map[string]string,
) (*ResourceBlockStoragePostResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceBlockStoragePostResponse{}).
		SetBody(map[string]interface{}{
			"zone_id":         api.zoneId,
			"organization_id": api.OrganizationId,
			"name":            name,
			"image_id":        imageId,
			"snapshot_id":     snapshotId,
			"size_gib":        sizeGiB,
			"dr":              dr,
			"tags":            tags,
		}).
		Post(fmt.Sprintf("%s/resource/storage/block_storage", api.pathPrefix))

	return handleAPIResponse[ResourceBlockStoragePostResponse](resp, err)
}

func (api *APIClient) GetBlockStorage(id string) (*ResourceBlockStorageGetResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceBlockStorageGetResponse{}).
		Get(fmt.Sprintf("%s/resource/storage/block_storage/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceBlockStorageGetResponse](resp, err)
}

func (api *APIClient) GetBlockStorages(
	filterAttachedMachineId *string,
) ([]ResourceBlockStorageGetResponse, error) {
	params := map[string]string{}
	setStrIfNotNil(params, "filter_attached_machine_id", filterAttachedMachineId)

	resp, err := api.restyClient.R().
		SetResult(&[]ResourceBlockStorageGetResponse{}).
		SetQueryParams(params).
		Get(fmt.Sprintf("%s/resource/storage/block_storage", api.pathPrefix))

	return handleListAPIResponse[ResourceBlockStorageGetResponse](resp, err)
}

func (api *APIClient) PatchBlockStorage(
	id string, namePtr *string, attachedMachineIdPtr **string, tagsPtr *map[string]string,
) (*ResourceBlockStoragePatchResponse, error) {
	params := map[string]interface{}{}
	setIfNotNil(params, "name", namePtr)
	setIfNotNil(params, "tags", tagsPtr)

	if attachedMachineIdPtr != nil {
		if *attachedMachineIdPtr != nil {
			params["attached_machine_id"] = **attachedMachineIdPtr
		} else {
			params["attached_machine_id"] = nil
		}
	}

	resp, err := api.restyClient.R().
		SetResult(&ResourceBlockStoragePatchResponse{}).
		SetBody(params).
		Patch(fmt.Sprintf("%s/resource/storage/block_storage/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceBlockStoragePatchResponse](resp, err)
}

func (api *APIClient) DeleteBlockStorage(id string) (*ResourceBlockStorageDeleteResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceBlockStorageDeleteResponse{}).
		Delete(fmt.Sprintf("%s/resource/storage/block_storage/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceBlockStorageDeleteResponse](resp, err)
}
