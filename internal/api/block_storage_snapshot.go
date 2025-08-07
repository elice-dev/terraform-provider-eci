package api

import (
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type ResourceBlockStorageSnapshotGetResponse struct {
	Id             uuid.UUID         `json:"id"`
	Name           string            `json:"name"`
	Tags           map[string]string `json:"tags"`
	Created        time.Time         `json:"created"`
	Modified       *time.Time        `json:"modified,omitempty"`
	ZoneId         uuid.UUID         `json:"zone_id"`
	OrganizationId uuid.UUID         `json:"organization_id"`
	BlockStorageId uuid.UUID         `json:"block_storage_id"`
	ImageId        *uuid.UUID        `json:"image_id,omitempty"`
	SizeGib        int               `json:"size_gib"`
	Assigned       *time.Time        `json:"assigned,omitempty"`
	Prepared       *time.Time        `json:"prepared,omitempty"`
	Deleting       *time.Time        `json:"deleting,omitempty"`
	Deleted        *time.Time        `json:"deleted,omitempty"`
	DR             bool              `json:"dr"`
	Status         string            `json:"status"`
}

type ResourceBlockStorageSnapshotPostResponse struct {
	Id uuid.UUID `json:"id"`
}

type ResourceBlockStorageSnapshotPatchResponse struct {
	Id uuid.UUID `json:"id"`
}
type ResourceBlockStorageSnapshotDeleteResponse struct {
	Id     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

func (api *APIClient) PostBlockStorageSnapshot(
	name string, blockStorageId string, tags map[string]string,
) (*ResourceBlockStoragePostResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceBlockStoragePostResponse{}).
		SetBody(map[string]interface{}{
			"zone_id":          api.ZoneId,
			"organization_id":  api.OrganizationId,
			"name":             name,
			"block_storage_id": blockStorageId,
			"tags":             tags,
		}).
		Post(fmt.Sprintf("%s/user/resource/storage/block_storage/snapshot", api.pathPrefix))

	return handleAPIResponse[ResourceBlockStoragePostResponse](resp, err)
}

func (api *APIClient) GetBlockStorageSnapshot(
	id string,
) (*ResourceBlockStorageSnapshotGetResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceBlockStorageSnapshotGetResponse{}).
		Get(fmt.Sprintf("%s/user/resource/storage/block_storage/snapshot/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceBlockStorageSnapshotGetResponse](resp, err)
}

func (api *APIClient) GetBlockStorageSnapshots(
	filterZoneId *string,
	filterOrganizationId *string,
	filterNameIlike *string,
	filterBlockStorageId *string,
	filterImageId *string,
	filterStatus *string,
	filterDr *bool,
	skip int,
	count int,
) ([]ResourceBlockStorageSnapshotGetResponse, error) {
	params := map[string]string{
		"skip":  strconv.FormatInt(int64(skip), 10),
		"count": strconv.FormatInt(int64(count), 10),
	}
	setStrIfNotNil(params, "filter_zone_id", filterZoneId)
	setStrIfNotNil(params, "filter_organization_id", filterOrganizationId)
	setStrIfNotNil(params, "filter_name_ilike", filterNameIlike)
	setStrIfNotNil(params, "filter_block_storage_id", filterBlockStorageId)
	setStrIfNotNil(params, "filter_image_id", filterImageId)
	setStrIfNotNil(params, "filter_status", filterStatus)

	if filterDr != nil {
		params["filter_dr"] = strconv.FormatBool(*filterDr)
	}

	resp, err := api.restyClient.R().
		SetResult(&[]ResourceBlockStorageSnapshotGetResponse{}).
		SetQueryParams(params).
		Get(fmt.Sprintf("%s/user/resource/storage/block_storage/snapshot", api.pathPrefix))

	return handleListAPIResponse[ResourceBlockStorageSnapshotGetResponse](resp, err)
}

func (api *APIClient) PatchBlockStorageSnapshot(
	id string, namePtr *string, tagsPtr *map[string]string,
) (*ResourceBlockStoragePatchResponse, error) {
	params := map[string]interface{}{}
	setIfNotNil(params, "name", namePtr)
	setIfNotNil(params, "tags", tagsPtr)

	resp, err := api.restyClient.R().
		SetResult(&ResourceBlockStoragePatchResponse{}).
		SetBody(params).
		Patch(fmt.Sprintf("%s/user/resource/storage/block_storage/snapshot/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceBlockStoragePatchResponse](resp, err)
}

func (api *APIClient) DeleteBlockStorageSnapshot(
	id string,
) (*ResourceBlockStorageDeleteResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceBlockStorageDeleteResponse{}).
		Delete(fmt.Sprintf("%s/user/resource/storage/block_storage/snapshot/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceBlockStorageDeleteResponse](resp, err)
}
