variable "instance_type_id" {
  type = string
}

variable "block_storage_image_id" {
  type = string
}

variable "block_storage_size" {
  type    = number
  default = 40
}

variable "subnet_id" {
  type = string
}

variable "name" {
  description = "Name prefix for resources"
  type        = string
  default     = "elice"
}

variable "tags" {
  description = "Tags to apply"
  type        = map(string)
  default     = {}
}

variable "username" {
  type    = string
  default = "elice"
}

variable "password" {
  type      = string
  default   = "secretpassword1!"
  sensitive = true
}

variable "init_script" {
  type    = string
  default = ""
}

variable "always_on" {
  type    = bool
  default = false
}

variable "dr" {
  type    = bool
  default = false
}