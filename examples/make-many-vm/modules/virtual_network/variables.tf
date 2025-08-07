variable "name" {
  description = "Name prefix for resources"
  type        = string
  default     = "elice"
}

variable "network_cidr" {
  description = "CIDR block for the internal network range"
  type        = string
  default     = "192.168.0.0/16"
}

variable "firewall_rules" {
  description = "List of default firewall rule settings for network traffic"
  type = list(object({
    proto       = string
    source      = string
    destination = string
    port        = number
    port_end    = number
    action      = string
    comment     = string
  }))
  default = [
    {
      proto       = "ALL"
      source      = "0.0.0.0/0"
      destination = "0.0.0.0/0"
      port        = 0
      port_end    = 65535
      action      = "ACCEPT"
      comment     = "sample network rule"
    }
  ]
}

variable "network_gw" {
  description = "Gateway IP address with subnet mask for the internal network"
  type        = string
  default     = "192.168.0.1/24"
}


variable "tags" {
  description = "Tags to apply"
  type        = map(string)
  default     = {}
}

variable "dr" {
  type    = bool
  default = false
}