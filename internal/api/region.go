package api

import (
	"fmt"
	"strconv"

	"github.com/google/uuid"
)

type RegionGetResponse struct {
	Id              uuid.UUID  `json:"id"`
	Name            string     `json:"name"`
	RegionId        uuid.UUID  `json:"region_id"`
	SecondaryZoneId *uuid.UUID `json:"secondary_zone_id"`
}

func (api *APIClient) GetRegion(id string) (*RegionGetResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&RegionGetResponse{}).
		Get(fmt.Sprintf("%s/user/region/%s", api.pathPrefix, id))

	return handleAPIResponse[RegionGetResponse](resp, err)
}

func (api *APIClient) GetRegions(
	filterNameIlike *string, skip int, count int,
) ([]RegionGetResponse, error) {
	params := map[string]string{
		"skip":  strconv.FormatInt(int64(skip), 10),
		"count": strconv.FormatInt(int64(count), 10),
	}
	setStrIfNotNil(params, "filter_name_ilike", filterNameIlike)

	resp, err := api.restyClient.R().
		SetResult(&[]RegionGetResponse{}).
		SetQueryParams(params).
		Get(fmt.Sprintf("%s/user/region", api.pathPrefix))

	return handleListAPIResponse[RegionGetResponse](resp, err)
}
