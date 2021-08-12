package baiducloud

import (
	"time"

	"github.com/baidubce/bce-sdk-go/services/bcc"
	"github.com/baidubce/bce-sdk-go/services/bcc/api"
	"github.com/hashicorp/terraform/helper/resource"

	"github.com/terraform-providers/terraform-provider-baiducloud/baiducloud/connectivity"
)

type BccService struct {
	client *connectivity.BaiduClient
}

func (s *BccService) InstanceStateRefresh(instanceId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		action := "Query BCC instance " + instanceId
		raw, err := s.client.WithBccClient(func(bccClient *bcc.Client) (i interface{}, e error) {
			return bccClient.GetInstanceDetail(instanceId)
		})
		addDebug(action, raw)
		if err != nil {
			if NotFoundError(err) {
				return 0, string(api.InstanceStatusDeleted), nil
			}
			return nil, "", WrapErrorf(err, DefaultErrorMsg, "baiducloud_instance", action, BCESDKGoERROR)
		}
		if raw == nil {
			return nil, "", nil
		}
		addDebug(action, raw)
		result, _ := raw.(*api.GetInstanceDetailResult)
		return result, string(result.Instance.Status), nil
	}
}

func (s *BccService) ListAllInstance(args *api.ListInstanceArgs) ([]api.InstanceModel, error) {
	result := make([]api.InstanceModel, 0)

	action := "List all BCC instance "
	for {
		raw, err := s.client.WithBccClient(func(bccClient *bcc.Client) (i interface{}, e error) {
			return bccClient.ListInstances(args)
		})
		if err != nil {
			return nil, err
		}
		addDebug(action, raw)

		response := raw.(*api.ListInstanceResult)
		result = append(result, response.Instances...)

		if response.IsTruncated {
			args.Marker = response.NextMarker
			args.MaxKeys = response.MaxKeys
		} else {
			return result, nil
		}
	}
}

func (s *BccService) GetInstanceDetail(instanceID string) (*api.InstanceModel, error) {
	action := "Get instance detail " + instanceID

	raw, err := s.client.WithBccClient(func(bccClient *bcc.Client) (i interface{}, e error) {
		return bccClient.GetInstanceDetail(instanceID)
	})
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "baiducloud_instance", action, BCESDKGoERROR)
	}

	result, _ := raw.(*api.GetInstanceDetailResult)
	return &result.Instance, nil
}

/**
*	get all disks include  Ephemeral volumes,Cds volumes and System volumes
 */
func (s *BccService) ListAllVolumesWithTypes(instanceId string) ([]api.VolumeModel, []api.VolumeModel, []api.VolumeModel, error) {
	allVolumes, err := s.ListAllVolumes(instanceId)
	if err != nil {
		return nil, nil, nil, err
	}
	// read all disks
	ephVolumes := make([]api.VolumeModel, 0)
	sysVolumes := make([]api.VolumeModel, 0)
	cdsVolumes := make([]api.VolumeModel, 0)
	for _, vol := range allVolumes {
		if vol.Type == api.VolumeTypeEPHEMERAL {
			ephVolumes = append(ephVolumes, vol)
		} else if vol.Type == api.VolumeTypeSYSTEM {
			sysVolumes = append(sysVolumes, vol)
		} else if vol.Type == api.VolumeTypeCDS {
			cdsVolumes = append(cdsVolumes, vol)
		}
	}
	return ephVolumes, cdsVolumes, sysVolumes, nil
}
func (s *BccService) ListAllVolumes(instanceId string) ([]api.VolumeModel, error) {
	args := &api.ListCDSVolumeArgs{
		InstanceId: instanceId,
	}
	action := "List all volumes " + instanceId

	volList := make([]api.VolumeModel, 0)
	for {
		raw, err := s.client.WithBccClient(func(bccClient *bcc.Client) (i interface{}, e error) {
			return bccClient.ListCDSVolume(args)
		})
		if err != nil {
			return nil, WrapErrorf(err, DefaultErrorMsg, "baiducloud_instance", action, BCESDKGoERROR)
		}
		addDebug(action, raw)

		result, _ := raw.(*api.ListCDSVolumeResult)
		volList = append(volList, result.Volumes...)

		if !result.IsTruncated {
			break
		}
		args.Marker = result.NextMarker
	}

	return volList, nil
}

func (s *BccService) ListAllEphemeralVolumes(instanceId string) ([]api.VolumeModel, error) {
	action := "List all ephemeral volumes " + instanceId

	volList, err := s.ListAllVolumes(instanceId)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "baiducloud_instance", action, BCESDKGoERROR)
	}

	ephList := make([]api.VolumeModel, 0)
	for _, vol := range volList {
		if vol.Type != api.VolumeTypeEPHEMERAL {
			continue
		}
		ephList = append(ephList, vol)
	}

	return ephList, nil
}

func (s *BccService) GetSystemVolume(instanceId string) (*api.VolumeModel, error) {
	action := "Get system volume " + instanceId

	volList, err := s.ListAllVolumes(instanceId)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "baiducloud_instance", action, BCESDKGoERROR)
	}

	for _, vol := range volList {
		if vol.Type == api.VolumeTypeSYSTEM {
			return &vol, nil
		}
	}

	return nil, nil
}

func (s *BccService) FlattenInstanceModelToMap(instances []api.InstanceModel) ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(instances))

	for _, inst := range instances {

		// read all disks
		ephVolumes, cdsVolumes, sysVolumes, err := s.ListAllVolumesWithTypes(inst.InstanceId)
		if err != nil {
			return nil, err
		}
		// set system disks
		if len(sysVolumes) == 0 {
			return nil, Error("System volume is missing")
		}
		sysVolume := sysVolumes[0]

		result = append(result, map[string]interface{}{
			"instance_id":              inst.InstanceId,
			"name":                     inst.InstanceName,
			"instance_type":            inst.InstanceType,
			"status":                   inst.Status,
			"description":              inst.Description,
			"payment_timing":           inst.PaymentTiming,
			"create_time":              inst.CreationTime,
			"expire_time":              inst.ExpireTime,
			"internal_ip":              inst.InternalIP,
			"public_ip":                inst.PublicIP,
			"cpu_count":                inst.CpuCount,
			"gpu_card":                 inst.GpuCard,
			"fpga_card":                inst.FpgaCard,
			"card_count":               inst.CardCount,
			"memory_capacity_in_gb":    inst.MemoryCapacityInGB,
			"image_id":                 inst.ImageId,
			"network_capacity_in_mbps": inst.NetworkCapacityInMbps,
			"placement_policy":         inst.PlacementPolicy,
			"zone_name":                inst.ZoneName,
			"subnet_id":                inst.SubnetId,
			"vpc_id":                   inst.VpcId,
			"ephemeral_disks":          s.FlattenCDSVolumeModelToMapForInstance(ephVolumes),
			"cds_disks":                s.FlattenCDSVolumeModelToMapForInstance(cdsVolumes),
			"system_disks":             s.FlattenCDSVolumeModelToMapForInstance(sysVolumes),
			"root_disk_size_in_gb":     sysVolume.DiskSizeInGB,
			"root_disk_storage_type":   sysVolume.StorageType,
			"dedicated_host_id":        inst.DedicatedHostId,
			"auto_renew":               inst.AutoRenew,
			"keypair_id":               inst.KeypairId,
			"keypair_name":             inst.KeypairName,
			"tags":                     flattenTagsToMap(inst.Tags),
		})
	}

	return result, nil
}

func (s *BccService) StartInstance(instanceID string, timeout time.Duration) error {
	action := "Start instance " + instanceID

	_, err := s.client.WithBccClient(func(bccClient *bcc.Client) (i interface{}, e error) {
		return nil, bccClient.StartInstance(instanceID)
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "baiducloud_instance", action, BCESDKGoERROR)
	}

	stateConf := buildStateConf(
		[]string{string(api.InstanceStatusStopped), string(api.InstanceStatusStarting)},
		[]string{string(api.InstanceStatusRunning)},
		timeout,
		s.InstanceStateRefresh(instanceID),
	)
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "baiducloud_instance", action, BCESDKGoERROR)
	}

	return nil
}

func (s *BccService) StopInstance(instanceID string, timeout time.Duration) error {
	action := "Stop instance " + instanceID

	_, err := s.client.WithBccClient(func(bccClient *bcc.Client) (i interface{}, e error) {
		return nil, bccClient.StopInstance(instanceID, false)
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "baiducloud_instance", action, BCESDKGoERROR)
	}

	stateConf := buildStateConf(
		[]string{string(api.InstanceStatusStopping), string(api.InstanceStatusRunning)},
		[]string{string(api.InstanceStatusStopped)},
		timeout,
		s.InstanceStateRefresh(instanceID),
	)
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "baiducloud_instance", action, BCESDKGoERROR)
	}

	return nil
}
