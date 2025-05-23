resource "eci_block_storage_snapshot" "my_block_storage_snapshot" {
  block_storage_id="d2ecd04e-b261-4af0-a0ca-e20f31109981"
  name="block-storage-snapshot-1"
  tags = {
    "created-by": "terraform"
  }
}