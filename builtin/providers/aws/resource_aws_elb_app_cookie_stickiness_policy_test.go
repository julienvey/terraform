package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/aws-sdk-go/aws"
	"github.com/hashicorp/aws-sdk-go/gen/elb"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSElbAppCookieStickinessPolicy(t *testing.T) {
	var v elb.AppCookieStickinessPolicy

	testCheck := func(*terraform.State) error {
		if *v.CookieName != "fookie" {
			return fmt.Errorf("bad cookie name: %s", *v.CookieName)
		}

		if *v.PolicyName != "foobar-terraform-test" {
			return fmt.Errorf("bad MapPublicIpOnLaunch: %s", *v.PolicyName)
		}

		return nil
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckElbAppCookieStickinessPolicyDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccElbAppCookieStickinessPolicyConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckElbAppCookieStickinessPolicyExists(
						"aws_elb_app_cookie_stickiness.bar", &v),
					testCheck,
				),
			},
		},
	})
}

func testAccCheckElbAppCookieStickinessPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elbconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elb_app_cookie_stickiness" {
			continue
		}

		// Try to find the resource
		resp, err := conn.DescribeLoadBalancerPolicies(&elb.DescribeLoadBalancerPoliciesInput{
			PolicyNames:      []string{rs.Primary.ID},
			LoadBalancerName: aws.String("foobar-terraform-test-lb"),
		})
		if err == nil {
			if len(resp.PolicyDescriptions) > 0 {
				return fmt.Errorf("still exists")
			}

			return nil
		}

		// Verify the error is what we want
		ec2err, ok := err.(aws.APIError)
		if !ok {
			return err
		}
		if ec2err.Code != "PolicyNotFound.NotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckElbAppCookieStickinessPolicyExists(n string, v *elb.AppCookieStickinessPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).elbconn
		resp, err := conn.DescribeLoadBalancerPolicies(&elb.DescribeLoadBalancerPoliciesInput{
			PolicyNames:      []string{rs.Primary.ID},
			LoadBalancerName: aws.String("foobar-terraform-test-lb"),
		})
		if err != nil {
			return err
		}
		if len(resp.PolicyDescriptions) == 0 {
			return fmt.Errorf("ElbAppCookieStickinessPolicy not found")
		}
		desc := resp.PolicyDescriptions[0]
		var cookieName string
		for _, attr := range desc.PolicyAttributeDescriptions {
			if *attr.AttributeName == "CookieName" {
				cookieName = *attr.AttributeValue
				break
			}
		}

		*v = elb.AppCookieStickinessPolicy{
			PolicyName: aws.String(rs.Primary.ID),
			CookieName: aws.String(cookieName),
		}

		return nil
	}
}

const testAccElbAppCookieStickinessPolicyConfig = `
resource "aws_elb" "bar" {
  name = "foobar-terraform-test-lb"
	availability_zones = ["us-west-2a"]

  listener {
    instance_port = 8000
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
  }
}

resource "aws_elb_app_cookie_stickiness" "bar" {
  name = "foobar-terraform-test"
	load_balancer_name = "${aws_elb.bar.name}"
	cookie_name = "fookie"
}
`
