data "eci_block_storage_image" "ubuntu" {
  name = "Ubuntu 22.04 LTS (20250116)"
}

resource "eci_block_storage" "boot_disk" {
  # ...
  image_id = data.eci_block_storage_image.ubuntu.id
}
