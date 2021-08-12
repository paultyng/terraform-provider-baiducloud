package baiducloud

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/baidubce/bce-sdk-go/services/rds"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/terraform-providers/terraform-provider-baiducloud/baiducloud/connectivity"
)

const (
	testAccRdsInstanceResourceType = "baiducloud_rds_instance"
	testAccRdsInstanceResourceName = testAccRdsInstanceResourceType + "." + BaiduCloudTestResourceName
)

func init() {
	resource.AddTestSweepers(testAccRdsInstanceResourceType, &resource.Sweeper{
		Name: testAccRdsInstanceResourceType,
		F:    testSweepRdsInstances,
	})
}

func testSweepRdsInstances(region string) error {
	rawClient, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("get BaiduCloud client error: %s", err)
	}

	client := rawClient.(*connectivity.BaiduClient)
	rdsClient := &RdsService{client}

	args := &rds.ListRdsArgs{}
	instList, err := rdsClient.ListAllInstances(args)
	if err != nil {
		return fmt.Errorf("get RDS instances error: %v", err)
	}

	for _, inst := range instList {
		if !strings.HasPrefix(inst.InstanceName, BaiduCloudTestResourceTypeName) || inst.InstanceStatus != RDSStatusRunning {
			log.Printf("[INFO] Skipping RDS instance: %s (%s)", inst.InstanceId, inst.InstanceName)
			continue
		}

		log.Printf("[INFO] Deleting RDS instance: %s (%s)", inst.InstanceId, inst.InstanceName)
		_, err := client.WithRdsClient(func(rdsClient *rds.Client) (interface{}, error) {
			return nil, rdsClient.DeleteRds(inst.InstanceId)
		})

		if err != nil {
			log.Printf("[ERROR] Failed to delete RDS instance %s (%s)", inst.InstanceId, inst.InstanceName)
		}
	}

	return nil
}

func TestAccBaiduCloudRdsInstance(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccRdsInstanceDestory,

		Steps: []resource.TestStep{
			{
				Config: testAccRdsInstanceConfig(BaiduCloudTestResourceTypeNameRdsInstance),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBaiduCloudDataSourceId(testAccRdsInstanceResourceName),
					resource.TestCheckResourceAttr(testAccRdsInstanceResourceName, "billing.payment_timing", "Postpaid"),
				),
			},
			{
				ResourceName:            testAccRdsInstanceResourceName,
				ImportState:             true,
				ImportStateVerifyIgnore: []string{"instance_status"},
			},
			{
				Config: testAccRdsInstanceConfigUpdate(BaiduCloudTestResourceTypeNameRdsInstance),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBaiduCloudDataSourceId(testAccRdsInstanceResourceName),
					resource.TestCheckResourceAttr(testAccRdsInstanceResourceName, "billing.payment_timing", "Postpaid"),
				),
			},
		},
	})
}

func testAccRdsInstanceDestory(s *terraform.State) error {
	client := testAccProvider.Meta().(*connectivity.BaiduClient)
	rdsService := RdsService{client}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != testAccRdsInstanceResourceType {
			continue
		}

		result, err := rdsService.GetInstanceDetail(rs.Primary.ID)
		if err != nil {
			if NotFoundError(err) {
				continue
			}
			return WrapError(err)
		}

		deletingStatus := []string{"Deleted", "Deleting"}
		if IsExpectValue(result.InstanceStatus, deletingStatus) {
			continue
		}

		return WrapError(Error("RDS Instance still exist"))
	}

	return nil
}

func testAccRdsInstanceConfig(name string) string {
	return fmt.Sprintf(`
resource "baiducloud_rds_instance" "default" {
    instance_name             = "%s"
    billing = {
        payment_timing        = "Postpaid"
    }
    engine_version            = "5.6"
    engine                    = "MySQL"
    cpu_count                 = 1
    memory_capacity           = 1
    volume_capacity           = 5
	security_ips 			  = ["192.168.1.1"]
	parameters{
		name  	= "connect_timeout"
		value 	= "15"
	}
	parameters{
		name  	= "lower_case_table_names"
		value 	= "1"
	}
}
`, name+"-rds")
}

func testAccRdsInstanceConfigUpdate(name string) string {
	return fmt.Sprintf(`
resource "baiducloud_rds_instance" "default" {
    instance_name             = "%s"
    billing = {
        payment_timing        = "Postpaid"
    }
    engine_version            = "5.6"
    engine                    = "MySQL"
    cpu_count                 = 1
    memory_capacity           = 2
    volume_capacity           = 5
	security_ips 			  = ["192.168.1.1","192.168.3.1"]
	parameters{
		name  	= "connect_timeout"
		value 	= "100"
	}
	parameters{
		name  	= "lower_case_table_names"
		value 	= "1"
	}
}
`, name+"-rds")
}
