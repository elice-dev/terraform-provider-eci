output "vm_id_to_public_ip" {
  value = {
    for vm in module.virtual_machines :
    vm.id => {
      name = vm.name
      ip   = vm.public_ip.ip
    }
  }
}
