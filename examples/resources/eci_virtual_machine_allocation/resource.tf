resource "eci_virtual_machine_allocation" "my_vm_alloc" {
  machine_id = eci_virtual_machine.my_vm.id
  tags = {
    "created-by" = "terraform"
  }
  depends_on = [
    eci_block_storage.boot_disk,
    eci_network_interface.my_nic,
    eci_public_ip.my_ip
  ]
}
