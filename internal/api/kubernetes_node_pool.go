package api

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type NodePoolPricingIds struct {
	VmAllocation string `json:"vm_allocation"`
	BlockStorage string `json:"block_storage"`
	PublicIp     string `json:"public_ip"`
}

type NodePoolSpec struct {
	ImageId        string             `json:"image_id"`
	StorageSizeGib int64              `json:"storage_size_gib"`
	PricingIds     NodePoolPricingIds `json:"pricing_ids"`
}

type ResourceKubernetesNodePoolGetResponse struct {
	Id              uuid.UUID         `json:"id"`
	Tags            map[string]string `json:"tags"`
	Created         time.Time         `json:"created"`
	Modified        *time.Time        `json:"modified,omitempty"`
	ZoneId          uuid.UUID         `json:"zone_id"`
	OrganizationId  uuid.UUID         `json:"organization_id"`
	ClusterId       uuid.UUID         `json:"cluster_id"`
	Name            string            `json:"name"`
	TargetVersion   string            `json:"target_version"`
	CurrentVersion  string            `json:"current_version"`
	TargetReplicas  int64             `json:"target_replicas"`
	CurrentReplicas int64             `json:"current_replicas"`
	NodePoolSpec    NodePoolSpec      `json:"node_pool_spec"`
	Username        string            `json:"username"`
	Ready           *time.Time        `json:"ready,omitempty"`
	Deleted         *time.Time        `json:"deleted,omitempty"`
	Status          string            `json:"status"`
}

type ResourceKubernetesNodePoolPostResponse struct {
	Id uuid.UUID `json:"id"`
}

type ResourceKubernetesNodePoolPatchResponse struct {
	Id uuid.UUID `json:"id"`
}

type ResourceKubernetesNodePoolDeleteResponse struct {
	Id     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

func (api *APIClient) GetKubernetesNodePool(
	id string,
) (*ResourceKubernetesNodePoolGetResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceKubernetesNodePoolGetResponse{}).
		Get(fmt.Sprintf("%s/user/resource/compute/kubernetes_node_pool/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceKubernetesNodePoolGetResponse](resp, err)
}

func (api *APIClient) PostKubernetesNodePool(
	clusterId string,
	name string,
	targetReplicas int64,
	nodePoolSpec NodePoolSpec,
	username string,
	password string,
	tags map[string]string,
) (*ResourceKubernetesNodePoolPostResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceKubernetesNodePoolPostResponse{}).
		SetBody(map[string]interface{}{
			"zone_id":         api.ZoneId,
			"cluster_id":      clusterId,
			"name":            name,
			"target_replicas": targetReplicas,
			"node_pool_spec":  nodePoolSpec,
			"username":        username,
			"password":        password,
			"tags":            tags,
		}).
		Post(fmt.Sprintf("%s/user/resource/compute/kubernetes_node_pool", api.pathPrefix))

	return handleAPIResponse[ResourceKubernetesNodePoolPostResponse](resp, err)
}

func (api *APIClient) PatchKubernetesNodePool(
	id string,
	targetReplicasPtr *int64,
	tagsPtr *map[string]string,
) (*ResourceKubernetesNodePoolPatchResponse, error) {
	params := map[string]interface{}{}
	setIfNotNil(params, "target_replicas", targetReplicasPtr)
	setIfNotNil(params, "tags", tagsPtr)

	resp, err := api.restyClient.R().
		SetResult(&ResourceKubernetesNodePoolPatchResponse{}).
		SetBody(params).
		Patch(fmt.Sprintf("%s/user/resource/compute/kubernetes_node_pool/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceKubernetesNodePoolPatchResponse](resp, err)
}

func (api *APIClient) DeleteKubernetesNodePool(
	id string,
) (*ResourceKubernetesNodePoolDeleteResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceKubernetesNodePoolDeleteResponse{}).
		Delete(fmt.Sprintf("%s/user/resource/compute/kubernetes_node_pool/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceKubernetesNodePoolDeleteResponse](resp, err)
}
