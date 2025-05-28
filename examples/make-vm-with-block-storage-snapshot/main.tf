terraform {
  required_providers {
    eci = {
      source = "elice-dev/eci"
    }
  }
}

provider "eci" {
  api_endpoint = "https://portal.elice.cloud/api"
  api_access_token = "ucGKWnD5OfS3PfQ79PR6dHmwRN3Ia18FpcxzIuBM6vX8"
  zone_id="cb67250d-0050-44fa-9872-c8dd7fb9e614"
}

data "eci_block_storage_image" "ubuntu2204" {
  name="Ubuntu 22.04 LTS (20250116)"
}

data "eci_region" "central_01" {
  name="central-01"
}

data "eci_zone" "central_01_a" {
  name="central-01-a"
  region_id="${data.eci_region.central_01.id}"
}

data "eci_instance_type" "m8" {
  name="M-8"
}

resource "eci_virtual_machine" "my_virtual_machine2" {
  name="terraform-test-vm-2"
  instance_type_id="${data.eci_instance_type.m8.id}"
  always_on=false
  username="elice"
  password="secretpassword1!"
  on_init_script="echo 'Hello World!'"
  dr=false
  tags = {
    "created-by": "terraform"
  }
}

resource "eci_virtual_machine_allocation" "my_vm_allocation2" {
  machine_id ="${eci_virtual_machine.my_virtual_machine2.id}"
  tags = {
    "created-by": "terraform"
  }
  depends_on = [  
    eci_block_storage.my_block_storage2,
    eci_network_interface.my_network_interface_two,

     eci_public_ip.my_public_ip2,
  ]
}


resource "eci_block_storage" "my_block_storage2" {
  attached_machine_id="${eci_virtual_machine.my_virtual_machine2.id}"
  name="terraform-test-2"
  dr=false
  size_gib=40
  snapshot_id="a72b896d-7a51-4914-b0f0-c094c3b69ab0"
  tags = {
    "created-by": "terraform"
  }
}

resource "eci_network_interface" "my_network_interface_two" {
  attached_subnet_id="01a3687f-41e9-4b92-9222-98fce2e7de7b"
  attached_machine_id="${eci_virtual_machine.my_virtual_machine2.id}"
  name="terraform-network-interace-1"
  dr=false
  tags = {
    "created-by": "terraform"
  }
}

resource "eci_public_ip" "my_public_ip2" {
  attached_network_interface_id=(
    "${eci_network_interface.my_network_interface_two.id}"
  )
  dr=false
  tags = {
    "created-by": "terraform"
  }
}

output "instance_public_ip_addr_2" {
  value = "${eci_public_ip.my_public_ip2.ip}"
  description = "The public IP address of the virtual machine"
}

