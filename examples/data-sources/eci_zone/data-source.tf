data "eci_zone" "zone_test" {
  name="Test-Zone"
  region_id="${data.eci_region.region_seoul1.id}"
}