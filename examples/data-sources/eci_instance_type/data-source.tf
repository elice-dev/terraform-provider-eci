data "eci_instance_type" "standard" {
  name = "C-2"
}

resource "eci_virtual_machine" "my_vm" {
  # ...
  instance_type_id = data.eci_instance_type.standard.id
}
