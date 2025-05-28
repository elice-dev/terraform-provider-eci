terraform {
  required_providers {
    eci = {
      source = "elice-dev/eci"
    }
  }
}

provider "eci" {
  api_endpoint = "https://portal.elice.cloud/api"
  api_access_token = "ucGKWnD5OfS3PfQ79PR6dHmwRN3Ia18FpcxzIuBM6vX8"
  zone_id="cb67250d-0050-44fa-9872-c8dd7fb9e614"
}



resource "eci_block_storage_snapshot" "my_block_storage_snapshot" {
  block_storage_id="5ef1973e-34f7-4b38-a9eb-bcd86b165d57"
  name="block-storage-snapshot-1"
  tags = {
    "created-by": "terraform"
  }
}
output "block_storage_snapshot_id" {
  value = eci_block_storage_snapshot.my_block_storage_snapshot.id
  description = "value of the block storage snapshot id"
}