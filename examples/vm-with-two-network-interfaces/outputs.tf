output "instance_ip_addr_1" {
  value = "${eci_network_interface.my_network_interface_one.ip}"
  description = "The private IP address of the virtual machine"
}

output "instance_mac_addr_1" {
  value = "${eci_network_interface.my_network_interface_one.mac}"
  description = "The MAC address of the virtual machine"
}


output "instance_ip_addr_2" {
  value = "${eci_network_interface.my_network_interface_two.ip}"
  description = "The private IP address of the virtual machine"
}

output "instance_mac_addr_2" {
  value = "${eci_network_interface.my_network_interface_two.mac}"
  description = "The MAC address of the virtual machine"
}

output "block_storage_image_name" {
  value = "${data.eci_block_storage_image.ubuntu2204.name}"
}