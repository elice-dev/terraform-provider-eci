package api

import (
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type InfraInstanceTypeGetResponse struct {
	Id           uuid.UUID         `json:"id"`
	Tags         map[string]string `json:"tags"`
	Created      time.Time         `json:"created"`
	Modified     *time.Time        `json:"modified,omitempty"`
	ZoneId       uuid.UUID         `json:"zone_id"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	CpuVcore     int               `json:"cpu_vcore"`
	MemoryGib    int               `json:"memory_gib"`
	Devices      []string          `json:"devices"`
	PricePerHour float64           `json:"price_per_hour"`
	Activated    bool              `json:"activated"`
}

func (api *APIClient) GetInstanceType(id string) (*InfraInstanceTypeGetResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&InfraInstanceTypeGetResponse{}).
		Get(fmt.Sprintf("%s/user/infra/instance_type/%s", api.pathPrefix, id))

	return handleAPIResponse[InfraInstanceTypeGetResponse](resp, err)
}

func (api *APIClient) GetInstanceTypes(
	filterNameIlike *string, filterActivated *bool, skip int, count int,
) ([]InfraInstanceTypeGetResponse, error) {
	params := map[string]string{
		"filter_zone_id": api.zoneId,
		"skip":           strconv.FormatInt(int64(skip), 10),
		"count":          strconv.FormatInt(int64(count), 10),
	}
	setStrIfNotNil(params, "filter_name_ilike", filterNameIlike)

	if filterActivated != nil {
		params["filter_activated"] = strconv.FormatBool(*filterActivated)
	}

	resp, err := api.restyClient.R().
		SetResult(&[]InfraInstanceTypeGetResponse{}).
		SetQueryParams(params).
		Get(fmt.Sprintf("%s/user/infra/instance_type", api.pathPrefix))

	return handleListAPIResponse[InfraInstanceTypeGetResponse](resp, err)
}
