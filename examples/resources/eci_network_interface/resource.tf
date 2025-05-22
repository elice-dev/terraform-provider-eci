resource "eci_network_interface" "my_network_interface" {
  attached_subnet_id="79169e74-7c87-4fa6-8ef4-3d4446dbeb50"
  attached_machine_id="02d41f09-6efa-487c-81a5-f40c9ac996c5"
  name="terraform-network-interace-1"
  dr=false
  tags = {
    "created-by": "terraform"
  }
}