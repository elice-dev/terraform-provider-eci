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
  name="Ubuntu 22.04"
}

data "eci_region" "test_region" {
  name="Test-Region"
}

data "eci_zone" "test_zone" {
  name="Test-Zone"
  region_id="${data.eci_region.test_region.id}"
}

data "eci_instance_type" "test_instance_type" {
  name="M-8"
}

resource "eci_virtual_machine" "my_virtual_machine" {
  name="terraform-test-vm-1"
  instance_type_id="${data.eci_instance_type.test_instance_type.id}"
  always_on=false
  username="elice"
  password="secretpassword1!"
  on_init_script=""
  dr=false
  tags = {
    "created-by": "terraform"
  }
}

resource "eci_virtual_machine_allocation" "my_vm_allocation" {
  machine_id ="${eci_virtual_machine.my_virtual_machine.id}"
  tags = {
    "created-by": "terraform"
  }
  depends_on = [  
    eci_block_storage.my_block_storage,
    eci_network_interface.my_network_interface_one,
    eci_network_interface.my_network_interface_two,
    //   eci_public_ip.my_public_ip,
  ]
}

resource "eci_block_storage" "my_block_storage" {
  attached_machine_id="${eci_virtual_machine.my_virtual_machine.id}"
  name="terraform-test-1"
  dr=false
  size_gib=40
  image_id="${data.eci_block_storage_image.ubuntu2204.id}"
  tags = {
    "created-by": "terraform"
  }
}

resource "eci_virtual_network" "my_virtual_network" {
  name="terraform-test-virtual-network_ii"
  network_cidr="192.168.0.0/16"
  firewall_rules= []
  tags = {
    "created-by": "terraform"
  }
}

resource "eci_subnet" "my_subnet" {
  name="terraform-test-subnet-1"
  attached_network_id="${eci_virtual_network.my_virtual_network.id}"
  purpose="virtual_machine"
  network_gw="192.168.0.1/24"
  tags = {
    "created-by": "terraform"
  }
}

resource "eci_network_interface" "my_network_interface_one" {
  attached_subnet_id="${eci_subnet.my_subnet.id}"
  attached_machine_id="${eci_virtual_machine.my_virtual_machine.id}"
  name="terraform-network-interace-1"
  dr=false
  tags = {
    "created-by": "terraform"
  }
}

resource "eci_network_interface" "my_network_interface_two" {
  attached_subnet_id="${eci_subnet.my_subnet.id}"
  attached_machine_id="${eci_virtual_machine.my_virtual_machine.id}"
  name="terraform-network-interace-2"
  dr=false
  tags = {
    "created-by": "terraform"
  }
}

/*
resource "eci_public_ip" "my_public_ip" {
  attached_network_interface_id=(
    "${eci_network_interface.my_network_interface_one.id}"
  )
  dr=false
  tags = {
    "created-by": "terraform"
  }
}

output "instance_public_ip_addr_1" {
  value = "${eci_public_ip.my_public_ip.ip}"
  description = "The public IP address of the virtual machine"
}

output "instance_public_ip_dr_addr_1" {
  value = "${eci_public_ip.my_public_ip.dr_ip}"
  description = "The public IP address of the virtual machine (DR)"
}
*/