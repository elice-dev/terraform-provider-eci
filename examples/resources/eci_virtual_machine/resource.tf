data "eci_instance_type" "standard" {
  name = "C-2"
}

data "eci_pricing" "vm" {
  name         = "C-2"
  pricing_type = "ondemand"
}

variable "vm_password" {
  description = "Password for the VM user"
  type        = string
  sensitive   = true
}

resource "eci_virtual_machine" "my_vm" {
  name             = "my-vm"
  instance_type_id = data.eci_instance_type.standard.id
  pricing_id       = data.eci_pricing.vm.id
  always_on        = false
  username         = "elice"
  password         = var.vm_password
  on_init_script   = "#!/bin/bash\necho 'Hello' > /home/elice/hello.txt"
  dr               = false
  tags = {
    "created-by" = "terraform"
  }
}
