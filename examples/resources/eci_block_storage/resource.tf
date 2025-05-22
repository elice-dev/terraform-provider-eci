resource "eci_block_storage" "my_block_storage" {
  attached_machine_id="4f3a9eeb-962f-4f9c-9074-13c422b3d726"
  name="my-block-strage"
  dr=false
  size_gib=40
  image_id="4f3a9eeb-962f-4f9c-9074-13c422b3d726"
  tags = {
    "created-by": "terraform"
  }
}
