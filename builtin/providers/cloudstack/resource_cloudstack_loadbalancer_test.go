package cloudstack

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/xanzy/go-cloudstack/cloudstack"
)

func TestAccCloudStackLoadBalancerRule_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackLoadBalancerRuleDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCloudStackLoadBalancerRule_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackLoadBalancerRuleExist("cloudstack_loadbalancer_rule.foo"),
					resource.TestCheckResourceAttr(
						"cloudstack_loadbalancer_rule.foo", "name", "terraform-lb"),
					resource.TestCheckResourceAttr(
						"cloudstack_loadbalancer_rule.foo", "algorithm", "roundrobin"),
					resource.TestCheckResourceAttr(
						"cloudstack_loadbalancer_rule.foo", "public_port", "80"),
					resource.TestCheckResourceAttr(
						"cloudstack_loadbalancer_rule.foo", "private_port", "80"),
				),
			},
		},
	})
}

func TestAccCloudStackLoadBalancerRule_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackLoadBalancerRuleDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCloudStackLoadBalancerRule_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackLoadBalancerRuleExist("cloudstack_loadbalancer_rule.foo"),
					resource.TestCheckResourceAttr(
						"cloudstack_loadbalancer_rule.foo", "name", "terraform-lb"),
					resource.TestCheckResourceAttr(
						"cloudstack_loadbalancer_rule.foo", "algorithm", "roundrobin"),
					resource.TestCheckResourceAttr(
						"cloudstack_loadbalancer_rule.foo", "public_port", "80"),
					resource.TestCheckResourceAttr(
						"cloudstack_loadbalancer_rule.foo", "private_port", "80"),
				),
			},

			resource.TestStep{
				Config: testAccCloudStackLoadBalancerRule_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackLoadBalancerRuleExist("cloudstack_loadbalancer_rule.foo"),
					resource.TestCheckResourceAttr(
						"cloudstack_loadbalancer_rule.foo", "name", "terraform-lb-update"),
					resource.TestCheckResourceAttr(
						"cloudstack_loadbalancer_rule.foo", "algorithm", "leastconn"),
					resource.TestCheckResourceAttr(
						"cloudstack_loadbalancer_rule.foo", "public_port", "443"),
					resource.TestCheckResourceAttr(
						"cloudstack_loadbalancer_rule.foo", "private_port", "443"),
				),
			},
		},
	})
}

func testAccCheckCloudStackLoadBalancerRuleExist(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No loadbalancer rule ID is set")
		}

		for k, uuid := range rs.Primary.Attributes {
			if !strings.Contains(k, "uuid") {
				continue
			}

			cs := testAccProvider.Meta().(*cloudstack.CloudStackClient)
			_, count, err := cs.LoadBalancer.GetLoadBalancerRuleByID(uuid)

			if err != nil {
				return err
			}

			if count == 0 {
				return fmt.Errorf("Loadbalancer rule for %s not found", k)
			}
		}

		return nil
	}
}

func testAccCheckCloudStackLoadBalancerRuleDestroy(s *terraform.State) error {
	cs := testAccProvider.Meta().(*cloudstack.CloudStackClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudstack_loadbalancer_rule" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Loadbalancer rule ID is set")
		}

		for k, uuid := range rs.Primary.Attributes {
			if !strings.Contains(k, "uuid") {
				continue
			}

			_, _, err := cs.LoadBalancer.GetLoadBalancerRuleByID(uuid)
			if err == nil {
				return fmt.Errorf("Loadbalancer rule %s still exists", rs.Primary.ID)
			}
		}
	}

	return nil
}

var testAccCloudStackLoadBalancerRule_basic = fmt.Sprintf(`
resource "cloudstack_vpc" "foobar" {
	name = "terraform-vpc"
	cidr = "%s"
	vpc_offering = "%s"
	zone = "%s"
}

resource "cloudstack_network" "foo" {
  name = "terraform-network"
  cidr = "%s"
  network_offering = "%s"
  vpc = "${cloudstack_vpc.foobar.name}"
  zone = "${cloudstack_vpc.foobar.zone}"
}

resource "cloudstack_ipaddress" "foo" {
  vpc = "${cloudstack_vpc.foobar.name}"
}

resource "cloudstack_instance" "foobar1" {
  name = "terraform-server1"
  display_name = "terraform"
  service_offering= "%s"
  network = "${cloudstack_network.foo.name}"
  template = "%s"
  zone = "${cloudstack_network.foo.zone}"
  user_data = "foobar\nfoo\nbar"
  expunge = true
}

resource "cloudstack_loadbalancer_rule" "foo" {
  name = "terraform-lb"
  ipaddress = "${cloudstack_ipaddress.foo.ipaddress}"
  algorithm = "roundrobin"
  network = "${cloudstack_network.foo.id}"
  public_port = 80
  private_port = 80
  members = ["${cloudstack_instance.foobar1.id}"]
}`,
	CLOUDSTACK_VPC_CIDR_1,
	CLOUDSTACK_VPC_OFFERING,
	CLOUDSTACK_ZONE,
	CLOUDSTACK_VPC_NETWORK_CIDR,
	CLOUDSTACK_VPC_NETWORK_OFFERING,
	CLOUDSTACK_SERVICE_OFFERING_1,
	CLOUDSTACK_TEMPLATE)

var testAccCloudStackLoadBalancerRule_update = fmt.Sprintf(`
resource "cloudstack_vpc" "foobar" {
    name = "terraform-vpc"
    cidr = "%s"
    vpc_offering = "%s"
    zone = "%s"
}

resource "cloudstack_network" "foo" {
  name = "terraform-network"
  cidr = "%s"
  network_offering = "%s"
  vpc = "${cloudstack_vpc.foobar.name}"
  zone = "${cloudstack_vpc.foobar.zone}"
}

resource "cloudstack_ipaddress" "foo" {
  vpc = "${cloudstack_vpc.foobar.name}"
}

resource "cloudstack_instance" "foobar1" {
  name = "terraform-server1"
  display_name = "terraform"
  service_offering= "%s"
  network = "${cloudstack_network.foo.name}"
  template = "%s"
  zone = "${cloudstack_network.foo.zone}"
  user_data = "foobar\nfoo\nbar"
  expunge = true
}

resource "cloudstack_instance" "foobar2" {
  name = "terraform-server2"
  display_name = "terraform"
  service_offering= "%s"
  network = "${cloudstack_network.foo.name}"
  template = "%s"
  zone = "${cloudstack_network.foo.zone}"
  user_data = "foobar\nfoo\nbar"
  expunge = true
}

resource "cloudstack_loadbalancer_rule" "foo" {
  name = "terraform-lb-update"
  ipaddress = "${cloudstack_ipaddress.foo.ipaddress}"
  algorithm = "leastconn"
  network = "${cloudstack_network.foo.id}"
  public_port = 443
  private_port = 443
  members = ["${cloudstack_instance.foobar2.id}", "${cloudstack_instance.foobar1.id}"]
}`,
	CLOUDSTACK_VPC_CIDR_1,
	CLOUDSTACK_VPC_OFFERING,
	CLOUDSTACK_ZONE,
	CLOUDSTACK_VPC_NETWORK_CIDR,
	CLOUDSTACK_VPC_NETWORK_OFFERING,
	CLOUDSTACK_SERVICE_OFFERING_1,
	CLOUDSTACK_TEMPLATE,
	CLOUDSTACK_SERVICE_OFFERING_1,
	CLOUDSTACK_TEMPLATE)
