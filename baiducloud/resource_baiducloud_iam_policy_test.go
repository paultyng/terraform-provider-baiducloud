package baiducloud

import (
	"fmt"
	"github.com/baidubce/bce-sdk-go/services/iam"
	"github.com/baidubce/bce-sdk-go/services/iam/api"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-baiducloud/baiducloud/connectivity"
	"log"
	"strings"
	"testing"
)

const (
	testAccIamPolicyResourceType = "baiducloud_iam_policy"
	testAccIamPolicyPrefix       = "test_BaiduAcc"
	testAccIamPolicyResourceName = testAccIamPolicyResourceType + "." + BaiduCloudTestResourceName
	testAccIamPolicyDescription  = "test_description"
)

func init() {
	resource.AddTestSweepers(testAccIamPolicyResourceType, &resource.Sweeper{
		Name: testAccIamPolicyResourceType,
		F:    testSweepIamPolicies,
	})
}

func testSweepIamPolicies(region string) error {
	rawClient, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("get BaiduCloud client error: %s", err)
	}

	client := rawClient.(*connectivity.BaiduClient)

	raw, err := client.WithIamClient(func(iamClient *iam.Client) (i interface{}, e error) {
		return iamClient.ListPolicy("", api.POLICY_TYPE_CUSTOM)
	})
	if err != nil {
		return fmt.Errorf("list policies error: %v", err)
	}

	result, _ := raw.(*api.ListPolicyResult)
	for _, policy := range result.Policies {
		if !strings.HasPrefix(policy.Name, testAccIamPolicyPrefix) {
			continue
		}
		log.Printf("[INFO] Deleting policy: %s", policy.Name)
		_, err := client.WithIamClient(func(iamClient *iam.Client) (i interface{}, e error) {
			return nil, iamClient.DeletePolicy(policy.Name)
		})
		if err != nil {
			return fmt.Errorf("delete policy error: %v", err)
		}
	}
	return nil
}
func TestAccBaiduCloudIamPolicy(t *testing.T) {
	name := strings.ReplaceAll(acctest.RandomWithPrefix(testAccIamPolicyPrefix), "-", "_")
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccIamPolicyDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccIamPolicyConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBaiduCloudDataSourceId(testAccIamPolicyResourceName),
					resource.TestCheckResourceAttr(testAccIamPolicyResourceName, "name", name),
					resource.TestCheckResourceAttr(testAccIamPolicyResourceName, "description", testAccIamPolicyDescription),
					resource.TestCheckResourceAttrSet(testAccIamPolicyResourceName, "unique_id"),
				),
			},
			{
				ResourceName:            testAccIamPolicyResourceName,
				ImportState:             true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccIamPolicyConfig(name string) string {
	return fmt.Sprintf(`
resource "%s" "%s" {
  name = "%s"
  description = "%s"
  document = <<EOF
  {"accessControlList": [{"region":"bj","service":"bcc","resource":["*"],"permission":["*"],"effect":"Allow"}]}
  EOF
}`, testAccIamPolicyResourceType, BaiduCloudTestResourceName, name, testAccIamPolicyDescription)
}

func testAccIamPolicyDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*connectivity.BaiduClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != testAccIamPolicyResourceType {
			continue
		}

		_, err := client.WithIamClient(func(iamClient *iam.Client) (i interface{}, e error) {
			return iamClient.GetPolicy(rs.Primary.ID, api.POLICY_TYPE_CUSTOM)
		})
		if err != nil {
			if NotFoundError(err) {
				continue
			}
			return WrapError(err)
		}

		return WrapError(Error("Iam Policy still exist"))
	}

	return nil
}
