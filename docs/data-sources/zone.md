---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "eci_zone Data Source - eci"
subcategory: ""
description: |-
  Zone
---

# eci_zone (Data Source)

Zone

## Example Usage

```terraform
data "eci_zone" "zone_test" {
  name="Test-Zone"
  region_id="02d41f09-6efa-487c-81a5-f40c9ac996c5"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) human-readable name of the zone
- `region_id` (String) id of the region that the zone belongs to

### Read-Only

- `id` (String) unique identifier of the zone
- `secondary_zone_id` (String) id of the secondary zone that this zone will fail over when DR
