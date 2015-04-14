package aws

import (
	"fmt"
	"log"

	"github.com/hashicorp/aws-sdk-go/aws"
	"github.com/hashicorp/aws-sdk-go/gen/elb"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsElbAppCookieStickinessPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsElbAppCookieStickinessPolicyCreate,
		Read:   resourceAwsElbAppCookieStickinessPolicyRead,
		Update: resourceAwsElbAppCookieStickinessPolicyUpdate,
		Delete: resourceAwsElbAppCookieStickinessPolicyDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"load_balancer_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"cookie_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsElbAppCookieStickinessPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbconn

	elbOpts := &elb.CreateAppCookieStickinessPolicyInput{
		LoadBalancerName: aws.String(d.Get("load_balancer_name").(string)),
		CookieName:       aws.String(d.Get("cookie_name").(string)),
		PolicyName:       aws.String(d.Get("name").(string)),
	}

	log.Printf("[DEBUG] ELB App Cookie Policy create configuration: %#v", elbOpts)
	if _, err := elbconn.CreateAppCookieStickinessPolicy(elbOpts); err != nil {
		return fmt.Errorf("Error creating ELB App Cookie Policy: %s", err)
	}

	// Assign the elb's unique identifier for use later
	d.SetId(d.Get("name").(string))
	log.Printf("[INFO] ELB App Cookie Policy ID: %s", d.Id())

	return resourceAwsElbAppCookieStickinessPolicyRead(d, meta)
}

func resourceAwsElbAppCookieStickinessPolicyRead(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbconn

	// Retrieve the ELB Policy properties for updating the state
	describeElbOpts := &elb.DescribeLoadBalancerPoliciesInput{
		PolicyNames:      []string{d.Id()},
		LoadBalancerName: aws.String(d.Get("load_balancer_name").(string)),
	}

	describeResp, err := elbconn.DescribeLoadBalancerPolicies(describeElbOpts)
	if err != nil {
		if ec2err, ok := err.(aws.APIError); ok && ec2err.Code == "PolicyNotFound" {
			// The ELB Policy is gone now, so just remove it from the state
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error retrieving ELB App Cookie Stickiness Policy: %s", err)
	}
	if len(describeResp.PolicyDescriptions) != 1 {
		return fmt.Errorf("Unable to find ELB App Cookie Stickiness Policy: %#v", describeResp.PolicyDescriptions)
	}

	desc := describeResp.PolicyDescriptions[0]

	d.Set("name", *desc.PolicyName)

	for _, attr := range desc.PolicyAttributeDescriptions {
		if *attr.AttributeName == "CookieName" {
			d.Set("cookie_name", *attr.AttributeValue)
			break
		}
	}
	// We don't set load_balancer_name since it's a required argument to get the policy

	return nil
}

func resourceAwsElbAppCookieStickinessPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	return fmt.Errorf("Updating an AwsElbAppCookieStickinessPolicy is not possible in AWS")
}

func resourceAwsElbAppCookieStickinessPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbconn

	log.Printf("[INFO] Deleting ELB: %s", d.Id())

	// Destroy the load balancer
	deleteElbOpts := elb.DeleteLoadBalancerPolicyInput{
		LoadBalancerName: aws.String(d.Get("load_balancer_name").(string)),
		PolicyName:       aws.String(d.Id()),
	}
	if _, err := elbconn.DeleteLoadBalancerPolicy(&deleteElbOpts); err != nil {
		return fmt.Errorf("Error deleting ELB Policy: %s", err)
	}

	return nil
}
