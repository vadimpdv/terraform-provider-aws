package iam_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/iam"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccIAMUserPolicyAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var out iam.ListAttachedUserPoliciesOutput
	rName := sdkacctest.RandString(10)
	policyName1 := fmt.Sprintf("test-policy-%s", sdkacctest.RandString(10))
	policyName2 := fmt.Sprintf("test-policy-%s", sdkacctest.RandString(10))
	policyName3 := fmt.Sprintf("test-policy-%s", sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPolicyAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPolicyAttachmentConfig_attach(rName, policyName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPolicyAttachmentExists(ctx, "aws_iam_user_policy_attachment.test-attach", 1, &out),
					testAccCheckUserPolicyAttachmentAttributes([]string{policyName1}, &out),
				),
			},
			{
				ResourceName:      "aws_iam_user_policy_attachment.test-attach",
				ImportState:       true,
				ImportStateIdFunc: testAccUserPolicyAttachmentImportStateIdFunc("aws_iam_user_policy_attachment.test-attach"),
				// We do not have a way to align IDs since the Create function uses id.PrefixedUniqueId()
				// Failed state verification, resource with ID USER-POLICYARN not found
				// ImportStateVerify: true,
				ImportStateCheck: func(s []*terraform.InstanceState) error {
					if len(s) != 1 {
						return fmt.Errorf("expected 1 state: %#v", s)
					}

					rs := s[0]

					if !arn.IsARN(rs.Attributes["policy_arn"]) {
						return fmt.Errorf("expected policy_arn attribute to be set and begin with arn:, received: %s", rs.Attributes["policy_arn"])
					}

					return nil
				},
			},
			{
				Config: testAccUserPolicyAttachmentConfig_attachUpdate(rName, policyName1, policyName2, policyName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPolicyAttachmentExists(ctx, "aws_iam_user_policy_attachment.test-attach", 2, &out),
					testAccCheckUserPolicyAttachmentAttributes([]string{policyName2, policyName3}, &out),
				),
			},
		},
	})
}

func testAccCheckUserPolicyAttachmentDestroy(s *terraform.State) error {
	return nil
}

func testAccCheckUserPolicyAttachmentExists(ctx context.Context, n string, c int, out *iam.ListAttachedUserPoliciesOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No policy name is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn()
		user := rs.Primary.Attributes["user"]

		attachedPolicies, err := conn.ListAttachedUserPoliciesWithContext(ctx, &iam.ListAttachedUserPoliciesInput{
			UserName: aws.String(user),
		})
		if err != nil {
			return fmt.Errorf("Error: Failed to get attached policies for user %s (%s)", user, n)
		}
		if c != len(attachedPolicies.AttachedPolicies) {
			return fmt.Errorf("Error: User (%s) has wrong number of policies attached on initial creation", n)
		}

		*out = *attachedPolicies
		return nil
	}
}

func testAccCheckUserPolicyAttachmentAttributes(policies []string, out *iam.ListAttachedUserPoliciesOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		matched := 0

		for _, p := range policies {
			for _, ap := range out.AttachedPolicies {
				// *ap.PolicyArn like arn:aws:iam::111111111111:policy/test-policy
				parts := strings.Split(*ap.PolicyArn, "/")
				if len(parts) == 2 && p == parts[1] {
					matched++
				}
			}
		}
		if matched != len(policies) || matched != len(out.AttachedPolicies) {
			return fmt.Errorf("Error: Number of attached policies was incorrect: expected %d matched policies, matched %d of %d", len(policies), matched, len(out.AttachedPolicies))
		}
		return nil
	}
}

func testAccUserPolicyAttachmentImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["user"], rs.Primary.Attributes["policy_arn"]), nil
	}
}

func testAccUserPolicyAttachmentConfig_attach(rName, policyName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "user" {
  name = "test-user-%s"
}

resource "aws_iam_policy" "policy" {
  name        = "%s"
  description = "A test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "iam:ChangePassword"
      ],
      "Resource": "*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_user_policy_attachment" "test-attach" {
  user       = aws_iam_user.user.name
  policy_arn = aws_iam_policy.policy.arn
}
`, rName, policyName)
}

func testAccUserPolicyAttachmentConfig_attachUpdate(rName, policyName1, policyName2, policyName3 string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "user" {
  name = "test-user-%s"
}

resource "aws_iam_policy" "policy" {
  name        = "%s"
  description = "A test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "iam:ChangePassword"
      ],
      "Resource": "*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "policy2" {
  name        = "%s"
  description = "A test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "iam:ChangePassword"
      ],
      "Resource": "*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "policy3" {
  name        = "%s"
  description = "A test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "iam:ChangePassword"
      ],
      "Resource": "*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_user_policy_attachment" "test-attach" {
  user       = aws_iam_user.user.name
  policy_arn = aws_iam_policy.policy2.arn
}

resource "aws_iam_user_policy_attachment" "test-attach2" {
  user       = aws_iam_user.user.name
  policy_arn = aws_iam_policy.policy3.arn
}
`, rName, policyName1, policyName2, policyName3)
}
