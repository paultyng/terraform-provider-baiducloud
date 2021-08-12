/*
Provide a resource to create a CDS.

Example Usage

```hcl
resource "baiducloud_cds" "default" {
  name                    = "terraformCreate"
  description             = "terraform create cds"
  payment_timing          = "Postpaid"
  auto_snapshot_policy_id = "asp-xyYk0XFC"
  snapshot_id             = "s-WTGlKBR1"
}
```

Import

CDS can be imported, e.g.

```hcl
$ terraform import baiducloud_cds.default id
```
*/
package baiducloud

import (
	"fmt"
	"time"

	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/bcc"
	"github.com/baidubce/bce-sdk-go/services/bcc/api"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"

	"github.com/terraform-providers/terraform-provider-baiducloud/baiducloud/connectivity"
)

func resourceBaiduCloudCDS() *schema.Resource {
	return &schema.Resource{
		Create: resourceBaiduCloudCDSCreate,
		Read:   resourceBaiduCloudCDSRead,
		Update: resourceBaiduCloudCDSUpdate,
		Delete: resourceBaiduCloudCDSDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "CDS volume name",
				Optional:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "CDS volume description",
				Optional:    true,
			},
			"disk_size_in_gb": {
				Type:         schema.TypeInt,
				Description:  "CDS disk size, support between 5 and 32765, if snapshot_id not set, this parameter is required.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(5, 32768),
			},
			"storage_type": {
				Type:         schema.TypeString,
				Description:  "CDS dist storage type, support hp1, std1, cloud_hp1 and hdd, default hp1, see https://cloud.baidu.com/doc/BCC/s/6jwvyo0q2/#storagetype for detail",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateStorageType(),
			},
			"payment_timing": {
				Type:         schema.TypeString,
				Description:  "payment method, support Prepaid or Postpaid",
				Required:     true,
				ValidateFunc: validatePaymentTiming(),
			},
			"reservation_length": {
				Type:             schema.TypeInt,
				Description:      "Prepaid reservation length, support [1, 2, 3, 4, 5, 6, 7, 8, 9, 12, 24, 36], only useful when payment_timing is Prepaid",
				Optional:         true,
				ValidateFunc:     validateReservationLength(),
				DiffSuppressFunc: postPaidDiffSuppressFunc,
			},
			"reservation_time_unit": {
				Type:             schema.TypeString,
				Description:      "Prepaid reservation time unit, only support Month now",
				Optional:         true,
				ValidateFunc:     validateReservationUnit(),
				DiffSuppressFunc: postPaidDiffSuppressFunc,
			},
			"snapshot_id": {
				Type:        schema.TypeString,
				Description: "Snapshot id, support create cds use snapshot, when set this parameter, cds_disk_size is ignored",
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
			},
			"manual_snapshot": {
				Type:        schema.TypeBool,
				Description: "Delete relate snapshot when release this cds volume",
				Optional:    true,
				Default:     false,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if v, ok := d.GetOk("snapshot_id"); ok && v.(string) != "" {
						return false
					}

					return true
				},
			},
			"zone_name": {
				Type:        schema.TypeString,
				Description: "Zone name",
				Optional:    true,
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
			"type": {
				Type:        schema.TypeString,
				Description: "CDS volume type",
				Computed:    true,
			},
			"auto_snapshot_policy_id": {
				Type:        schema.TypeString,
				Description: "CDS bind Auto Snapshot policy id",
				Computed:    true,
			},
			"auto_snapshot": {
				Type:        schema.TypeBool,
				Description: "Delete relate auto snapshot when release this cds volume",
				Optional:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: "CDS volume status",
				Computed:    true,
			},
		},
	}
}

func resourceBaiduCloudCDSCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.BaiduClient)
	bccService := BccService{client}

	args, err := buildBaiduCloudCreateCDSArgs(d, meta)
	if err != nil {
		return WrapError(err)
	}
	action := "Create CDS volume"

	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		raw, err := client.WithBccClient(func(client *bcc.Client) (i interface{}, e error) {
			return client.CreateCDSVolume(args)
		})
		addDebug(action, raw)

		if err != nil {
			if IsExceptedErrors(err, []string{bce.EINTERNAL_ERROR}) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}

		response := raw.(*api.CreateCDSVolumeResult)
		d.SetId(response.VolumeIds[0])

		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "baiducloud_cds", action, BCESDKGoERROR)
	}

	stateConf := buildStateConf(
		[]string{string(api.VolumeStatusCREATING)},
		[]string{string(api.VolumeStatusAVAILABLE)},
		d.Timeout(schema.TimeoutCreate),
		bccService.CDSVolumeStateRefreshFunc(d.Id(), CDSFailedStatus))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapError(err)
	}

	return resourceBaiduCloudCDSRead(d, meta)
}

func resourceBaiduCloudCDSRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.BaiduClient)

	id := d.Id()
	action := "Query CDS volume detail " + id

	raw, err := client.WithBccClient(func(client *bcc.Client) (i interface{}, e error) {
		return client.GetCDSVolumeDetail(id)
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "baiducloud_cds", action, BCESDKGoERROR)
	}

	volume := raw.(*api.GetVolumeDetailResult).Volume
	d.Set("create_time", volume.CreateTime)
	d.Set("expire_time", volume.ExpireTime)
	d.Set("status", volume.Status)
	d.Set("disk_size_in_gb", volume.DiskSizeInGB)
	d.Set("type", volume.Type)
	d.Set("storage_type", volume.StorageType)
	d.Set("status", volume.Status)
	d.Set("name", volume.Name)
	d.Set("payment_timing", volume.PaymentTiming)
	d.Set("zone_name", volume.ZoneName)
	d.Set("snapshot_id", volume.SourceSnapshotId)

	if volume.AutoSnapshotPolicy != nil {
		d.Set("auto_snapshot_policy_id", volume.AutoSnapshotPolicy.Id)
	}

	return nil
}

func resourceBaiduCloudCDSUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.BaiduClient)
	bccService := BccService{client}

	d.Partial(true)
	id := d.Id()

	if d.HasChange("name") || d.HasChange("description") {
		name := ""
		if v, ok := d.GetOk("name"); ok && v.(string) != "" {
			name = v.(string)
		}
		desc := ""
		if v, ok := d.GetOk("description"); ok && v.(string) != "" {
			desc = v.(string)
		}

		err := bccService.ModifyCDSVolume(id, name, desc)
		if err != nil {
			return WrapError(err)
		}

		d.SetPartial("name")
	}

	if d.HasChange("payment_timing") {
		args := &api.ModifyChargeTypeCSDVolumeArgs{
			Billing: &api.Billing{
				PaymentTiming: api.PaymentTimingType(d.Get("payment_timing").(string)),
			},
		}
		if args.Billing.PaymentTiming == api.PaymentTimingPrePaid {
			if v, ok := d.GetOk("reservation_length"); ok {
				args.Billing.Reservation.ReservationLength = v.(int)
			} else {
				return WrapError(fmt.Errorf("reservation_length is required if payment_timing set Prepaid"))
			}
		}

		if err := bccService.ModifyChargeTypeCDSVolume(id, args); err != nil {
			return WrapError(err)
		}

		d.SetPartial("payment_timing")
	}

	if d.HasChange("disk_size_in_gb") {
		o, n := d.GetChange("disk_size_in_gb")
		oldSize := o.(int)
		newSize := n.(int)
		if oldSize > newSize {
			return Error("Cds only support scaling size, old size should bigger than new size")
		}

		storageType := api.StorageTypeHP1
		if d.HasChange("storage_type") {
			_, n := d.GetChange("disk_size_in_gb")
			storageType = api.StorageType(n.(string))
		} else if v, ok := d.GetOk("storage_type"); ok && v.(string) != "" {
			storageType = api.StorageType(v.(string))
		}

		if err := bccService.ResizeCDSVolume(id, newSize, storageType); err != nil {
			return WrapError(err)
		}

		d.SetPartial("disk_size_in_gb")
	}

	d.Partial(false)
	return resourceBaiduCloudCDSRead(d, meta)
}

func resourceBaiduCloudCDSDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.BaiduClient)
	bccService := BccService{client: client}
	id := d.Id()

	action := "Delete CDS volume " + id
	args := &api.DeleteCDSVolumeArgs{}
	if v, ok := d.GetOk("manual_snapshot"); ok && v.(bool) {
		args.ManualSnapshot = "on"
	} else {
		args.ManualSnapshot = "off"
	}

	if v, ok := d.GetOk("auto_snapshot"); ok && v.(bool) {
		args.AutoSnapshot = "on"
	} else {
		args.AutoSnapshot = "off"
	}

	err := resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		detail, errDetail := bccService.GetCDSVolumeDetail(id)
		if errDetail != nil {
			return resource.NonRetryableError(errDetail)
		}

		if stringInSlice(append(CDSProcessingStatus, string(api.VolumeStatusINUSE)), string(detail.Status)) {
			return resource.RetryableError(cdsStillInUsed)
		}

		_, err := client.WithBccClient(func(client *bcc.Client) (i interface{}, e error) {
			return id, client.DeleteCDSVolumeNew(id, args)
		})
		addDebug(action, args)

		if err != nil {
			if IsExceptedErrors(err, []string{bce.EINTERNAL_ERROR}) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if err != nil {
		if IsExceptedErrors(err, ObjectNotFound) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, "baiducloud_cds", action, BCESDKGoERROR)
	}

	return nil
}

func buildBaiduCloudCreateCDSArgs(d *schema.ResourceData, meta interface{}) (*api.CreateCDSVolumeArgs, error) {
	result := &api.CreateCDSVolumeArgs{
		PurchaseCount: 1,
		Billing:       &api.Billing{},
		ClientToken:   buildClientToken(),
	}

	if v, ok := d.GetOk("storage_type"); ok && v.(string) != "" {
		result.StorageType = api.StorageType(v.(string))
	}

	if v, ok := d.GetOk("snapshot_id"); ok && v.(string) != "" {
		result.SnapshotId = v.(string)
	}

	if v, ok := d.GetOk("zone_name"); ok && v.(string) != "" {
		result.ZoneName = v.(string)
	}

	if v, ok := d.GetOk("name"); ok && v.(string) != "" {
		result.Name = v.(string)
	}

	if v, ok := d.GetOk("description"); ok && v.(string) != "" {
		result.Description = v.(string)
	}

	if result.SnapshotId == "" {
		if v, ok := d.GetOk("disk_size_in_gb"); ok {
			result.CdsSizeInGB = v.(int)
		} else {
			return nil, fmt.Errorf("disk_size_in_gb is required if snapshot_id is not set")
		}
	}

	if v, ok := d.GetOk("payment_timing"); ok {
		result.Billing.PaymentTiming = api.PaymentTimingType(v.(string))
	}

	if result.Billing.PaymentTiming == api.PaymentTimingPrePaid {
		if v, ok := d.GetOk("reservation_length"); ok {
			result.Billing.Reservation.ReservationLength = v.(int)
		}

		if v, ok := d.GetOk("reservation_time_unit"); ok {
			result.Billing.Reservation.ReservationTimeUnit = v.(string)
		}
	}

	return result, nil
}
