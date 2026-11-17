package api

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ControlPlanePricingIds struct {
	VmAllocation string `json:"vm_allocation"`
	BlockStorage string `json:"block_storage"`
	PublicIp     string `json:"public_ip"`
	LoadBalancer string `json:"load_balancer"`
}

type ControlPlaneSpec struct {
	ImageId        string                 `json:"image_id"`
	StorageSizeGib int64                  `json:"storage_size_gib"`
	PricingIds     ControlPlanePricingIds `json:"pricing_ids"`
}

type ResourceKubernetesClusterGetResponse struct {
	Id               uuid.UUID         `json:"id"`
	Tags             map[string]string `json:"tags"`
	Created          time.Time         `json:"created"`
	Modified         *time.Time        `json:"modified,omitempty"`
	ZoneId           uuid.UUID         `json:"zone_id"`
	OrganizationId   uuid.UUID         `json:"organization_id"`
	Name             string            `json:"name"`
	TargetVersion    string            `json:"target_version"`
	CurrentVersion   string            `json:"current_version"`
	AttachedSubnetId uuid.UUID         `json:"attached_subnet_id"`
	TargetReplicas   int64             `json:"target_replicas"`
	CurrentReplicas  int64             `json:"current_replicas"`
	ControlPlaneSpec ControlPlaneSpec  `json:"control_plane_spec"`
	Username         string            `json:"username"`
	Endpoint         *string           `json:"endpoint,omitempty"`
	Ready            *time.Time        `json:"ready,omitempty"`
	Deleted          *time.Time        `json:"deleted,omitempty"`
	Status           string            `json:"status"`
}

type ResourceKubernetesClusterPostResponse struct {
	Id uuid.UUID `json:"id"`
}

type ResourceKubernetesClusterPatchResponse struct {
	Id uuid.UUID `json:"id"`
}

type ResourceKubernetesClusterDeleteResponse struct {
	Id     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

func (api *APIClient) GetKubernetesCluster(
	id string,
) (*ResourceKubernetesClusterGetResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceKubernetesClusterGetResponse{}).
		Get(fmt.Sprintf("%s/user/resource/compute/kubernetes_cluster/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceKubernetesClusterGetResponse](resp, err)
}

func (api *APIClient) PostKubernetesCluster(
	name string,
	targetVersion string,
	attachedSubnetId string,
	targetReplicas int64,
	controlPlaneSpec ControlPlaneSpec,
	username string,
	password string,
	tags map[string]string,
) (*ResourceKubernetesClusterPostResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceKubernetesClusterPostResponse{}).
		SetBody(map[string]interface{}{
			"zone_id":            api.ZoneId,
			"name":               name,
			"target_version":     targetVersion,
			"attached_subnet_id": attachedSubnetId,
			"target_replicas":    targetReplicas,
			"control_plane_spec": controlPlaneSpec,
			"username":           username,
			"password":           password,
			"tags":               tags,
		}).
		Post(fmt.Sprintf("%s/user/resource/compute/kubernetes_cluster", api.pathPrefix))

	return handleAPIResponse[ResourceKubernetesClusterPostResponse](resp, err)
}

func (api *APIClient) PatchKubernetesCluster(
	id string,
	targetReplicasPtr *int64,
	tagsPtr *map[string]string,
) (*ResourceKubernetesClusterPatchResponse, error) {
	params := map[string]interface{}{}
	setIfNotNil(params, "target_replicas", targetReplicasPtr)
	setIfNotNil(params, "tags", tagsPtr)

	resp, err := api.restyClient.R().
		SetResult(&ResourceKubernetesClusterPatchResponse{}).
		SetBody(params).
		Patch(fmt.Sprintf("%s/user/resource/compute/kubernetes_cluster/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceKubernetesClusterPatchResponse](resp, err)
}

func (api *APIClient) DeleteKubernetesCluster(
	id string,
) (*ResourceKubernetesClusterDeleteResponse, error) {
	resp, err := api.restyClient.R().
		SetResult(&ResourceKubernetesClusterDeleteResponse{}).
		Delete(fmt.Sprintf("%s/user/resource/compute/kubernetes_cluster/%s", api.pathPrefix, id))

	return handleAPIResponse[ResourceKubernetesClusterDeleteResponse](resp, err)
}
