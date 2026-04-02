data "eci_pricing" "ip" {
  name         = "Public IP"
  pricing_type = "ondemand"
}

resource "eci_public_ip" "my_ip" {
  attached_network_interface_id = eci_network_interface.my_nic.id
  pricing_id                    = data.eci_pricing.ip.id
  dr                            = false
  tags = {
    "created-by" = "terraform"
  }
}
