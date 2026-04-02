data "eci_region" "central" {
  name = "central-01"
}

data "eci_zone" "zone_a" {
  name      = "central-01-a"
  region_id = data.eci_region.central.id
}
