resource "eci_public_ip" "my_public_ip" {
  attached_network_interface_id=(
    "4adf2682-d8f3-451c-8bfc-3383deb424a5"
  )
  dr=false
  tags = {
    "created-by": "terraform"
  }
}