resource "eci_kubernetes_node_pool" "workers" {
  cluster_id      = eci_kubernetes_cluster.my_cluster.id
  name            = "workers"
  target_replicas = 3

  node_pool_spec = {
    image_id         = "2c4e6a8b-0d1f-4e3a-9b5c-7d9e1f3a5b7c"
    storage_size_gib = 100
    pricing_ids = {
      vm_allocation = data.eci_pricing.vm_pricing.id
      block_storage = data.eci_pricing.storage_pricing.id
      public_ip     = data.eci_pricing.ip_pricing.id
    }
  }

  username = "elice"
  password = var.node_password

  tags = {
    "created-by" : "terraform"
  }
}
