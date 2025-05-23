terraform {
  required_providers {
    eci = {
      source = "elice-dev/eci"
    }
  }
}

provider "eci" {
  api_endpoint = "https://portal.elice.cloud/api"
  api_access_token = "uIfcvZuf9d6s9q5_hnngnL2Ylqnts1O4lPDFcQzVmBeQ"
  zone_id="cb67250d-0050-44fa-9872-c8dd7fb9e614"
}



resource "eci_block_storage_snapshot" "my_block_storage_snapshot" {
  block_storage_id="5ef1973e-34f7-4b38-a9eb-bcd86b165d57"
  name="block-storage-snapshot-1"
  tags = {
    "created-by": "terraform"
  }
}