---
page_title: "Hetzner Cloud: hcloud_placement_groups"
description: |-
  Provides details about multiple Hetzner Cloud Placement Groups.
---

# hcloud_placement_groups

Provides details about multiple Hetzner Cloud Placement Groups.

## Example Usage

```terraform
data "hcloud_placement_groups" "sample_placement_group_1" {

}

data "hcloud_placement_groups" "sample_placement_group_2" {
  with_selector = "key=value"
}
```

## Argument Reference

- `with_selector` - (Optional, string) [Label selector](https://docs.hetzner.cloud/reference/cloud#label-selector)
- `most_recent` - (Optional, bool) Sorts list by date.

## Attribute Reference

- `placement_groups` - (list) List of all matching placement groups. See `data.hcloud_placement_group` for schema.
