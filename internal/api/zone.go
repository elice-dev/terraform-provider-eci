package api

import (
	"fmt"
	"strconv"

	"github.com/google/uuid"
)

type InfraZoneGetResponse struct {
	Id              uuid.UUID  `json:"id"`
	Name            string     `json:"name"`
	RegionId        uuid.UUID  `json:"region_id"`
	SecondaryZoneId *uuid.UUID `json:"secondary_zone_id,omitempty"`
}

func (api *APIClient) GetZone(id string) (*InfraZoneGetResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&InfraZoneGetResponse{}).
		Get(fmt.Sprintf("%s/user/infra/zone/%s", api.pathPrefix, id))

	return handleAPIResponse[InfraZoneGetResponse](resp, err)
}

func (api *APIClient) GetZones(
	filterRegionId *string, filterNameIlike *string, skip int, count int,
) ([]InfraZoneGetResponse, error) {
	params := map[string]string{
		"skip":  strconv.FormatInt(int64(skip), 10),
		"count": strconv.FormatInt(int64(count), 10),
	}
	setStrIfNotNil(params, "filter_region_id", filterRegionId)
	setStrIfNotNil(params, "filter_name_ilike", filterNameIlike)

	resp, err := api.restyClient.R().
		SetResult(&[]InfraZoneGetResponse{}).
		SetQueryParams(params).
		Get(fmt.Sprintf("%s/user/infra/zone", api.pathPrefix))

	return handleListAPIResponse[InfraZoneGetResponse](resp, err)
}
