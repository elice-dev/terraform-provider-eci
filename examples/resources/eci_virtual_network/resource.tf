resource "eci_virtual_network" "my_virtual_network" {
  name="terraform-test-virtual-network_ii"
  network_cidr="192.168.0.0/16"
  firewall_rules= []
  tags = {
    "created-by": "terraform"
  }
}