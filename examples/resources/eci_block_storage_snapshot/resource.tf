resource "eci_block_storage_snapshot" "my_snapshot" {
  name             = "my-snapshot"
  block_storage_id = eci_block_storage.boot_disk.id
  tags = {
    "created-by" = "terraform"
  }
}
