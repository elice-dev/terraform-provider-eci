resource "eci_virtual_network" "my_virtual_network" {
  name="my-virtual-network"
  network_cidr="192.168.0.0/16"
  firewall_rules= []
  tags = {
    "created-by": "terraform"
  }
}