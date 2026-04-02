# VM pricing
data "eci_pricing" "vm" {
  name         = "C-2"
  pricing_type = "ondemand"
}

# Block storage pricing
data "eci_pricing" "storage" {
  name         = "Block Storage"
  pricing_type = "ondemand"
}

# Public IP pricing
data "eci_pricing" "ip" {
  name         = "Public IP"
  pricing_type = "ondemand"
}
