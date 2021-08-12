package baiducloud

import (
	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/bcc"
	"github.com/baidubce/bce-sdk-go/services/bcc/api"
	"github.com/hashicorp/terraform/helper/resource"
)

func (s *BccService) CDSVolumeStateRefreshFunc(id string, failState []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		action := "Query CDS volume " + id
		raw, err := s.client.WithBccClient(func(bccClient *bcc.Client) (i interface{}, e error) {
			return bccClient.GetCDSVolumeDetail(id)
		})
		if err != nil {
			addDebug(action, err)
			if NotFoundError(err) {
				return 0, "", nil
			}

			return nil, "", WrapErrorf(err, DefaultErrorMsg, "baiducloud_cds", action, BCESDKGoERROR)
		}
		if raw == nil {
			return nil, "", nil
		}
		addDebug(action, raw)
		result, _ := raw.(*api.GetVolumeDetailResult)
		status := result.Volume.Status
		for _, state := range failState {
			if string(status) == state {
				return string(status), string(status), WrapError(Error(GetFailTargetStatus, string(status)))
			}
		}
		return result, string(status), nil
	}
}

func (s *BccService) AttachCDSVolume(volumeId, instanceId string) error {
	args := &api.AttachVolumeArgs{
		InstanceId: instanceId,
	}

	action := "Attach CDS volume " + volumeId + " to instance " + instanceId

	raw, err := s.client.WithBccClient(func(client *bcc.Client) (i interface{}, e error) {
		return client.AttachCDSVolume(volumeId, args)
	})
	addDebug(action, raw)

	if err != nil {
		if !IsExceptedErrors(err, []string{bce.EINTERNAL_ERROR}) {
			return WrapErrorf(err, DefaultErrorMsg, "baiducloud_cds", action, BCESDKGoERROR)
		}
	}

	stateConf := buildStateConf(
		append(CDSProcessingStatus, string(api.VolumeStatusAVAILABLE)),
		[]string{string(api.VolumeStatusINUSE)},
		DefaultTimeout,
		s.CDSVolumeStateRefreshFunc(volumeId, CDSFailedStatus))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapError(err)
	}

	return nil
}

func (s *BccService) DetachCDSVolume(volumeId, instanceId string) error {
	args := &api.DetachVolumeArgs{
		InstanceId: instanceId,
	}

	action := "Detach CDS volume " + volumeId + " to instance " + instanceId

	raw, err := s.client.WithBccClient(func(client *bcc.Client) (i interface{}, e error) {
		return nil, client.DetachCDSVolume(volumeId, args)
	})

	addDebug(action, raw)

	if err != nil && !NotFoundError(err) {
		// if before detach, relate resource like instance has been deleted,
		// may return DiskNotAttachedInstance error
		// so we check cds status again
		cdsDetail, errDetail := s.GetCDSVolumeDetail(volumeId)
		if errDetail != nil {
			if NotFoundError(errDetail) {
				return nil
			}

			// return detach err
			return WrapErrorf(err, DefaultErrorMsg, "baiducloud_cds", action, BCESDKGoERROR)
		}

		if stringInSlice([]string{
			string(api.VolumeStatusCREATING),
			string(api.VolumeStatusATTACHING),
			string(api.VolumeStatusINUSE)}, cdsDetail.SourceSnapshotId) {
			return WrapErrorf(err, DefaultErrorMsg, "baiducloud_cds", action, BCESDKGoERROR)
		}
		return nil
	}

	stateConf := buildStateConf(
		append(CDSProcessingStatus, string(api.VolumeStatusINUSE)),
		[]string{string(api.VolumeStatusAVAILABLE)},
		DefaultTimeout,
		s.CDSVolumeStateRefreshFunc(volumeId, CDSFailedStatus))
	if _, err := stateConf.WaitForState(); err != nil && !NotFoundError(err) {
		return WrapError(err)
	}

	return nil
}

func (s *BccService) ModifyCDSVolume(volumeId, name, desc string) error {
	args := &api.ModifyCSDVolumeArgs{
		CdsName: name,
		Desc:    desc,
	}
	action := "Modify CDS volume " + volumeId

	_, err := s.client.WithBccClient(func(client *bcc.Client) (i interface{}, e error) {
		return nil, client.ModifyCDSVolume(volumeId, args)
	})
	addDebug(action, nil)

	if err != nil {
		if err != nil {
			if !IsExceptedErrors(err, []string{bce.EINTERNAL_ERROR}) {
				return WrapErrorf(err, DefaultErrorMsg, "baiducloud_cds", action, BCESDKGoERROR)
			}
		}
	}

	return nil
}

func (s *BccService) ModifyChargeTypeCDSVolume(volumeId string, args *api.ModifyChargeTypeCSDVolumeArgs) error {
	action := "Modify CDS volume " + volumeId + "charge type to " + string(args.Billing.PaymentTiming)

	_, err := s.client.WithBccClient(func(client *bcc.Client) (i interface{}, e error) {
		return nil, client.ModifyChargeTypeCDSVolume(volumeId, args)
	})
	addDebug(action, nil)

	if err != nil {
		if err != nil {
			if !IsExceptedErrors(err, []string{bce.EINTERNAL_ERROR}) {
				return WrapErrorf(err, DefaultErrorMsg, "baiducloud_cds", action, BCESDKGoERROR)
			}
		}
	}

	stateConf := buildStateConf(
		CDSProcessingStatus,
		[]string{string(api.VolumeStatusAVAILABLE), string(api.VolumeStatusINUSE)},
		DefaultTimeout,
		s.CDSVolumeStateRefreshFunc(volumeId, CDSFailedStatus))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapError(err)
	}

	return nil
}

func (s *BccService) ResizeCDSVolume(volumeId string, newSize int, volumeType api.StorageType) error {
	args := &api.ResizeCSDVolumeArgs{
		NewCdsSizeInGB: newSize,
		NewVolumeType:  volumeType,
		ClientToken:    buildClientToken(),
	}
	action := "Resize CDS volume " + volumeId

	_, err := s.client.WithBccClient(func(client *bcc.Client) (i interface{}, e error) {
		return nil, client.ResizeCDSVolume(volumeId, args)
	})
	addDebug(action, nil)

	if err != nil {
		if err != nil {
			if !IsExceptedErrors(err, []string{bce.EINTERNAL_ERROR}) {
				return WrapErrorf(err, DefaultErrorMsg, "baiducloud_cds", action, BCESDKGoERROR)
			}
		}
	}

	stateConf := buildStateConf(
		CDSProcessingStatus,
		[]string{string(api.VolumeStatusAVAILABLE), string(api.VolumeStatusINUSE)},
		DefaultTimeout,
		s.CDSVolumeStateRefreshFunc(volumeId, CDSFailedStatus))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapError(err)
	}

	return nil
}

func (s *BccService) GetCDSVolumeDetail(volumeId string) (*api.VolumeModel, error) {
	action := "Query CDS volume " + volumeId
	raw, err := s.client.WithBccClient(func(bccClient *bcc.Client) (i interface{}, e error) {
		return bccClient.GetCDSVolumeDetail(volumeId)
	})
	addDebug(action, err)
	if err != nil {
		return nil, WrapError(err)
	}

	return raw.(*api.GetVolumeDetailResult).Volume, nil
}

func (s *BccService) ListAllCDSVolumeDetail(args *api.ListCDSVolumeArgs) ([]api.VolumeModel, error) {
	action := "List all CDS volume Detail"

	cdsList := make([]api.VolumeModel, 0)
	for {
		raw, err := s.client.WithBccClient(func(bccClient *bcc.Client) (i interface{}, e error) {
			return bccClient.ListCDSVolume(args)
		})
		if err != nil {
			return nil, WrapErrorf(err, DefaultErrorMsg, "baiducloud_cdss", action, BCESDKGoERROR)
		}
		addDebug(action, raw)

		result, _ := raw.(*api.ListCDSVolumeResult)
		for _, vol := range result.Volumes {
			if vol.Type != "Cds" {
				continue
			}
			cdsList = append(cdsList, vol)
		}

		if !result.IsTruncated {
			break
		}
		args.Marker = result.NextMarker
	}

	for idx, c := range cdsList {
		cdsDetail, err := s.GetCDSVolumeDetail(c.Id)
		if err != nil {
			return nil, WrapError(err)
		}

		cdsList[idx] = *cdsDetail
	}

	return cdsList, nil
}

func (s *BccService) FlattenVolumeAttachmentModelToMap(attachments []api.VolumeAttachmentModel) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(attachments))

	for _, attach := range attachments {
		result = append(result, map[string]interface{}{
			"volume_id":   attach.VolumeId,
			"instance_id": attach.InstanceId,
			"device":      attach.Device,
			"serial":      attach.Serial,
		})
	}

	return result
}

func (s *BccService) FlattenCDSVolumeModelToMapForInstance(cdsList []api.VolumeModel) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(cdsList))

	for _, c := range cdsList {
		cdsMap := map[string]interface{}{
			"cds_id":          c.Id,
			"name":            c.Name,
			"disk_size_in_gb": c.DiskSizeInGB,
			"payment_timing":  c.PaymentTiming,
			"create_time":     c.CreateTime,
			"expire_time":     c.ExpireTime,
			"status":          c.Status,
			"type":            c.Type,
			"storage_type":    c.StorageType,
			"description":     c.Desc,
			"snapshot_num":    c.SnapshotNum,
		}
		result = append(result, cdsMap)
	}

	return result
}

func (s *BccService) FlattenCDSVolumeModelToMap(cdsList []api.VolumeModel) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(cdsList))

	for _, c := range cdsList {
		cdsMap := map[string]interface{}{
			"cds_id":           c.Id,
			"name":             c.Name,
			"disk_size_in_gb":  c.DiskSizeInGB,
			"payment_timing":   c.PaymentTiming,
			"create_time":      c.CreateTime,
			"expire_time":      c.ExpireTime,
			"status":           c.Status,
			"type":             c.Type,
			"storage_type":     c.StorageType,
			"description":      c.Desc,
			"attachments":      s.FlattenVolumeAttachmentModelToMap(c.Attachments),
			"zone_name":        c.ZoneName,
			"tags":             flattenTagsToMap(c.Tags),
			"is_system_volume": c.IsSystemVolume,
			"region_id":        c.RegionId,
			"snapshot_num":     c.SnapshotNum,
		}

		if c.AutoSnapshotPolicy != nil {
			cdsMap["auto_snapshot_policy"] = s.FlattenAutoSnapshotPolicyModelToMap([]api.AutoSnapshotPolicyModel{*c.AutoSnapshotPolicy})
		}

		result = append(result, cdsMap)
	}

	return result
}
