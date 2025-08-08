output "id" {
  description = "ID of the virtual machine"
  value       = eci_virtual_machine.virtual_machine.id
}

output "name" {
  description = "Name of the virtual machine"
  value       = eci_virtual_machine.virtual_machine.name
}

output "public_ip" {
  description = "Public IP assigned to network interfaces"
  value       = eci_public_ip.public_ip
}
