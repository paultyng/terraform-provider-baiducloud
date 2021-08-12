/*
Use this data source to query BCC Instance list.

Example Usage

```hcl
data "baiducloud_instances" "default" {}

output "instances" {
 value = "${data.baiducloud_instances.default.instances}"
}
```
*/
package baiducloud

import (
	"github.com/baidubce/bce-sdk-go/services/bcc/api"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"

	"github.com/terraform-providers/terraform-provider-baiducloud/baiducloud/connectivity"
)

func dataSourceBaiduCloudInstances() *schema.Resource {
	diskSchema := map[string]*schema.Schema{
		"cds_id": {
			Type:        schema.TypeString,
			Description: "The id of the ephemeral disk.",
			Computed:    true,
		},
		"name": {
			Type:        schema.TypeString,
			Description: "The name of the ephemeral disk.",
			Computed:    true,
		},
		"description": {
			Type:        schema.TypeString,
			Description: "CDS volume description",
			Optional:    true,
		},
		"status": {
			Type:        schema.TypeString,
			Description: "The status of the ephemeral disk.",
			Computed:    true,
		},
		"type": {
			Type:        schema.TypeString,
			Description: "CDS disk type",
			Computed:    true,
		},
		"create_time": {
			Type:        schema.TypeString,
			Description: "CDS volume create time",
			Computed:    true,
		},
		"expire_time": {
			Type:        schema.TypeString,
			Description: "CDS volume expire time",
			Computed:    true,
		},
		"payment_timing": {
			Type:        schema.TypeString,
			Description: "payment method, support Prepaid or Postpaid",
			Computed:    true,
		},
		"snapshot_num": {
			Type:        schema.TypeString,
			Description: "CDS disk snapshot num",
			Computed:    true,
		},
		"disk_size_in_gb": {
			Type:        schema.TypeInt,
			Description: "The size(GB) of CDS.",
			Computed:    true,
		},
		"storage_type": {
			Type:        schema.TypeString,
			Description: "Storage type of the CDS.",
			Computed:    true,
		},
	}
	return &schema.Resource{
		Read: dataSourceBaiduCloudInstancesRead,

		Schema: map[string]*schema.Schema{
			"internal_ip": {
				Type:         schema.TypeString,
				Description:  "Internal ip address of the instance to retrieve.",
				Optional:     true,
				ValidateFunc: validation.SingleIP(),
			},
			"dedicated_host_id": {
				Type:        schema.TypeString,
				Description: "Dedicated host id of the instance to retrieve.",
				Optional:    true,
			},
			"zone_name": {
				Type:        schema.TypeString,
				Description: "Name of the available zone to which the instance belongs.",
				Optional:    true,
			},
			"output_file": {
				Type:        schema.TypeString,
				Description: "Output file of the instances search result",
				Optional:    true,
				ForceNew:    true,
			},
			"filter": dataSourceFiltersSchema(),

			"instances": {
				Type:        schema.TypeList,
				Description: "The result of the instances list.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instance_id": {
							Type:        schema.TypeString,
							Description: "The ID of the instance.",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the instance.",
							Computed:    true,
						},
						"instance_type": {
							Type:        schema.TypeString,
							Description: "The type of the instance.",
							Computed:    true,
						},
						"status": {
							Type:        schema.TypeString,
							Description: "The status of the instance.",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "The description of the instance.",
							Computed:    true,
						},
						"payment_timing": {
							Type:        schema.TypeString,
							Description: "The payment timing of the instance.",
							Computed:    true,
						},
						"create_time": {
							Type:        schema.TypeString,
							Description: "The creation time of the instance.",
							Computed:    true,
						},
						"expire_time": {
							Type:        schema.TypeString,
							Description: "The expire time of the instance.",
							Computed:    true,
						},
						"internal_ip": {
							Type:        schema.TypeString,
							Description: "The internal ip of the instance.",
							Computed:    true,
						},
						"public_ip": {
							Type:        schema.TypeString,
							Description: "The public ip of the instance.",
							Computed:    true,
						},
						"cpu_count": {
							Type:        schema.TypeInt,
							Description: "The cpu count of the instance.",
							Computed:    true,
						},
						"gpu_card": {
							Type:        schema.TypeString,
							Description: "The gpu card of the instance.",
							Computed:    true,
						},
						"fpga_card": {
							Type:        schema.TypeString,
							Description: "The fgpa card of the instance.",
							Computed:    true,
						},
						"card_count": {
							Type:        schema.TypeString,
							Description: "The card count of the instance.",
							Computed:    true,
						},
						"memory_capacity_in_gb": {
							Type:        schema.TypeInt,
							Description: "The memory capacity in GB of the instance.",
							Computed:    true,
						},
						"root_disk_size_in_gb": {
							Type:        schema.TypeInt,
							Description: "The system disk size in GB of the instance.",
							Computed:    true,
						},
						"root_disk_storage_type": {
							Type:        schema.TypeString,
							Description: "The system disk storage type of the instance.",
							Computed:    true,
						},
						"ephemeral_disks": {
							Type:        schema.TypeList,
							Description: "The ephemeral disks of the instance.",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: diskSchema,
							},
						},
						"cds_disks": {
							Type:        schema.TypeList,
							Description: "CDS disks of the instance.",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: diskSchema,
							},
						},
						"system_disks": {
							Type:        schema.TypeList,
							Description: "System disk of the instance.",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: diskSchema,
							},
						},
						"image_id": {
							Type:        schema.TypeString,
							Description: "The image id of the instance.",
							Computed:    true,
						},
						"network_capacity_in_mbps": {
							Type:        schema.TypeInt,
							Description: "The network capacity in Mbps of the instance.",
							Computed:    true,
						},
						"placement_policy": {
							Type:        schema.TypeString,
							Description: "The placement policy of the instance.",
							Computed:    true,
						},
						"zone_name": {
							Type:        schema.TypeString,
							Description: "The zone name of the instance.",
							Computed:    true,
						},
						"subnet_id": {
							Type:        schema.TypeString,
							Description: "The subnet ID of the instance.",
							Computed:    true,
						},
						"vpc_id": {
							Type:        schema.TypeString,
							Description: "The VPC ID of the instance.",
							Computed:    true,
						},
						"dedicated_host_id": {
							Type:        schema.TypeString,
							Description: "The dedicated host id of the instance.",
							Computed:    true,
						},
						"auto_renew": {
							Type:        schema.TypeBool,
							Description: "Whether to automatically renew.",
							Computed:    true,
						},
						"keypair_id": {
							Type:        schema.TypeString,
							Description: "The key pair id of the instance.",
							Computed:    true,
						},
						"keypair_name": {
							Type:        schema.TypeString,
							Description: "The key pair name of the instance.",
							Computed:    true,
						},
						"tags": tagsComputedSchema(),
					},
				},
			},
		},
	}
}

func dataSourceBaiduCloudInstancesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.BaiduClient)
	bccService := BccService{client}

	listArgs := &api.ListInstanceArgs{}
	if v, ok := d.GetOk("internal_ip"); ok {
		listArgs.InternalIp = v.(string)
	}
	if v, ok := d.GetOk("dedicated_host_id"); ok && v.(string) != "" {
		listArgs.DedicatedHostId = v.(string)
	}
	if v, ok := d.GetOk("zone_name"); ok && v.(string) != "" {
		listArgs.ZoneName = v.(string)
	}

	action := "List all Instance "
	instanceList, err := bccService.ListAllInstance(listArgs)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "baiducloud_instances", action, BCESDKGoERROR)
	}

	instanceMap, err := bccService.FlattenInstanceModelToMap(instanceList)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "baiducloud_instances", action, BCESDKGoERROR)
	}

	FilterDataSourceResult(d, &instanceMap)
	if err = d.Set("instances", instanceMap); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "baiducloud_instances", action, BCESDKGoERROR)
	}
	d.SetId(resource.UniqueId())

	if v, ok := d.GetOk("output_file"); ok && v.(string) != "" {
		if err := writeToFile(v.(string), instanceMap); err != nil {
			return WrapErrorf(err, DefaultErrorMsg, "baiducloud_instances", action, BCESDKGoERROR)
		}
	}

	return nil
}
