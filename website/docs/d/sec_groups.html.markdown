---
subcategory: "VPC"
layout: "ucloud"
page_title: "UCloud: ucloud_sec_groups"
description: |-
  Provides a list of Security Group (SecGroup) resources.
---

# ucloud_sec_groups

This data source provides a list of Security Group resources.

## Example Usage

```hcl
data "ucloud_sec_groups" "example" {
  vpc_id = "uvnet-xxxxx"
}

output "first" {
  value = data.ucloud_sec_groups.example.sec_groups[0].id
}
```

## Argument Reference

The following arguments are supported:

* `vpc_id` - (Optional) The ID of VPC to filter security groups.
* `ids` - (Optional) A list of security group IDs to filter.
* `name_regex` - (Optional) A regex string to filter results by security group name.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `total_count` - Total number of security groups that satisfy the condition.
* `ids` - A list of security group IDs.
* `sec_groups` - It is a nested type which documented below.

- - -

The attribute (`sec_groups`) supports the following:

* `id` - The ID of the security group.
* `name` - The name of the security group.
* `vpc_id` - The ID of VPC linked to the security group.
* `type` - The type of the security group.
* `tag` - A tag assigned to the security group.
* `remark` - The remarks of the security group.
* `create_time` - The time of creation, formatted in RFC3339 time string.
* `rules` - A list of security group rules. Each element contains the following attributes:
    * `rule_id` - The ID of the rule.
    * `direction` - The direction of the rule.
    * `protocol_type` - The protocol type.
    * `dst_port` - The destination port.
    * `ip_range` - The IP address range.
    * `rule_action` - The action of the rule.
    * `priority` - The priority of the rule.
    * `remark` - The remarks of the rule.
