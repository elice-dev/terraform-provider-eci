resource "eci_subnet" "my_subnet" {
  name="terraform-test-subnet-1"
  attached_network_id="02d41f09-6efa-487c-81a5-f40c9ac996c5"
  purpose="virtual_machine"
  network_gw="192.168.0.1/24"
  tags = {
    "created-by": "terraform"
  }
}