resource "eci_virtual_machine_allocation" "my_vm_allocation" {
  machine_id ="d0ba1aed-1414-4388-9c2a-9083ae3154d2"
  tags = {
    "created-by": "terraform"
  }
}