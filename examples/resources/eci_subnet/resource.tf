resource "eci_subnet" "my_subnet" {
  name                = "my-subnet"
  attached_network_id = eci_virtual_network.my_network.id
  purpose             = "virtual_machine"
  network_gw          = "192.168.0.1/24"
  tags = {
    "created-by" = "terraform"
  }
}
