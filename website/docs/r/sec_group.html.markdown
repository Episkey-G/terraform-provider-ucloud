---
subcategory: "VPC"
layout: "ucloud"
page_title: "UCloud: ucloud_sec_group"
description: |-
  Provides a Security Group (SecGroup) resource.
---

# ucloud_sec_group

Provides a Security Group resource under VPC, which is different from the firewall (`ucloud_security_group`).

~> **Note** Security Group (SecGroup) is a VPC-level security product. It is different from the firewall (`ucloud_security_group`). When creating an instance with `security_mode = "SecGroup"`, you must use `ucloud_sec_group` instead of `ucloud_security_group`.

## Example Usage

```hcl
# Query default VPC
data "ucloud_vpcs" "default" {
}

# Create a security group
resource "ucloud_sec_group" "example" {
  name   = "tf-example-sec-group"
  vpc_id = data.ucloud_vpcs.default.vpcs[0].id

  rules {
    direction     = "Ingress"
    protocol_type = "TCP"
    dst_port      = "80,443"
    ip_range      = "0.0.0.0/0"
    rule_action   = "Accept"
    priority      = 100
    remark        = "allow web traffic"
  }

  rules {
    direction     = "Ingress"
    protocol_type = "TCP"
    dst_port      = "22"
    ip_range      = "0.0.0.0/0"
    rule_action   = "Accept"
    priority      = 101
    remark        = "allow ssh"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the security group, which contains 1-63 characters.
* `vpc_id` - (Required, ForceNew) The ID of VPC linked to the security group.

- - -

* `remark` - (Optional) The remarks of the security group.
* `rules` - (Optional) A list of security group rules. Each element contains the following attributes:
    * `direction` - (Required) The direction of the rule. Possible values are: `Ingress` and `Egress`.
    * `protocol_type` - (Required) The protocol type. Possible values are: `TCP`, `UDP`, `ICMP`, `ICMPv6` and `ALL`.
    * `dst_port` - (Required) The destination port. Multiple ports can be separated by commas, e.g. `80,443`. Port ranges are supported, e.g. `2000-10000`.
    * `ip_range` - (Required) The IP address range in CIDR notation, e.g. `0.0.0.0/0`.
    * `rule_action` - (Required) The action of the rule. Possible values are: `Accept` and `Drop`.
    * `priority` - (Required) The priority of the rule. Range: 1-200.
    * `remark` - (Optional) The remarks of the rule.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the security group.
* `create_time` - The time of creation, formatted in RFC3339 time string.
* `rules` - In addition to the arguments above, each rule also exports:
    * `rule_id` - The ID of the rule.

## Import

Security Group can be imported using the `id`, e.g.

```
$ terraform import ucloud_sec_group.example secgroup-abcdefg
```
