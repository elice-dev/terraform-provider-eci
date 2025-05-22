package api

import (
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type ResourceBlockStorageImageGetResponse struct {
	Id          uuid.UUID `json:"id"`
	Created     time.Time `json:"created"`
	ZoneId      uuid.UUID `json:"zone_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Keywords    []string  `json:"keywords"`
	SizeGib     int       `json:"size_gib"`
	Status      string    `json:"status"`
}

func (api *APIClient) GetBlockStorageImage(
	id string,
) (*ResourceBlockStorageImageGetResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceBlockStorageImageGetResponse{}).
		Get(fmt.Sprintf("%s/user/infra/block_storage_image/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceBlockStorageImageGetResponse](resp, err)
}

func (api *APIClient) GetBlockStorageImages(
	filterNameIlike *string, skip int, count int,
) ([]ResourceBlockStorageImageGetResponse, error) {
	params := map[string]string{
		"filter_zone_id": api.zoneId,
		"skip":           strconv.FormatInt(int64(skip), 10),
		"count":          strconv.FormatInt(int64(count), 10),
	}
	setStrIfNotNil(params, "filter_name_ilike", filterNameIlike)

	resp, err := api.restyClient.R().
		SetResult(&[]ResourceBlockStorageImageGetResponse{}).
		SetQueryParams(params).
		Get(fmt.Sprintf("%s/user/infra/block_storage_image", api.pathPrefix))

	return handleListAPIResponse[ResourceBlockStorageImageGetResponse](resp, err)
}
