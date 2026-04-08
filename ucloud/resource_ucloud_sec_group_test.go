package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
)

func TestAccUCloudSecGroup_basic(t *testing.T) {
	rInt := acctest.RandInt()
	var sgSet vpc.SecGroupInfo

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_sec_group.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckSecGroupDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccSecGroupConfig(rInt),

				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecGroupExists("ucloud_sec_group.foo", &sgSet),
					testAccCheckSecGroupAttributes(&sgSet),
					resource.TestCheckResourceAttr("ucloud_sec_group.foo", "name", fmt.Sprintf("tf-acc-sec-group-%d", rInt)),
				),
			},

			{
				Config: testAccSecGroupConfigUpdate(rInt),

				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecGroupExists("ucloud_sec_group.foo", &sgSet),
					testAccCheckSecGroupAttributes(&sgSet),
					resource.TestCheckResourceAttr("ucloud_sec_group.foo", "name", fmt.Sprintf("tf-acc-sec-group-%d-update", rInt)),
				),
			},
		},
	})
}

func TestAccUCloudSecGroup_instance(t *testing.T) {
	rInt := acctest.RandInt()
	var sgSet vpc.SecGroupInfo

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSecGroupDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccSecGroupInstanceConfig(rInt),

				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecGroupExists("ucloud_sec_group.foo", &sgSet),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "security_mode", "SecGroup"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "name", fmt.Sprintf("tf-acc-sec-group-instance-%d", rInt)),
				),
			},
		},
	})
}

func testAccCheckSecGroupExists(n string, sgSet *vpc.SecGroupInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("sec group id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeSecGroupById(rs.Primary.ID)

		log.Printf("[INFO] sec group id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*sgSet = *ptr
		return nil
	}
}

func testAccCheckSecGroupAttributes(sgSet *vpc.SecGroupInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if sgSet.SecGroupId == "" {
			return fmt.Errorf("sec group id is empty")
		}
		return nil
	}
}

func testAccCheckSecGroupDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_sec_group" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeSecGroupById(rs.Primary.ID)

		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.SecGroupId != "" {
			return fmt.Errorf("sec group still exist")
		}
	}

	return nil
}

func testAccSecGroupConfig(rInt int) string {
	return fmt.Sprintf(`
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-sec-group-vpc-%d"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_sec_group" "foo" {
	name   = "tf-acc-sec-group-%d"
	vpc_id = ucloud_vpc.foo.id

	rules {
		direction     = "Ingress"
		protocol_type = "TCP"
		dst_port      = "80"
		ip_range      = "0.0.0.0/0"
		rule_action   = "Accept"
		priority      = 100
		remark        = "allow http"
	}
}`, rInt, rInt)
}

func testAccSecGroupConfigUpdate(rInt int) string {
	return fmt.Sprintf(`
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-sec-group-vpc-%d"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_sec_group" "foo" {
	name   = "tf-acc-sec-group-%d-update"
	vpc_id = ucloud_vpc.foo.id

	rules {
		direction     = "Ingress"
		protocol_type = "TCP"
		dst_port      = "80,443"
		ip_range      = "0.0.0.0/0"
		rule_action   = "Accept"
		priority      = 100
		remark        = "allow web"
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
}`, rInt, rInt)
}

func testAccSecGroupInstanceConfig(rInt int) string {
	return fmt.Sprintf(`
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-sec-group-vpc-%d"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "foo" {
	name      = "tf-acc-sec-group-subnet-%d"
	tag       = "tf-acc"
	cidr_block = "192.168.1.0/24"
	vpc_id    = ucloud_vpc.foo.id
}

data "ucloud_images" "default" {
	availability_zone = "cn-bj2-05"
	name_regex        = "^CentOS 7.[1-9] 64"
	image_type        = "base"
}

resource "ucloud_sec_group" "foo" {
	name   = "tf-acc-sec-group-%d"
	vpc_id = ucloud_vpc.foo.id

	rules {
		direction     = "Ingress"
		protocol_type = "TCP"
		dst_port      = "22"
		ip_range      = "0.0.0.0/0"
		rule_action   = "Accept"
		priority      = 100
		remark        = "allow ssh"
	}
}

resource "ucloud_instance" "foo" {
	availability_zone = "cn-bj2-05"
	image_id          = data.ucloud_images.default.images[0].id
	instance_type     = "o-standard-2"
	root_password     = "wA1234567"
	charge_type       = "dynamic"
	name              = "tf-acc-sec-group-instance-%d"
	boot_disk_type    = "cloud_rssd"
	min_cpu_platform  = "Amd/Auto"
	vpc_id            = ucloud_vpc.foo.id
	subnet_id         = ucloud_subnet.foo.id

	security_mode = "SecGroup"
	sec_group_id  = [ucloud_sec_group.foo.id]
}`, rInt, rInt, rInt, rInt)
}
