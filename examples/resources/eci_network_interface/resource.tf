resource "eci_network_interface" "my_nic" {
  name                = "my-nic"
  attached_subnet_id  = eci_subnet.my_subnet.id
  attached_machine_id = eci_virtual_machine.my_vm.id
  dr                  = false
  tags = {
    "created-by" = "terraform"
  }
}
