package api

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type PricingAmount struct {
	PerHour       string `json:"per_hour"        tfsdk:"per_hour"`
	ListedPerHour string `json:"listed_per_hour" tfsdk:"listed_per_hour"`
}

type PricingGetResponse struct {
	Id                 uuid.UUID                `json:"id"`
	Tags               map[string]string        `json:"tags"`
	Created            time.Time                `json:"created"`
	Modified           *time.Time               `json:"modified,omitempty"`
	OrganizationId     *uuid.UUID               `json:"organization_id,omitempty"`
	ZoneId             uuid.UUID                `json:"zone_id"`
	ResourceKind       string                   `json:"resource_kind"`
	ResourceId         *uuid.UUID               `json:"resource_id,omitempty"`
	Name               string                   `json:"name"`
	PricingType        string                   `json:"pricing_type"`
	PricePerHour       string                   `json:"price_per_hour"`
	ListedPricePerHour string                   `json:"listed_price_per_hour"`
	Amount             map[string]PricingAmount `json:"amount,omitempty"`
	Start              *time.Time               `json:"start,omitempty"`
	End                *time.Time               `json:"end,omitempty"`
	Activated          bool                     `json:"activated"`
	Quota              *int                     `json:"quota,omitempty"`
}

func (api *APIClient) GetPricings(
	filterNameIlike *string,
	filterResourceKind *string,
	filterPricingType *string,
	filterActivated *bool,
	skip int,
	count int,
) ([]PricingGetResponse, error) {
	queryParams := map[string]string{}

	if filterNameIlike != nil {
		queryParams["filter_name_ilike"] = *filterNameIlike
	}

	if filterResourceKind != nil {
		queryParams["filter_resource_kind"] = *filterResourceKind
	}

	if filterPricingType != nil {
		queryParams["filter_pricing_type"] = *filterPricingType
	}

	if filterActivated != nil {
		if *filterActivated {
			queryParams["filter_activated"] = "true"
		} else {
			queryParams["filter_activated"] = "false"
		}
	}

	queryParams["skip"] = fmt.Sprintf("%d", skip)
	queryParams["count"] = fmt.Sprintf("%d", count)

	resp, err := api.restyClient.R().
		SetResult(&[]PricingGetResponse{}).
		SetQueryParams(queryParams).
		Get(fmt.Sprintf("%s/user/pricing", api.pathPrefix))

	return handleListAPIResponse[PricingGetResponse](resp, err)
}
