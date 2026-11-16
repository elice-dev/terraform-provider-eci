package api

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type LoadBalancerHealthCheck struct {
	Proto              string  `json:"proto"`
	Port               int64   `json:"port"`
	Interval           int64   `json:"interval"`
	Timeout            int64   `json:"timeout"`
	HealthyThreshold   int64   `json:"healthy_threshold"`
	UnhealthyThreshold int64   `json:"unhealthy_threshold"`
	Path               *string `json:"path,omitempty"`
}

type LoadBalancerRuleServer struct {
	Ip                 string `json:"ip"`
	NetworkInterfaceId string `json:"network_interface_id"`
	Port               int64  `json:"port"`
}

type LoadBalancerRule struct {
	Proto       string                   `json:"proto"`
	Port        int64                    `json:"port"`
	Servers     []LoadBalancerRuleServer `json:"servers"`
	Healthcheck LoadBalancerHealthCheck  `json:"healthcheck"`
	Algorithm   string                   `json:"algorithm"`
}

type ResourceLoadBalancerGetResponse struct {
	Id                  uuid.UUID          `json:"id"`
	Tags                map[string]string  `json:"tags"`
	Created             time.Time          `json:"created"`
	Modified            *time.Time         `json:"modified,omitempty"`
	ZoneId              uuid.UUID          `json:"zone_id"`
	OrganizationId      uuid.UUID          `json:"organization_id"`
	Name                string             `json:"name"`
	ShortId             string             `json:"short_id"`
	AttachedNetworkId   uuid.UUID          `json:"attached_network_id"`
	KubernetesClusterId *uuid.UUID         `json:"kubernetes_cluster_id,omitempty"`
	Rules               []LoadBalancerRule `json:"rules"`
	PricingType         string             `json:"pricing_type"`
	PricingId           uuid.UUID          `json:"pricing_id"`
	Activated           *time.Time         `json:"activated,omitempty"`
	Deleted             *time.Time         `json:"deleted,omitempty"`
	Status              string             `json:"status"`
}

type ResourceLoadBalancerPostResponse struct {
	Id uuid.UUID `json:"id"`
}

type ResourceLoadBalancerPatchResponse struct {
	Id uuid.UUID `json:"id"`
}

type ResourceLoadBalancerDeleteResponse struct {
	Id     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

func (api *APIClient) GetLoadBalancer(id string) (*ResourceLoadBalancerGetResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceLoadBalancerGetResponse{}).
		Get(fmt.Sprintf("%s/user/resource/network/load_balancer/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceLoadBalancerGetResponse](resp, err)
}

func (api *APIClient) PostLoadBalancer(
	name string,
	attachedNetworkId string,
	publicIpId string,
	rules []LoadBalancerRule,
	pricingId string,
	tags map[string]string,
) (*ResourceLoadBalancerPostResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceLoadBalancerPostResponse{}).
		SetBody(map[string]interface{}{
			"zone_id":             api.ZoneId,
			"name":                name,
			"attached_network_id": attachedNetworkId,
			"public_ip_id":        publicIpId,
			"rules":               rules,
			"pricing_id":          pricingId,
			"tags":                tags,
		}).
		Post(fmt.Sprintf("%s/user/resource/network/load_balancer", api.pathPrefix))

	return handleAPIResponse[ResourceLoadBalancerPostResponse](resp, err)
}

func (api *APIClient) PatchLoadBalancer(
	id string,
	namePtr *string,
	rulesPtr *[]LoadBalancerRule,
	tagsPtr *map[string]string,
) (*ResourceLoadBalancerPatchResponse, error) {
	params := map[string]interface{}{}
	setIfNotNil(params, "name", namePtr)
	setIfNotNil(params, "rules", rulesPtr)
	setIfNotNil(params, "tags", tagsPtr)

	resp, err := api.restyClient.R().
		SetResult(&ResourceLoadBalancerPatchResponse{}).
		SetBody(params).
		Patch(fmt.Sprintf("%s/user/resource/network/load_balancer/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceLoadBalancerPatchResponse](resp, err)
}

func (api *APIClient) DeleteLoadBalancer(id string) (*ResourceLoadBalancerDeleteResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceLoadBalancerDeleteResponse{}).
		Delete(fmt.Sprintf("%s/user/resource/network/load_balancer/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceLoadBalancerDeleteResponse](resp, err)
}
