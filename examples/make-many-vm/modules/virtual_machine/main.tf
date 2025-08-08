terraform {
  required_providers {
    eci = {
      source = "elice-dev/eci"
    }
  }
}

resource "eci_virtual_machine" "virtual_machine" {
  name             = var.name
  instance_type_id = var.instance_type_id
  always_on        = var.always_on
  username         = var.username
  password         = var.password
  on_init_script   = var.init_script
  dr               = var.dr
  tags             = var.tags
}

resource "eci_virtual_machine_allocation" "start_many_vm" {
  machine_id = eci_virtual_machine.virtual_machine.id
  tags       = var.tags
  depends_on = [
    eci_block_storage.block_storage,
    eci_network_interface.network_interface
  ]
}

resource "eci_block_storage" "block_storage" {
  attached_machine_id = eci_virtual_machine.virtual_machine.id
  name                = "${var.name}-block-stroage"
  dr                  = var.dr
  size_gib            = var.block_storage_size
  image_id            = var.block_storage_image_id
  tags                = var.tags
}

resource "eci_network_interface" "network_interface" {
  attached_subnet_id  = var.subnet_id
  attached_machine_id = eci_virtual_machine.virtual_machine.id
  name                = "${var.name}-interace"
  dr                  = var.dr
  tags                = var.tags
}

resource "eci_public_ip" "public_ip" {
  attached_network_interface_id = eci_network_interface.network_interface.id
  dr                            = var.dr
  tags                          = var.tags
}