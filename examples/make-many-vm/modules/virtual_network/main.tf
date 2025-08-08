terraform {
  required_providers {
    eci = {
      source = "elice-dev/eci"
    }
  }
}

resource "eci_virtual_network" "virtual_network" {
  name           = "${var.name}-virtual-network"
  network_cidr   = var.network_cidr
  firewall_rules = var.firewall_rules
  tags           = var.tags
}

resource "eci_subnet" "subnet" {
  name                = "${var.name}-subnet"
  attached_network_id = eci_virtual_network.virtual_network.id
  purpose             = "virtual_machine"
  network_gw          = var.network_gw
  tags                = var.tags
}
