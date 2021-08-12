/*
Use this resource to get information about a RDS readonly instance.

~> **NOTE:** The terminate operation of rds readonly instance does NOT take effect immediately，maybe takes for several minites.

Example Usage

```hcl
resource "baiducloud_rds_readonly_instance" "default" {
    billing = {
        payment_timing        = "Postpaid"
    }
    source_instance_id        = baiducloud_rds_instance.default.instance_id
    cpu_count                 = 1
    memory_capacity           = 1
    volume_capacity           = 5
}
```

Import

RDS readonly instance can be imported, e.g.

```hcl
$ terraform import baiducloud_rds_readonly_instance.default id
```
*/
package baiducloud

import (
	"time"

	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/rds"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"

	"github.com/terraform-providers/terraform-provider-baiducloud/baiducloud/connectivity"
)

func resourceBaiduCloudRdsReadOnlyInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceBaiduCloudRdsReadOnlyInstanceCreate,
		Read:   resourceBaiduCloudRdsReadOnlyInstanceRead,
		Update: resourceBaiduCloudRdsReadOnlyInstanceUpdate,
		Delete: resourceBaiduCloudRdsReadOnlyInstanceDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"source_instance_id": {
				Type:        schema.TypeString,
				Description: "ID of the master instance",
				Required:    true,
				ForceNew:    true,
			},
			"instance_name": {
				Type:        schema.TypeString,
				Description: "Name of the instance. Support for uppercase and lowercase letters, numbers, Chinese and special characters, such as \"-\",\"_\",\"/\",\".\", the value must start with a letter, length 1-65.",
				Optional:    true,
				Computed:    true,
			},
			"category": {
				Type:         schema.TypeString,
				Description:  "Category of the instance. Available values are Basic、Standard(Default), only SQLServer 2012sp3 support Basic.",
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"Standard", "Basic"}, false),
			},
			"cpu_count": {
				Type:         schema.TypeInt,
				Description:  "The number of CPU",
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"memory_capacity": {
				Type:         schema.TypeFloat,
				Description:  "Memory capacity(GB) of the instance.",
				Required:     true,
				ValidateFunc: validation.FloatBetween(1, 480),
			},
			"volume_capacity": {
				Type:         schema.TypeInt,
				Description:  "Volume capacity(GB) of the instance",
				Required:     true,
				ValidateFunc: validation.IntAtLeast(5),
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Description: "ID of the specific VPC",
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
			"subnets": {
				Type:        schema.TypeList,
				Description: "Subnets of the instance.",
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_id": {
							Type:        schema.TypeString,
							Description: "ID of the subnet.",
							Optional:    true,
							Computed:    true,
							ForceNew:    true,
						},
						"zone_name": {
							Type:        schema.TypeString,
							Description: "Zone name of the subnet.",
							Optional:    true,
							Computed:    true,
							ForceNew:    true,
						},
					},
				},
			},
			"zone_names": {
				Type:        schema.TypeList,
				Description: "Zone name list",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"tags": tagsSchema(),
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
			"billing": {
				Type:        schema.TypeMap,
				Description: "Billing information of the Rds.",
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"payment_timing": {
							Type:         schema.TypeString,
							Description:  "Payment timing of billing, which can be Prepaid or Postpaid. The default is Postpaid.",
							Required:     true,
							Default:      PaymentTimingPostpaid,
							ValidateFunc: validatePaymentTiming(),
						},
						"reservation": {
							Type:             schema.TypeMap,
							Description:      "Reservation of the Rds.",
							Optional:         true,
							DiffSuppressFunc: postPaidDiffSuppressFunc,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"reservation_length": {
										Type:             schema.TypeInt,
										Description:      "The reservation length that you will pay for your resource. It is valid when payment_timing is Prepaid. Valid values: [1, 2, 3, 4, 5, 6, 7, 8, 9, 12, 24, 36].",
										Required:         true,
										Default:          1,
										ValidateFunc:     validateReservationLength(),
										DiffSuppressFunc: postPaidDiffSuppressFunc,
									},
									"reservation_time_unit": {
										Type:             schema.TypeString,
										Description:      "The reservation time unit that you will pay for your resource. It is valid when payment_timing is Prepaid. The value can only be month currently, which is also the default value.",
										Required:         true,
										Default:          "Month",
										ValidateFunc:     validateReservationUnit(),
										DiffSuppressFunc: postPaidDiffSuppressFunc,
									},
								},
							},
						},
					},
				},
			},
			"payment_timing": {
				Type:        schema.TypeString,
				Description: "RDS payment timing",
				Computed:    true,
			},
		},
	}
}

func resourceBaiduCloudRdsReadOnlyInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.BaiduClient)
	rdsService := RdsService{client}

	createRdsArgs, err := buildBaiduCloudRdsReadOnlyInstanceArgs(d, meta)
	if err != nil {
		return WrapError(err)
	}

	action := "Create RDS Instance " + createRdsArgs.InstanceName
	addDebug(action, createRdsArgs)

	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		raw, err := client.WithRdsClient(func(rdsClient *rds.Client) (interface{}, error) {
			return rdsClient.CreateReadReplica(createRdsArgs)
		})
		if err != nil {
			if IsExceptedErrors(err, []string{bce.EINTERNAL_ERROR}) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, raw)
		response, _ := raw.(*rds.CreateResult)
		d.SetId(response.InstanceIds[0])
		return nil
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "baiducloud_rds_readonly_instance", action, BCESDKGoERROR)
	}

	stateConf := buildStateConf(
		[]string{RDSStatusCreating},
		[]string{RDSStatusRunning},
		d.Timeout(schema.TimeoutCreate),
		rdsService.InstanceStateRefresh(d.Id(), []string{}),
	)
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "baiducloud_rds_readonly_instance", action, BCESDKGoERROR)
	}

	return resourceBaiduCloudRdsReadOnlyInstanceRead(d, meta)
}

func resourceBaiduCloudRdsReadOnlyInstanceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.BaiduClient)
	rdsService := RdsService{client}
	instanceID := d.Id()
	action := "Query RDS Instance " + instanceID

	raw, err := client.WithRdsClient(func(rdsClient *rds.Client) (interface{}, error) {
		return rdsClient.GetDetail(instanceID)
	})

	addDebug(action, raw)

	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, "baiducloud_rds_readonly_instance", action, BCESDKGoERROR)
	}

	result, _ := raw.(*rds.Instance)

	d.Set("instance_id", result.InstanceId)
	d.Set("instance_name", result.InstanceName)
	d.Set("engine", result.Engine)
	d.Set("engine_version", result.EngineVersion)
	d.Set("category", result.Category)
	d.Set("instance_status", result.InstanceStatus)
	d.Set("source_instance_id", result.SourceInstanceId)
	d.Set("source_region", result.SourceRegion)
	d.Set("cpu_count", result.CpuCount)
	d.Set("memory_capacity", result.MemoryCapacity)
	d.Set("volume_capacity", result.VolumeCapacity)
	d.Set("node_amount", result.NodeAmount)
	d.Set("used_storage", result.UsedStorage)
	d.Set("create_time", result.InstanceCreateTime)
	d.Set("expire_time", result.InstanceExpireTime)
	d.Set("region", result.Region)
	d.Set("instance_type", result.InstanceType)
	d.Set("payment_timing", result.PaymentTiming)
	d.Set("zone_names", result.ZoneNames)
	d.Set("vpc_id", result.VpcId)
	d.Set("port", result.Endpoint.Port)
	d.Set("address", result.Endpoint.Address)
	d.Set("v_net_ip", result.Endpoint.VnetIp)
	d.Set("subnets", rdsService.TransRdsSubnetsToSchema(result.Subnets))

	return nil
}

func resourceBaiduCloudRdsReadOnlyInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	instanceID := d.Id()

	d.Partial(true)

	// resize instance
	if err := resizeRds(d, meta, instanceID); err != nil {
		return err
	}

	d.Partial(false)

	return resourceBaiduCloudRdsReadOnlyInstanceRead(d, meta)
}

func resourceBaiduCloudRdsReadOnlyInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.BaiduClient)

	instanceId := d.Id()
	action := "Delete RDS Instance " + instanceId

	err := resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		raw, err := client.WithRdsClient(func(rdsClient *rds.Client) (interface{}, error) {
			return instanceId, rdsClient.DeleteRds(instanceId)
		})
		if err != nil {
			if IsExceptedErrors(err, []string{InvalidInstanceStatus, bce.EINTERNAL_ERROR}) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, raw)
		return nil
	})
	if err != nil {
		if IsExceptedErrors(err, []string{InvalidInstanceStatus, InstanceNotExist, bce.EINTERNAL_ERROR}) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, "baiducloud_rds_readonly_instance", action, BCESDKGoERROR)
	}

	return nil
}

func buildBaiduCloudRdsReadOnlyInstanceArgs(d *schema.ResourceData, meta interface{}) (*rds.CreateReadReplicaArgs, error) {
	request := &rds.CreateReadReplicaArgs{
		ClientToken: buildClientToken(),
	}

	if v, ok := d.GetOk("billing"); ok {
		billing := v.(map[string]interface{})
		billingRequest := rds.Billing{
			PaymentTiming: "",
			Reservation:   rds.Reservation{},
		}
		if p, ok := billing["payment_timing"]; ok {
			paymentTiming := p.(string)
			billingRequest.PaymentTiming = paymentTiming
		}
		if billingRequest.PaymentTiming == PaymentTimingPostpaid {
			if r, ok := billing["reservation"]; ok {
				reservation := r.(map[string]interface{})
				if reservationLength, ok := reservation["reservation_length"]; ok {
					billingRequest.Reservation.ReservationLength = reservationLength.(int)
				}
				if reservationTimeUnit, ok := reservation["reservation_time_unit"]; ok {
					billingRequest.Reservation.ReservationTimeUnit = reservationTimeUnit.(string)
				}
			}
		}
		request.Billing = billingRequest
	}

	if instanceName, ok := d.GetOk("instance_name"); ok {
		request.InstanceName = instanceName.(string)
	}

	if sourceInstanceId, ok := d.GetOk("source_instance_id"); ok {
		request.SourceInstanceId = sourceInstanceId.(string)
	}

	if cpuCount, ok := d.GetOk("cpu_count"); ok {
		request.CpuCount = cpuCount.(int)
	}

	if memoryCapacity, ok := d.GetOk("memory_capacity"); ok {
		request.MemoryCapacity = memoryCapacity.(float64)
	}

	if volumeCapacity, ok := d.GetOk("volume_capacity"); ok {
		request.VolumeCapacity = volumeCapacity.(int)
	}

	if isDirectPay, ok := d.GetOk("is_direct_pay"); ok {
		request.IsDirectPay = isDirectPay.(bool)
	}

	if purchaseCount, ok := d.GetOk("purchase_count"); ok {
		request.PurchaseCount = purchaseCount.(int)
	}

	if vpcID, ok := d.GetOk("vpc_id"); ok {
		request.VpcId = vpcID.(string)
	}

	if v, ok := d.GetOk("subnets"); ok {
		subnetList := v.([]interface{})
		subnetRequests := make([]rds.SubnetMap, len(subnetList))
		for id := range subnetList {
			subnet := subnetList[id].(map[string]interface{})

			subnetRequest := rds.SubnetMap{
				SubnetId: subnet["subnet_id"].(string),
				ZoneName: subnet["zone_name"].(string),
			}

			subnetRequests[id] = subnetRequest
		}
		request.Subnets = subnetRequests
	}

	return request, nil

}
