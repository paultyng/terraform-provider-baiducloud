/*
Use this data source to query RDS list.

Example Usage

```hcl
data "baiducloud_rdss" "default" {}

output "rdss" {
 value = "${data.baiducloud_rdss.default.rdss}"
}
```
*/
package baiducloud

import (
	"regexp"

	"github.com/baidubce/bce-sdk-go/services/rds"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/terraform-providers/terraform-provider-baiducloud/baiducloud/connectivity"
)

func dataSourceBaiduCloudRdss() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceBaiduCloudRdssRead,

		Schema: map[string]*schema.Schema{
			"name_regex": {
				Type:         schema.TypeString,
				Description:  "Regex pattern of the search name of rds instance",
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateNameRegex,
			},
			"instance_id": {
				Type:        schema.TypeString,
				Description: "rds instance id",
				Optional:    true,
				ForceNew:    true,
			},
			"load_parameters": {
				Type:        schema.TypeBool,
				Description: "load parameters of instance , default false",
				Optional:    true,
				ForceNew:    true,
				Default:     false,
			},
			"load_security_ips": {
				Type:        schema.TypeBool,
				Description: "load security_ips of instance , default false",
				Optional:    true,
				ForceNew:    true,
				Default:     false,
			},
			"output_file": {
				Type:        schema.TypeString,
				Description: "Output file of the instances search result",
				Optional:    true,
				ForceNew:    true,
			},
			"filter": dataSourceFiltersSchema(),

			"rdss": {
				Type:        schema.TypeList,
				Description: "The result of the instances list.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instance_id": {
							Type:        schema.TypeString,
							Description: "ID of the instance.",
							Computed:    true,
						},
						"instance_status": {
							Type:        schema.TypeString,
							Description: "Status of the instance.",
							Computed:    true,
						},
						"instance_name": {
							Type:        schema.TypeString,
							Description: "Name of the instance. Support for uppercase and lowercase letters, numbers, Chinese and special characters, such as \"-\",\"_\",\"/\",\".\", the value must start with a letter, length 1-65.",
							Computed:    true,
						},
						"engine_version": {
							Type:        schema.TypeString,
							Description: "Engine version of the instance. MySQL support 5.5、5.6、5.7, SQLServer support 2008r2、2012sp3、2016sp1, PostgreSQL support 9.4",
							Computed:    true,
						},
						"engine": {
							Type:        schema.TypeString,
							Description: "Engine of the instance. Available values are MySQL、SQLServer、PostgreSQL.",
							Required:    true,
						},
						"category": {
							Type:        schema.TypeString,
							Description: "Category of the instance. Available values are Basic、Standard(Default), only SQLServer 2012sp3 support Basic.",
							Computed:    true,
						},
						"cpu_count": {
							Type:        schema.TypeInt,
							Description: "The number of CPU",
							Computed:    true,
						},
						"memory_capacity": {
							Type:        schema.TypeFloat,
							Description: "Memory capacity(GB) of the instance.",
							Computed:    true,
						},
						"volume_capacity": {
							Type:        schema.TypeInt,
							Description: "Volume capacity(GB) of the instance",
							Computed:    true,
						},
						"vpc_id": {
							Type:        schema.TypeString,
							Description: "ID of the specific VPC",
							Computed:    true,
						},
						"subnets": {
							Type:        schema.TypeList,
							Description: "Subnets of the instance.",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"subnet_id": {
										Type:        schema.TypeString,
										Description: "ID of the subnet.",
										Computed:    true,
									},
									"zone_name": {
										Type:        schema.TypeString,
										Description: "Zone name of the subnet.",
										Computed:    true,
									},
								},
							},
						},
						"security_ips": {
							Type:        schema.TypeList,
							Description: "Security ip list",
							Computed:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"parameters": {
							Type: schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"value": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
							Computed: true,
						},
						"zone_names": {
							Type:        schema.TypeList,
							Description: "Zone name list",
							Computed:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"node_amount": {
							Type:        schema.TypeInt,
							Description: "Number of proxy node.",
							Computed:    true,
						},
						"used_storage": {
							Type:        schema.TypeFloat,
							Description: "Memory capacity(GB) of the instance to be used.",
							Computed:    true,
						},
						"create_time": {
							Type:        schema.TypeString,
							Description: "Create time of the instance.",
							Computed:    true,
						},
						"expire_time": {
							Type:        schema.TypeString,
							Description: "Expire time of the instance.",
							Computed:    true,
						},
						"address": {
							Type:        schema.TypeString,
							Description: "The domain used to access a instance.",
							Computed:    true,
						},
						"port": {
							Type:        schema.TypeInt,
							Description: "The port used to access a instance.",
							Computed:    true,
						},
						"v_net_ip": {
							Type:        schema.TypeString,
							Description: "The internal ip used to access a instance.",
							Computed:    true,
						},
						"region": {
							Type:        schema.TypeString,
							Description: "Region of the instance.",
							Computed:    true,
						},
						"instance_type": {
							Type:        schema.TypeString,
							Description: "Type of the instance,  Available values are Master, ReadReplica, RdsProxy.",
							Computed:    true,
						},
						"payment_timing": {
							Type:        schema.TypeString,
							Description: "RDS payment timing",
							Computed:    true,
						},
						"source_instance_id": {
							Type:        schema.TypeString,
							Description: "ID of the master instance",
							Computed:    true,
						},
						"source_region": {
							Type:        schema.TypeString,
							Description: "Region of the master instance",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceBaiduCloudRdssRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.BaiduClient)
	rdsService := RdsService{client}

	action := "List all rds instances"
	rdsList := make([]rds.Instance, 0)
	listArgs := &rds.ListRdsArgs{}
	if value, ok := d.GetOk("instance_id"); ok {
		id := value.(string)
		if len(id) > 0 {
			instance, err := rdsService.GetInstanceDetail(id)
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, "baiducloud_rdss", action, BCESDKGoERROR)
			} else {
				rdsList = append(rdsList, *instance)
			}
		}
	} else {
		allList, err := rdsService.ListAllInstances(listArgs)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, "baiducloud_rdss", action, BCESDKGoERROR)
		} else {
			rdsList = allList
		}
	}

	var nameRegex string
	var specNameRegex *regexp.Regexp

	if value, ok := d.GetOk("name_regex"); ok {
		nameRegex = value.(string)
		if len(nameRegex) > 0 {
			specNameRegex = regexp.MustCompile(nameRegex)
		}
	}

	rdsMap := make([]map[string]interface{}, 0, len(rdsList))
	load_parameters := false
	if value, ok := d.GetOk("load_parameters"); ok {
		load_parameters = value.(bool)
	}
	load_security_ips := false
	if value, ok := d.GetOk("load_security_ips"); ok {
		load_security_ips = value.(bool)
	}
	for _, e := range rdsList {
		if len(nameRegex) > 0 && specNameRegex != nil {
			if !specNameRegex.MatchString(e.InstanceName) {
				continue
			}
		}

		instance := map[string]interface{}{
			"instance_id":        e.InstanceId,
			"instance_name":      e.InstanceName,
			"instance_status":    e.InstanceStatus,
			"source_instance_id": e.SourceInstanceId,
			"source_region":      e.SourceRegion,
			"engine":             e.Engine,
			"engine_version":     e.EngineVersion,
			"category":           e.Category,
			"cpu_count":          e.CpuCount,
			"memory_capacity":    e.MemoryCapacity,
			"volume_capacity":    e.VolumeCapacity,
			"node_amount":        e.NodeAmount,
			"used_storage":       e.UsedStorage,
			"create_time":        e.InstanceCreateTime,
			"expire_time":        e.InstanceExpireTime,
			"address":            e.Endpoint.Address,
			"port":               e.Endpoint.Port,
			"v_net_ip":           e.Endpoint.VnetIp,
			"region":             e.Region,
			"instance_type":      e.InstanceType,
			"payment_timing":     e.PaymentTiming,
			"zone_names":         e.ZoneNames,
			"vpc_id":             e.VpcId,
			"subnets":            rdsService.TransRdsSubnetsToSchema(e.Subnets),
		}

		if load_parameters {
			paramResult, err := rdsService.ListParameters(e.InstanceId)
			if err == nil {
				instance["parameters"] = rdsService.TransRdsParametersToSchema(paramResult.Parameters)
			}
		}
		if load_security_ips {
			ipResult, err := rdsService.ListSecurityIps(e.InstanceId)
			if err == nil {
				instance["security_ips"] = ipResult.SecurityIps
			}
		}
		rdsMap = append(rdsMap, instance)
	}

	FilterDataSourceResult(d, &rdsMap)

	addDebug("List filtered rds instances", rdsMap)
	if err := d.Set("rdss", rdsMap); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "baiducloud_rdss", action, BCESDKGoERROR)
	}
	d.SetId(resource.UniqueId())
	if v, ok := d.GetOk("output_file"); ok && v.(string) != "" {
		if err := writeToFile(v.(string), rdsMap); err != nil {
			return WrapErrorf(err, DefaultErrorMsg, "baiducloud_rdss", action, BCESDKGoERROR)
		}
	}
	return nil
}
