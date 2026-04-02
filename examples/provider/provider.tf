variable "eci_api_token" {
  description = "ECI API access token"
  type        = string
  sensitive   = true
}

provider "eci" {
  api_endpoint     = "https://portal.elice.cloud/api/"
  api_access_token = var.eci_api_token
  zone_id          = "cb67250d-0050-44fa-9872-c8dd7fb9e614"
}
