package ucloud

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudSecGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudSecGroupCreate,
		Read:   resourceUCloudSecGroupRead,
		Update: resourceUCloudSecGroupUpdate,
		Delete: resourceUCloudSecGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateName,
			},

			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"remark": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"rules": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"direction": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"Ingress",
								"Egress",
							}, false),
						},

						"protocol_type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"TCP",
								"UDP",
								"ICMP",
								"ICMPv6",
								"ALL",
							}, false),
						},

						"dst_port": {
							Type:     schema.TypeString,
							Required: true,
						},

						"ip_range": {
							Type:     schema.TypeString,
							Required: true,
						},

						"rule_action": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"Accept",
								"Drop",
							}, false),
						},

						"priority": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 200),
						},

						"remark": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},

						"rule_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUCloudSecGroupCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.vpcconn

	req := conn.NewCreateSecGroupRequest()
	req.Name = ucloud.String(d.Get("name").(string))
	req.VPCID = ucloud.String(d.Get("vpc_id").(string))

	resp, err := conn.CreateSecGroup(req)
	if err != nil {
		return fmt.Errorf("error on creating sec group, %s", err)
	}

	d.SetId(resp.SecGroupId)

	// create rules if specified
	if v, ok := d.GetOk("rules"); ok {
		rules := v.([]interface{})
		if len(rules) > 0 {
			if err := createSecGroupRules(conn, d.Id(), rules); err != nil {
				return err
			}
		}
	}

	// update remark if specified
	if v, ok := d.GetOk("remark"); ok {
		req := client.genericClient.NewGenericRequest()
		err := req.SetPayload(map[string]interface{}{
			"Action":     "UpdateSecGroup",
			"SecGroupId": []string{d.Id()},
			"Remark":     v.(string),
		})
		if err != nil {
			return fmt.Errorf("error on setting payload for UpdateSecGroup %q, %s", d.Id(), err)
		}
		_, err = client.genericClient.GenericInvoke(req)
		if err != nil {
			return fmt.Errorf("error on setting remark for sec group %q, %s", d.Id(), err)
		}
	}

	return resourceUCloudSecGroupRead(d, meta)
}

func resourceUCloudSecGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.vpcconn

	d.Partial(true)

	if d.HasChange("name") || d.HasChange("remark") {
		payload := map[string]interface{}{
			"Action":     "UpdateSecGroup",
			"SecGroupId": []string{d.Id()},
		}

		if d.HasChange("name") {
			payload["Name"] = d.Get("name").(string)
		}

		if d.HasChange("remark") {
			payload["Remark"] = d.Get("remark").(string)
		}

		req := client.genericClient.NewGenericRequest()
		err := req.SetPayload(payload)
		if err != nil {
			return fmt.Errorf("error on setting payload for UpdateSecGroup %q, %s", d.Id(), err)
		}
		_, err = client.genericClient.GenericInvoke(req)
		if err != nil {
			return fmt.Errorf("error on %s to sec group %q, %s", "UpdateSecGroup", d.Id(), err)
		}

		d.SetPartial("name")
		d.SetPartial("remark")
	}

	if d.HasChange("rules") {
		o, n := d.GetChange("rules")
		oldRules := o.([]interface{})
		newRules := n.([]interface{})

		// delete old rules
		if len(oldRules) > 0 {
			var ruleIds []string
			for _, r := range oldRules {
				rule := r.(map[string]interface{})
				if ruleId, ok := rule["rule_id"]; ok && ruleId.(string) != "" {
					ruleIds = append(ruleIds, ruleId.(string))
				}
			}
			if len(ruleIds) > 0 {
				reqDel := conn.NewDeleteSecGroupRuleRequest()
				reqDel.SecGroupId = ucloud.String(d.Id())
				reqDel.RuleId = ruleIds
				_, err := conn.DeleteSecGroupRule(reqDel)
				if err != nil {
					return fmt.Errorf("error on deleting sec group rules for %q, %s", d.Id(), err)
				}
			}
		}

		// create new rules
		if len(newRules) > 0 {
			if err := createSecGroupRules(conn, d.Id(), newRules); err != nil {
				return err
			}
		}

		d.SetPartial("rules")
	}

	d.Partial(false)

	return resourceUCloudSecGroupRead(d, meta)
}

func resourceUCloudSecGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	sgSet, err := client.describeSecGroupById(d.Id())

	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading sec group %q, %s", d.Id(), err)
	}

	d.Set("name", sgSet.Name)
	d.Set("vpc_id", sgSet.VPCId)
	d.Set("remark", sgSet.Remark)
	d.Set("create_time", timestampToString(sgSet.CreateTime))

	rules := []map[string]interface{}{}
	for _, item := range sgSet.Rule {
		rules = append(rules, map[string]interface{}{
			"direction":     item.Direction,
			"protocol_type": item.ProtocolType,
			"dst_port":      item.DstPort,
			"ip_range":      item.IPRange,
			"rule_action":   item.RuleAction,
			"priority":      item.Priority,
			"remark":        item.Remark,
			"rule_id":       item.RuleId,
		})
	}

	if err := d.Set("rules", rules); err != nil {
		return err
	}

	return nil
}

func resourceUCloudSecGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.vpcconn

	req := conn.NewDeleteSecGroupRequest()
	req.SecGroupId = []string{d.Id()}

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteSecGroup(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting sec group %q, %s", d.Id(), err))
		}

		_, err := client.describeSecGroupById(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading sec group when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified sec group %q has not been deleted due to unknown error", d.Id()))
	})
}

func createSecGroupRules(conn *vpc.VPCClient, secGroupId string, rules []interface{}) error {
	req := conn.NewCreateSecGroupRuleRequest()
	req.SecGroupId = ucloud.String(secGroupId)

	for _, r := range rules {
		rule := r.(map[string]interface{})
		req.Rule = append(req.Rule, vpc.CreateSecGroupRuleParamRule{
			Direction:    ucloud.String(rule["direction"].(string)),
			ProtocolType: ucloud.String(rule["protocol_type"].(string)),
			DstPort:      ucloud.String(rule["dst_port"].(string)),
			IPRange:      ucloud.String(rule["ip_range"].(string)),
			RuleAction:   ucloud.String(rule["rule_action"].(string)),
			Priority:     ucloud.Int(rule["priority"].(int)),
			Remark:       ucloud.String(rule["remark"].(string)),
		})
	}

	_, err := conn.CreateSecGroupRule(req)
	if err != nil {
		return fmt.Errorf("error on creating sec group rules for %q, %s", secGroupId, err)
	}

	return nil
}
