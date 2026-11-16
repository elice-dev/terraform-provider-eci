package api

import (
	"net/http"
	"testing"
)

const (
	clusterId  = "77777777-7777-7777-7777-777777777777"
	nodePoolId = "88888888-8888-8888-8888-888888888888"
)

func TestPostKubernetesCluster_SendsFullBody(t *testing.T) {
	client, fs := newFakeClient(t, http.StatusOK, `{"id":"`+clusterId+`"}`)

	spec := ControlPlaneSpec{
		ImageId:        "img-1",
		StorageSizeGib: 50,
		PricingIds: ControlPlanePricingIds{
			VmAllocation: "p-vm", BlockStorage: "p-bs", PublicIp: "p-ip", LoadBalancer: "p-lb",
		},
	}
	res, err := client.PostKubernetesCluster(
		"prod", "v1.36.0", "subnet-1", 3, spec, "admin", "s3cret", map[string]string{},
	)
	if err != nil {
		t.Fatalf("PostKubernetesCluster: %v", err)
	}
	if res.Id.String() != clusterId {
		t.Errorf("Id = %s", res.Id)
	}

	req := fs.only()
	assertRequest(t, req, http.MethodPost, "/user/resource/compute/kubernetes_cluster")
	assertJSONEqual(t, "body", req.Body, `{
		"zone_id": "`+testZoneId+`",
		"name": "prod",
		"target_version": "v1.36.0",
		"attached_subnet_id": "subnet-1",
		"target_replicas": 3,
		"control_plane_spec": {
			"image_id": "img-1", "storage_size_gib": 50,
			"pricing_ids": {"vm_allocation": "p-vm", "block_storage": "p-bs",
			                "public_ip": "p-ip", "load_balancer": "p-lb"}
		},
		"username": "admin",
		"password": "s3cret",
		"tags": {}
	}`)
}

func TestPatchKubernetesCluster_SendsOnlySetFields(t *testing.T) {
	replicas := int64(1)
	tags := map[string]string{"team": "cloud"}

	cases := []struct {
		name     string
		replicas *int64
		tags     *map[string]string
		want     string
	}{
		{"nothing", nil, nil, `{}`},
		{"replicas", &replicas, nil, `{"target_replicas": 1}`},
		{"both", &replicas, &tags, `{"target_replicas": 1, "tags": {"team": "cloud"}}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, fs := newFakeClient(t, http.StatusOK, `{"id":"`+clusterId+`"}`)
			if _, err := client.PatchKubernetesCluster(
				clusterId,
				tc.replicas,
				tc.tags,
			); err != nil {
				t.Fatalf("PatchKubernetesCluster: %v", err)
			}
			req := fs.only()
			assertRequest(
				t,
				req,
				http.MethodPatch,
				"/user/resource/compute/kubernetes_cluster/"+clusterId,
			)
			if req.Body == nil {
				req.Body = map[string]interface{}{}
			}
			assertJSONEqual(t, "body", req.Body, tc.want)
		})
	}
}

func TestGetKubernetesCluster_ParsesResponse(t *testing.T) {
	client, fs := newFakeClient(t, http.StatusOK, `{
		"id": "`+clusterId+`", "tags": {}, "created": "`+testCreated+`", "modified": null,
		"zone_id": "`+testZoneId+`", "organization_id": "`+testOrgId+`", "user_id": null,
		"name": "prod", "target_version": "v1.36.0", "current_version": "v1.35.2",
		"attached_subnet_id": "99999999-9999-9999-9999-999999999999",
		"target_replicas": 3, "current_replicas": 1,
		"control_plane_spec": {"image_id": "img-1", "storage_size_gib": 50,
			"pricing_ids": {"vm_allocation": "p-vm", "block_storage": "p-bs",
			                "public_ip": "p-ip", "load_balancer": "p-lb"}},
		"username": "admin", "endpoint": null, "ready": null, "deleted": null,
		"status": "provisioning"
	}`)

	c, err := client.GetKubernetesCluster(clusterId)
	if err != nil {
		t.Fatalf("GetKubernetesCluster: %v", err)
	}
	assertRequest(
		t,
		fs.only(),
		http.MethodGet,
		"/user/resource/compute/kubernetes_cluster/"+clusterId,
	)

	if c.Status != "provisioning" || c.CurrentVersion != "v1.35.2" || c.CurrentReplicas != 1 {
		t.Errorf("unexpected scalar fields: %+v", c)
	}
	if c.Endpoint != nil || c.Ready != nil {
		t.Error("Endpoint and Ready should be nil before the cluster is ready")
	}
	if c.ControlPlaneSpec.PricingIds.LoadBalancer != "p-lb" ||
		c.ControlPlaneSpec.StorageSizeGib != 50 {
		t.Errorf("unexpected control_plane_spec: %+v", c.ControlPlaneSpec)
	}
}

func TestDeleteKubernetesCluster(t *testing.T) {
	client, fs := newFakeClient(t, http.StatusOK, `{"id":"`+clusterId+`","status":"deleting"}`)

	if _, err := client.DeleteKubernetesCluster(clusterId); err != nil {
		t.Fatalf("DeleteKubernetesCluster: %v", err)
	}
	assertRequest(
		t,
		fs.only(),
		http.MethodDelete,
		"/user/resource/compute/kubernetes_cluster/"+clusterId,
	)
}

func TestPostKubernetesNodePool_SendsFullBody(t *testing.T) {
	client, fs := newFakeClient(t, http.StatusOK, `{"id":"`+nodePoolId+`"}`)

	spec := NodePoolSpec{
		ImageId:        "img-2",
		StorageSizeGib: 100,
		PricingIds: NodePoolPricingIds{
			VmAllocation: "p-vm",
			BlockStorage: "p-bs",
			PublicIp:     "p-ip",
		},
	}
	if _, err := client.PostKubernetesNodePool(
		clusterId, "workers", 4, spec, "admin", "s3cret", map[string]string{"pool": "a"},
	); err != nil {
		t.Fatalf("PostKubernetesNodePool: %v", err)
	}

	req := fs.only()
	assertRequest(t, req, http.MethodPost, "/user/resource/compute/kubernetes_node_pool")
	// node_pool_spec.pricing_ids has no load_balancer key, unlike the control plane.
	assertJSONEqual(t, "body", req.Body, `{
		"zone_id": "`+testZoneId+`",
		"cluster_id": "`+clusterId+`",
		"name": "workers",
		"target_replicas": 4,
		"node_pool_spec": {
			"image_id": "img-2", "storage_size_gib": 100,
			"pricing_ids": {"vm_allocation": "p-vm", "block_storage": "p-bs", "public_ip": "p-ip"}
		},
		"username": "admin",
		"password": "s3cret",
		"tags": {"pool": "a"}
	}`)
}

func TestPatchKubernetesNodePool_SendsOnlySetFields(t *testing.T) {
	replicas := int64(8)
	client, fs := newFakeClient(t, http.StatusOK, `{"id":"`+nodePoolId+`"}`)

	if _, err := client.PatchKubernetesNodePool(nodePoolId, &replicas, nil); err != nil {
		t.Fatalf("PatchKubernetesNodePool: %v", err)
	}
	req := fs.only()
	assertRequest(
		t,
		req,
		http.MethodPatch,
		"/user/resource/compute/kubernetes_node_pool/"+nodePoolId,
	)
	assertJSONEqual(t, "body", req.Body, `{"target_replicas": 8}`)
}

func TestGetKubernetesNodePool_ParsesResponse(t *testing.T) {
	client, fs := newFakeClient(t, http.StatusOK, `{
		"id": "`+nodePoolId+`", "tags": {}, "created": "`+testCreated+`", "modified": null,
		"zone_id": "`+testZoneId+`", "organization_id": "`+testOrgId+`", "user_id": null,
		"cluster_id": "`+clusterId+`", "name": "workers",
		"target_version": "v1.36.0", "current_version": "v1.36.0",
		"target_replicas": 4, "current_replicas": 4,
		"node_pool_spec": {"image_id": "img-2", "storage_size_gib": 100,
			"pricing_ids": {"vm_allocation": "p-vm", "block_storage": "p-bs", "public_ip": "p-ip"}},
		"username": "admin", "ready": "`+testCreated+`", "deleted": null, "status": "ready"
	}`)

	np, err := client.GetKubernetesNodePool(nodePoolId)
	if err != nil {
		t.Fatalf("GetKubernetesNodePool: %v", err)
	}
	assertRequest(
		t,
		fs.only(),
		http.MethodGet,
		"/user/resource/compute/kubernetes_node_pool/"+nodePoolId,
	)

	if np.ClusterId.String() != clusterId || np.Status != "ready" ||
		np.NodePoolSpec.StorageSizeGib != 100 {
		t.Errorf("unexpected node pool: %+v", np)
	}
}

func TestDeleteKubernetesNodePool(t *testing.T) {
	client, fs := newFakeClient(t, http.StatusOK, `{"id":"`+nodePoolId+`","status":"deleting"}`)

	if _, err := client.DeleteKubernetesNodePool(nodePoolId); err != nil {
		t.Fatalf("DeleteKubernetesNodePool: %v", err)
	}
	assertRequest(
		t,
		fs.only(),
		http.MethodDelete,
		"/user/resource/compute/kubernetes_node_pool/"+nodePoolId,
	)
}
