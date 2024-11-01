---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_load_balancer_types"
sidebar_current: "docs-hcloud-datasource-load-balancer-types"
description: |-
  List all available Hetzner Cloud Load Balancer Types.
---

# Data Source: hcloud_load_balancer_types

Provides a list of available Hetzner Cloud Load Balancer Types.

## Example Usage

```hcl
data "hcloud_load_balancer_types" "all" {}
```

## Attributes Reference

- `load_balancer_types` - (list) List of all load balancer types. See `data.hcloud_load_balancer_type` for the schema.
