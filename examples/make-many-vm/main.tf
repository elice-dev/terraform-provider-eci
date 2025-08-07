terraform {
  required_providers {
    eci = {
      source = "elice-dev/eci"
    }
  }
}

provider "eci" {
  api_endpoint     = "https://portal.elice.cloud/api"
  api_access_token = "u_Zb0eS2Orcu9Pv8EBfdf9D9aQiANHnqCsNt_Hy3TBIA"
  zone_id          = "cb67250d-0050-44fa-9872-c8dd7fb9e614"
}

data "eci_block_storage_image" "ubuntu2204" {
  name = "Ubuntu 22.04 LTS (20250116)"
}

data "eci_region" "central_01" {
  name = "central-01"
}

variable "virtual_machine_number" {
  type    = number
  default = 3
}

module "virtual_machines" {
  source = "./modules/virtual_machine"

  providers = {
    eci = eci
  }

  count = var.virtual_machine_number

  name                   = "many-terraform-test-${count.index}"
  password               = "Secretpa$$w0rd1!"
  instance_type_id       = data.eci_instance_type.test_instance_type.id
  block_storage_image_id = data.eci_block_storage_image.ubuntu2204.id
  block_storage_size     = 100
  subnet_id              = module.network.subnet_id
}

module "network" {
  source = "./modules/virtual_network"

  providers = {
    eci = eci
  }

  name = "many-terraform-test-virtual-network"
}

