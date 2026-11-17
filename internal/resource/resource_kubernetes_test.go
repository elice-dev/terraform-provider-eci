package resource_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func buildSpec(specType tftypes.Object, storageGib int64) tftypes.Value {
	pricingType := specType.AttributeTypes["pricing_ids"].(tftypes.Object)
	pricing := map[string]tftypes.Value{
		"vm_allocation": str("p-vm"),
		"block_storage": str("p-bs"),
		"public_ip":     str("p-ip"),
	}
	if _, ok := pricingType.AttributeTypes["load_balancer"]; ok {
		pricing["load_balancer"] = str("p-lb")
	}
	return object(specType, map[string]tftypes.Value{
		"image_id":         str("img-1"),
		"storage_size_gib": num(storageGib),
		"pricing_ids":      object(pricingType, pricing),
	})
}

type k8sOverrides map[string]tftypes.Value

func buildCluster(o k8sOverrides, storageGib int64) func(tftypes.Object) tftypes.Value {
	return func(typ tftypes.Object) tftypes.Value {
		attrs := map[string]tftypes.Value{
			"tags":               strMap(map[string]string{}),
			"name":               str("prod-1"),
			"target_version":     str("v1.36.0"),
			"attached_subnet_id": str("subnet-1"),
			"target_replicas":    num(3),
			"control_plane_spec": buildSpec(
				typ.AttributeTypes["control_plane_spec"].(tftypes.Object), storageGib,
			),
			"username": str("admin"),
			"password": str("s3cret"),
		}
		for k, v := range o {
			attrs[k] = v
		}
		return object(typ, attrs)
	}
}

func buildNodePool(o k8sOverrides) func(tftypes.Object) tftypes.Value {
	return func(typ tftypes.Object) tftypes.Value {
		attrs := map[string]tftypes.Value{
			"tags":            strMap(map[string]string{}),
			"cluster_id":      str("cluster-1"),
			"name":            str("workers"),
			"target_replicas": num(4),
			"node_pool_spec": buildSpec(
				typ.AttributeTypes["node_pool_spec"].(tftypes.Object), 100,
			),
			"username": str("_svc-user"),
			"password": str("s3cret"),
		}
		for k, v := range o {
			attrs[k] = v
		}
		return object(typ, attrs)
	}
}

func TestKubernetesClusterConfig(t *testing.T) {
	assertValid(t, validateConfig(t, "eci_kubernetes_cluster", buildCluster(nil, 50)))
	assertValid(t, validateConfig(t, "eci_kubernetes_cluster",
		buildCluster(k8sOverrides{"target_replicas": num(1)}, 10)))

	cases := []struct {
		name       string
		o          k8sOverrides
		storageGib int64
		want       string
	}{
		{"upper case name", k8sOverrides{"name": str("Prod")}, 50, "lower case"},
		{"name ends with dash", k8sOverrides{"name": str("prod-")}, 50, "lower case"},
		{
			"name too long",
			k8sOverrides{"name": str("a234567890123456789012345678901234")},
			50,
			"length",
		},
		{"version without v", k8sOverrides{"target_version": str("1.36.0")}, 50, "v1.36.0"},
		{"version minor only", k8sOverrides{"target_version": str("v1.36")}, 50, "v1.36.0"},
		{"two replicas", k8sOverrides{"target_replicas": num(2)}, 50, "must be one of"},
		{"username starts with digit", k8sOverrides{"username": str("1admin")}, 50, "must start"},
		{"empty password", k8sOverrides{"password": str("")}, 50, "length"},
		{"storage too small", nil, 9, "between 10 and 32768"},
		{"storage too large", nil, 32769, "between 10 and 32768"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertInvalid(
				t,
				validateConfig(
					t,
					"eci_kubernetes_cluster",
					buildCluster(tc.o, tc.storageGib),
				),
				tc.want,
			)
		})
	}
}

func TestKubernetesNodePoolConfig(t *testing.T) {
	assertValid(t, validateConfig(t, "eci_kubernetes_node_pool", buildNodePool(nil)))
	assertValid(t, validateConfig(t, "eci_kubernetes_node_pool",
		buildNodePool(k8sOverrides{"target_replicas": num(64)})))

	cases := []struct {
		name string
		o    k8sOverrides
		want string
	}{
		{"zero replicas", k8sOverrides{"target_replicas": num(0)}, "between 1 and 64"},
		{"too many replicas", k8sOverrides{"target_replicas": num(65)}, "between 1 and 64"},
		{"name with underscore", k8sOverrides{"name": str("my_pool")}, "lower case"},
		{
			"target_version is not configurable",
			k8sOverrides{"target_version": str("v1.36.0")},
			"Read-Only Attribute",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertInvalid(
				t,
				validateConfig(t, "eci_kubernetes_node_pool", buildNodePool(tc.o)),
				tc.want,
			)
		})
	}
}
