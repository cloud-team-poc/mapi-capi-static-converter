package converter

import (
	"errors"
	"fmt"

	"github.com/cloud-team-poc/mapi-capi-static-converter/pkg/capi"
	"github.com/cloud-team-poc/mapi-capi-static-converter/pkg/mapi"
	"github.com/cloud-team-poc/mapi-capi-static-converter/pkg/util"
	"sigs.k8s.io/yaml"
)

type AWSConverter struct {
	MachineFile []byte
}

func (converter *AWSConverter) ConvertAPI(apiType string) ([]byte, error) {
	switch apiType {
	case "capi":
		return converter.ToCAPI()
	case "mapi":
		return converter.ToMAPI()
	default:
		return nil, errors.New("unkown api type")
	}
}

func (converter *AWSConverter) ToCAPI() ([]byte, error) {
	machine := &mapi.Machine{}
	if err := yaml.Unmarshal(converter.MachineFile, machine); err != nil {
		return nil, fmt.Errorf("error unmarshalling machine: %v", err)
	}

	mapiProviderConfig, err := mapi.ProviderSpecFromRawExtension(machine.Spec.ProviderSpec.Value)
	if err != nil {
		return nil, err
	}

	capiAWSTemplate := &capi.AWSMachineTemplate{}
	capiAWSTemplate.Spec.Template.Spec.AMI = convertAWSResourceReferenceToCAPI(mapiProviderConfig.AMI)
	capiAWSTemplate.Spec.Template.Spec.InstanceType = mapiProviderConfig.InstanceType
	capiAWSTemplate.Spec.Template.Spec.AdditionalTags = convertAWSTagsToCAPI(mapiProviderConfig.Tags)
	capiAWSTemplate.Spec.Template.Spec.IAMInstanceProfile = util.DerefString(mapiProviderConfig.IAMInstanceProfile.ID)
	capiAWSTemplate.Spec.Template.Spec.SSHKeyName = mapiProviderConfig.KeyName
	capiAWSTemplate.Spec.Template.Spec.PublicIP = mapiProviderConfig.PublicIP
	capiAWSTemplate.Spec.Template.Spec.FailureDomain = &mapiProviderConfig.Placement.AvailabilityZone
	capiAWSTemplate.Spec.Template.Spec.Tenancy = string(mapiProviderConfig.Placement.Tenancy)
	capiAWSTemplate.Spec.Template.Spec.AdditionalSecurityGroups = convertAWSSecurityGroupstoCAPI(mapiProviderConfig.SecurityGroups)
	capiSubnet := convertAWSResourceReferenceToCAPI(mapiProviderConfig.Subnet)
	capiAWSTemplate.Spec.Template.Spec.Subnet = &capiSubnet
	capiAWSTemplate.Spec.Template.Spec.SpotMarketOptions = convertAWSSpotMarketOptionsToCAPI(mapiProviderConfig.SpotMarketOptions)
	rootVolume, nonRootVolumes := convertAWSBlockDeviceMappingSpecToCAPI(mapiProviderConfig.BlockDevices)
	capiAWSTemplate.Spec.Template.Spec.RootVolume = rootVolume
	capiAWSTemplate.Spec.Template.Spec.NonRootVolumes = nonRootVolumes

	return yaml.Marshal(capiAWSTemplate)
}

func (converter *AWSConverter) ToMAPI() ([]byte, error) {
	// TODO
	return []byte{}, nil
}

func convertAWSResourceReferenceToCAPI(mapiReference mapi.AWSResourceReference) capi.AWSResourceReference {
	return capi.AWSResourceReference{
		ID:      mapiReference.ID,
		ARN:     mapiReference.ARN,
		Filters: convertAWSFiltersToCAPI(mapiReference.Filters),
	}
}

func convertAWSFiltersToCAPI(mapiFilters []mapi.Filter) []capi.Filter {
	capiFilters := []capi.Filter{}
	for _, filter := range mapiFilters {
		capiFilters = append(capiFilters, capi.Filter{
			Name:   filter.Name,
			Values: filter.Values,
		})
	}
	return capiFilters
}

func convertAWSTagsToCAPI(mapiTags []mapi.TagSpecification) capi.Tags {
	capiTags := map[string]string{}
	for _, tag := range mapiTags {
		capiTags[tag.Name] = tag.Value
	}
	return capiTags
}

func convertAWSSecurityGroupstoCAPI(sgs []mapi.AWSResourceReference) []capi.AWSResourceReference {
	capiSGs := []capi.AWSResourceReference{}
	for _, sg := range sgs {
		capiSGs = append(capiSGs, convertAWSResourceReferenceToCAPI(sg))
	}
	return capiSGs
}

func convertAWSSpotMarketOptionsToCAPI(mapiSpotMarketOptions *mapi.SpotMarketOptions) *capi.SpotMarketOptions {
	if mapiSpotMarketOptions == nil {
		return nil
	}
	return &capi.SpotMarketOptions{
		MaxPrice: mapiSpotMarketOptions.MaxPrice,
	}
}

func convertAWSBlockDeviceMappingSpecToCAPI(mapiBlockDeviceMapping []mapi.BlockDeviceMappingSpec) (*capi.Volume, []capi.Volume) {
	rootVolume := &capi.Volume{}
	nonRootVolumes := []capi.Volume{}

	for _, mapping := range mapiBlockDeviceMapping {
		if mapping.DeviceName == nil {
			rootVolume = &capi.Volume{
				Size:          *mapping.EBS.VolumeSize,
				Type:          *mapping.EBS.VolumeType,
				IOPS:          *mapping.EBS.Iops,
				Encrypted:     *mapping.EBS.Encrypted,
				EncryptionKey: convertKMSKeyToCAPI(mapping.EBS.KMSKey),
			}
			continue
		}
		nonRootVolumes = append(nonRootVolumes, capi.Volume{
			DeviceName:    *mapping.DeviceName,
			Size:          *mapping.EBS.VolumeSize,
			Type:          *mapping.EBS.VolumeType,
			IOPS:          *mapping.EBS.Iops,
			Encrypted:     *mapping.EBS.Encrypted,
			EncryptionKey: convertKMSKeyToCAPI(mapping.EBS.KMSKey),
		})
	}

	return rootVolume, nonRootVolumes
}

func convertKMSKeyToCAPI(kmsKey mapi.AWSResourceReference) string {
	if kmsKey.ID != nil {
		return *kmsKey.ID
	}

	if kmsKey.ARN != nil {
		return *kmsKey.ID
	}

	return ""
}
