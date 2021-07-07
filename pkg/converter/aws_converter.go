package converter

import (
	"errors"
	"fmt"

	"github.com/cloud-team-poc/mapi-capi-static-converter/pkg/capi"
	"github.com/cloud-team-poc/mapi-capi-static-converter/pkg/mapi"
	"github.com/cloud-team-poc/mapi-capi-static-converter/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/yaml"
)

const (
	awsTemplateAPIVersion    = "infrastructure.cluster.x-k8s.io/v1alpha4"
	awsTemplateKind          = "AWSMachineTemplate"
	capiMachineSetAPIVersion = "cluster.x-k8s.io"
	capiMachineSetKind       = "MachineSet"
	workerUserDataSecretName = "worker-user-data"
)

type AWSConverter struct {
	MachineFile []byte
}

func (converter *AWSConverter) ConvertAPI(apiType string) ([][]byte, error) {
	switch apiType {
	case "capi":
		return converter.ToCAPI()
	case "mapi":
		return converter.ToMAPI()
	default:
		return nil, errors.New("unkown api type")
	}
}

func (converter *AWSConverter) ToCAPI() ([][]byte, error) {
	machineSet := &mapi.MachineSet{}
	if err := yaml.Unmarshal(converter.MachineFile, machineSet); err != nil {
		return nil, fmt.Errorf("error unmarshalling machine: %v", err)
	}

	mapiProviderConfig, err := mapi.ProviderSpecFromRawExtension(machineSet.Spec.Template.Spec.ProviderSpec.Value)
	if err != nil {
		return nil, err
	}

	capiAWSTemplate := &capi.AWSMachineTemplate{}
	capiAWSTemplate.ObjectMeta = metav1.ObjectMeta{
		Name:      machineSet.Name,
		Namespace: machineSet.Namespace,
	}
	capiAWSTemplate.TypeMeta = metav1.TypeMeta{
		Kind:       awsTemplateKind,
		APIVersion: awsTemplateAPIVersion,
	}
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

	capiMachineSet := &capi.MachineSet{}
	capiMachineSet.ObjectMeta = metav1.ObjectMeta{
		Name:      machineSet.Name,
		Namespace: machineSet.Namespace,
	}
	capiMachineSet.TypeMeta = metav1.TypeMeta{
		Kind:       capiMachineSetKind,
		APIVersion: capiMachineSetAPIVersion,
	}
	capiMachineSet.Spec.Selector = machineSet.Spec.Selector
	capiMachineSet.Spec.Template.Labels = machineSet.Spec.Template.Labels
	capiMachineSet.Spec.ClusterName = "" // TODO: this should be fetched from infra object
	capiMachineSet.Spec.Replicas = machineSet.Spec.Replicas
	capiMachineSet.Spec.Template.Spec.Bootstrap = capi.Bootstrap{
		DataSecretName: pointer.String(workerUserDataSecretName),
	}
	capiMachineSet.Spec.Template.Spec.ClusterName = "" // TODO: this should be fetched from infra object
	capiMachineSet.Spec.Template.Spec.InfrastructureRef = corev1.ObjectReference{
		APIVersion: awsTemplateAPIVersion,
		Kind:       awsTemplateKind,
		Name:       machineSet.Name,
	}

	yamlCAPIAWSTemplate, err := yaml.Marshal(capiAWSTemplate)
	if err != nil {
		return nil, err
	}

	yamlCAPIMachineSet, err := yaml.Marshal(capiMachineSet)
	if err != nil {
		return nil, err
	}

	return [][]byte{yamlCAPIAWSTemplate, yamlCAPIMachineSet}, nil
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
		return *kmsKey.ARN
	}

	return ""
}

func (converter *AWSConverter) ToMAPI() ([][]byte, error) {
	// TODO
	return [][]byte{}, nil
}
