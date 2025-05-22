package api

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type OrganizationGetResponse struct {
	Id         uuid.UUID  `json:"id"`
	Created    time.Time  `json:"created"`
	Modified   *time.Time `json:"modified,omitempty"`
	Name       string     `json:"name"`
	Ident      string     `json:"ident"`
	AllowedIps []string   `json:"allowed_ips"`
}

func (api *APIClient) GetOrganization() (*OrganizationGetResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&OrganizationGetResponse{}).
		Get(fmt.Sprintf("%s/user/organization", api.pathPrefix))

	println(fmt.Sprintf("organization id: %s", resp.String()))
	return handleAPIResponse[OrganizationGetResponse](resp, err)
}
