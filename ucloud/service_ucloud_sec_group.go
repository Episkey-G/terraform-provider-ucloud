package ucloud

import (
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func (c *UCloudClient) describeSecGroupsByVPCId(vpcId string) ([]vpc.SecGroupInfo, error) {
	conn := c.vpcconn

	var allSecGroups []vpc.SecGroupInfo
	var offset int
	var limit = 100

	for {
		req := conn.NewDescribeSecGroupRequest()
		if vpcId != "" {
			req.VPCId = ucloud.String(vpcId)
		}
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)

		resp, err := conn.DescribeSecGroup(req)
		if err != nil {
			return nil, err
		}

		allSecGroups = append(allSecGroups, resp.DataSet...)

		if len(resp.DataSet) < limit {
			break
		}

		offset += limit
	}

	return allSecGroups, nil
}

func (c *UCloudClient) describeResourceSecGroup(resourceId string) ([]vpc.BindingSecGroupInfo, error) {
	conn := c.vpcconn

	req := conn.NewDescribeResourceSecGroupRequest()
	req.ResourceId = []string{resourceId}

	resp, err := conn.DescribeResourceSecGroup(req)
	if err != nil {
		return nil, err
	}

	if len(resp.DataSet) < 1 {
		return nil, nil
	}

	return resp.DataSet[0].SecGroupInfo, nil
}
