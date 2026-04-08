package ucloud

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
)

func dataSourceUCloudSecGroups() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudSecGroupsRead,

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set:      schema.HashString,
				Computed: true,
			},

			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
			},

			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"total_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"sec_groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"rules": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"direction": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"protocol_type": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"dst_port": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"ip_range": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"rule_action": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"priority": {
										Type:     schema.TypeInt,
										Computed: true,
									},

									"remark": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"rule_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},

						"tag": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"remark": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"create_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceUCloudSecGroupsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	var vpcId string
	if v, ok := d.GetOk("vpc_id"); ok {
		vpcId = v.(string)
	}

	allSecGroups, err := client.describeSecGroupsByVPCId(vpcId)
	if err != nil {
		return fmt.Errorf("error on reading sec group list, %s", err)
	}

	var secGroups []vpc.SecGroupInfo

	ids, idsOk := d.GetOk("ids")
	nameRegex, nameRegexOk := d.GetOk("name_regex")
	if idsOk || nameRegexOk {
		var r *regexp.Regexp
		if nameRegex != "" {
			r = regexp.MustCompile(nameRegex.(string))
		}
		for _, v := range allSecGroups {
			if r != nil && !r.MatchString(v.Name) {
				continue
			}

			if idsOk && !isStringIn(v.SecGroupId, schemaSetToStringSlice(ids)) {
				continue
			}
			secGroups = append(secGroups, v)
		}
	} else {
		secGroups = allSecGroups
	}

	err = dataSourceUCloudSecGroupsSave(d, secGroups)
	if err != nil {
		return fmt.Errorf("error on reading sec group list, %s", err)
	}

	return nil
}

func dataSourceUCloudSecGroupsSave(d *schema.ResourceData, secGroups []vpc.SecGroupInfo) error {
	ids := []string{}
	data := []map[string]interface{}{}

	for _, item := range secGroups {
		ids = append(ids, item.SecGroupId)

		rules := []map[string]interface{}{}
		for _, v := range item.Rule {
			rules = append(rules, map[string]interface{}{
				"direction":     v.Direction,
				"protocol_type": v.ProtocolType,
				"dst_port":      v.DstPort,
				"ip_range":      v.IPRange,
				"rule_action":   v.RuleAction,
				"priority":      v.Priority,
				"remark":        v.Remark,
				"rule_id":       v.RuleId,
			})
		}

		data = append(data, map[string]interface{}{
			"id":          item.SecGroupId,
			"name":        item.Name,
			"vpc_id":      item.VPCId,
			"type":        item.Type,
			"tag":         item.Tag,
			"remark":      item.Remark,
			"rules":       rules,
			"create_time": timestampToString(item.CreateTime),
		})
	}

	d.SetId(hashStringArray(ids))
	d.Set("total_count", len(data))
	d.Set("ids", ids)
	if err := d.Set("sec_groups", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}

	return nil
}
