data "eci_pricing" "vm_pricing" {
  name         = "M-8"
  pricing_type = "ondemand"
}

data "eci_pricing" "storage_pricing" {
  name         = "Block Storage"
  pricing_type = "ondemand"
}

data "eci_pricing" "ip_pricing" {
  name         = "Public IP"
  pricing_type = "ondemand"
}

data "eci_pricing" "lb_pricing" {
  name         = "Load Balancer"
  pricing_type = "ondemand"
}

resource "eci_kubernetes_cluster" "my_cluster" {
  name               = "my-cluster"
  target_version     = "v1.36.0"
  attached_subnet_id = "9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d"
  target_replicas    = 3

  control_plane_spec = {
    image_id         = "2c4e6a8b-0d1f-4e3a-9b5c-7d9e1f3a5b7c"
    storage_size_gib = 50
    pricing_ids = {
      vm_allocation = data.eci_pricing.vm_pricing.id
      block_storage = data.eci_pricing.storage_pricing.id
      public_ip     = data.eci_pricing.ip_pricing.id
      load_balancer = data.eci_pricing.lb_pricing.id
    }
  }

  username = "elice"
  password = var.node_password

  tags = {
    "created-by" : "terraform"
  }
}

variable "node_password" {
  type      = string
  sensitive = true
}
