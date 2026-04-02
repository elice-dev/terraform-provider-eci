resource "eci_virtual_network" "my_network" {
  name         = "my-network"
  network_cidr = "192.168.0.0/16"
  firewall_rules = [
    {
      proto       = "TCP"
      source      = "0.0.0.0/0"
      destination = "0.0.0.0/0"
      port        = 22
      port_end    = 22
      action      = "ACCEPT"
      comment     = "Allow SSH"
    },
    {
      proto       = "TCP"
      source      = "0.0.0.0/0"
      destination = "0.0.0.0/0"
      port        = 80
      port_end    = 443
      action      = "ACCEPT"
      comment     = "Allow HTTP/HTTPS"
    },
  ]
  tags = {
    "created-by" = "terraform"
  }
}
