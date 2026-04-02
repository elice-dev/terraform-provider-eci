data "eci_block_storage_image" "ubuntu" {
  name = "Ubuntu 22.04 LTS (20250116)"
}

data "eci_pricing" "storage" {
  name         = "Block Storage"
  pricing_type = "ondemand"
}

resource "eci_block_storage" "boot_disk" {
  name                = "my-boot-disk"
  attached_machine_id = eci_virtual_machine.my_vm.id
  pricing_id          = data.eci_pricing.storage.id
  image_id            = data.eci_block_storage_image.ubuntu.id
  size_gib            = 40
  dr                  = false
  tags = {
    "created-by" = "terraform"
  }
}
