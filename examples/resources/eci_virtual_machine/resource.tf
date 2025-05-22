resource "eci_virtual_machine" "my_virtual_machine" {
  name="my-vm-1"
  instance_type_id="d0ba1aed-1414-4388-9c2a-9083ae3154d2"
  always_on=false
  username="elice"
  password="secretpassword1!"
  dr=false
  tags = {
    "created-by": "terraform"
  }
}
