package converter

import (
	"errors"
	"fmt"

	"github.com/cloud-team-poc/mapi-capi-static-converter/pkg/capi"
	"github.com/cloud-team-poc/mapi-capi-static-converter/pkg/mapi"
	"github.com/cloud-team-poc/mapi-capi-static-converter/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/yaml"
)

const (
	awsTemplateAPIVersion    = "infrastructure.cluster.x-k8s.io/v1alpha4"
	awsTemplateKind          = "AWSMachineTemplate"
	capiMachineSetAPIVersion = "cluster.x-k8s.io"
	capiMachineSetKind       = "MachineSet"
	mapiMachineSetKind       = "machine.openshift.io"
	mapiMachineSetAPIVersion = "MachineSet"
	workerUserDataSecretName = "worker-user-data"
)

type AWSConverter struct {
	MachineSetFile      []byte
	MachineTemplateFile []byte
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
	if err := yaml.Unmarshal(converter.MachineSetFile, machineSet); err != nil {
		return nil, fmt.Errorf("error unmarshalling machineset: %v", err)
	}

	mapiProviderConfig, err := mapi.ProviderSpecFromRawExtension(machineSet.Spec.Template.Spec.ProviderSpec.Value)
	if err != nil {
		return nil, err
	}

	capiAWSTemplate := convertProviderConfigToAWSMachineTemplate(machineSet.Name, machineSet.Namespace, mapiProviderConfig)

	capiMachineSet := convertMachineSetToCAPI(machineSet)

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

func convertProviderConfigToAWSMachineTemplate(name, namespace string, mapiProviderConfig *mapi.AWSMachineProviderConfig) *capi.AWSMachineTemplate {
	capiAWSTemplate := &capi.AWSMachineTemplate{}
	capiAWSTemplate.ObjectMeta = metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
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
	capiAWSTemplate.Spec.Template.Spec.CloudInit = capi.CloudInit{
		InsecureSkipSecretsManager: false,
		SecureSecretsBackend:       capi.SecretBackendSecretsManager,
	}

	return capiAWSTemplate
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

func convertMachineSetToCAPI(mapiMachineSet *mapi.MachineSet) *capi.MachineSet {
	capiMachineSet := &capi.MachineSet{}
	capiMachineSet.ObjectMeta = metav1.ObjectMeta{
		Name:      mapiMachineSet.Name,
		Namespace: mapiMachineSet.Namespace,
	}
	capiMachineSet.TypeMeta = metav1.TypeMeta{
		Kind:       capiMachineSetKind,
		APIVersion: capiMachineSetAPIVersion,
	}
	capiMachineSet.Spec.Selector = mapiMachineSet.Spec.Selector
	capiMachineSet.Spec.Template.Labels = mapiMachineSet.Spec.Template.Labels
	capiMachineSet.Spec.ClusterName = "" // TODO: this should be fetched from infra object
	capiMachineSet.Spec.Replicas = mapiMachineSet.Spec.Replicas
	capiMachineSet.Spec.Template.Spec.Bootstrap = capi.Bootstrap{
		DataSecretName: pointer.String(workerUserDataSecretName),
	}
	capiMachineSet.Spec.Template.Spec.ClusterName = "" // TODO: this should be fetched from infra object
	capiMachineSet.Spec.Template.Spec.InfrastructureRef = corev1.ObjectReference{
		APIVersion: awsTemplateAPIVersion,
		Kind:       awsTemplateKind,
		Name:       mapiMachineSet.Name,
	}

	return capiMachineSet
}

func (converter *AWSConverter) ToMAPI() ([][]byte, error) {
	machineSet := &capi.MachineSet{}
	if err := yaml.Unmarshal(converter.MachineSetFile, machineSet); err != nil {
		return nil, fmt.Errorf("error unmarshalling machineset: %v", err)
	}

	machineTemplate := &capi.AWSMachineTemplate{}
	if err := yaml.Unmarshal(converter.MachineTemplateFile, machineTemplate); err != nil {
		return nil, fmt.Errorf("error unmarshalling machine template: %v", err)
	}

	mapiProviderConfig := convertAWSMachineTemplateToroviderConfig(machineTemplate)

	rawProviderConfig, err := mapi.RawExtensionFromProviderSpec(mapiProviderConfig)
	if err != nil {
		return nil, err
	}

	mapiMachineSet := convertMachineSetToMAPI(machineSet, rawProviderConfig)

	yamlMAPIMachineSet, err := yaml.Marshal(mapiMachineSet)
	if err != nil {
		return nil, err
	}

	return [][]byte{yamlMAPIMachineSet}, nil
}

func convertAWSMachineTemplateToroviderConfig(awsMachineTemplate *capi.AWSMachineTemplate) *mapi.AWSMachineProviderConfig {
	mapiProviderConfig := &mapi.AWSMachineProviderConfig{}

	mapiProviderConfig.AMI = convertAWSResourceReferenceToMAPI(awsMachineTemplate.Spec.Template.Spec.AMI)
	mapiProviderConfig.InstanceType = awsMachineTemplate.Spec.Template.Spec.InstanceType
	mapiProviderConfig.Tags = convertAWSTagsToMAPI(awsMachineTemplate.Spec.Template.Spec.AdditionalTags)
	mapiProviderConfig.IAMInstanceProfile = &mapi.AWSResourceReference{
		ID: &awsMachineTemplate.Spec.Template.Spec.IAMInstanceProfile,
	}
	mapiProviderConfig.KeyName = awsMachineTemplate.Spec.Template.Spec.SSHKeyName
	mapiProviderConfig.PublicIP = awsMachineTemplate.Spec.Template.Spec.PublicIP
	mapiProviderConfig.Placement = mapi.Placement{
		AvailabilityZone: util.DerefString(awsMachineTemplate.Spec.Template.Spec.FailureDomain),
		Tenancy:          convertAWSTenancyToMAPI(awsMachineTemplate.Spec.Template.Spec.Tenancy),
		Region:           "", // TODO: fetch region from cluster object
	}
	mapiProviderConfig.SecurityGroups = convertAWSSecurityGroupstoMAPI(awsMachineTemplate.Spec.Template.Spec.AdditionalSecurityGroups)
	mapiProviderConfig.Subnet = convertAWSResourceReferenceToMAPI(*awsMachineTemplate.Spec.Template.Spec.Subnet)
	mapiProviderConfig.SpotMarketOptions = convertAWSSpotMarketOptionsToMAPI(awsMachineTemplate.Spec.Template.Spec.SpotMarketOptions)
	mapiProviderConfig.BlockDevices = convertAWSBlockDeviceMappingSpecToMAPI(awsMachineTemplate.Spec.Template.Spec.RootVolume, awsMachineTemplate.Spec.Template.Spec.NonRootVolumes)
	return mapiProviderConfig
}

func convertAWSResourceReferenceToMAPI(mapiReference capi.AWSResourceReference) mapi.AWSResourceReference {
	return mapi.AWSResourceReference{
		ID:      mapiReference.ID,
		ARN:     mapiReference.ARN,
		Filters: convertAWSFiltersToMAPI(mapiReference.Filters),
	}
}

func convertAWSFiltersToMAPI(capiFilters []capi.Filter) []mapi.Filter {
	mapiFilters := []mapi.Filter{}
	for _, filter := range capiFilters {
		mapiFilters = append(mapiFilters, mapi.Filter{
			Name:   filter.Name,
			Values: filter.Values,
		})
	}
	return mapiFilters
}

func convertAWSTagsToMAPI(capiTags capi.Tags) []mapi.TagSpecification {
	mapiTags := []mapi.TagSpecification{}
	for key, value := range capiTags {
		mapiTags = append(mapiTags, mapi.TagSpecification{
			Name:  key,
			Value: value,
		})
	}
	return mapiTags
}

func convertAWSTenancyToMAPI(capiTenancy string) mapi.InstanceTenancy {
	switch capiTenancy {
	case "default":
		return mapi.DefaultTenancy
	case "dedicated":
		return mapi.DedicatedTenancy
	default:
		return mapi.HostTenancy
	}
}

func convertAWSSecurityGroupstoMAPI(sgs []capi.AWSResourceReference) []mapi.AWSResourceReference {
	mapiSGs := []mapi.AWSResourceReference{}
	for _, sg := range sgs {
		mapiSGs = append(mapiSGs, convertAWSResourceReferenceToMAPI(sg))
	}
	return mapiSGs
}

func convertAWSSpotMarketOptionsToMAPI(capiSpotMarketOptions *capi.SpotMarketOptions) *mapi.SpotMarketOptions {
	if capiSpotMarketOptions == nil {
		return nil
	}
	return &mapi.SpotMarketOptions{
		MaxPrice: capiSpotMarketOptions.MaxPrice,
	}
}

func convertAWSBlockDeviceMappingSpecToMAPI(rootVolume *capi.Volume, nonRootVolumes []capi.Volume) []mapi.BlockDeviceMappingSpec {
	blockDeviceMapping := []mapi.BlockDeviceMappingSpec{}

	blockDeviceMapping = append(blockDeviceMapping, mapi.BlockDeviceMappingSpec{
		EBS: &mapi.EBSBlockDeviceSpec{
			VolumeSize: &rootVolume.Size,
			VolumeType: &rootVolume.Type,
			Iops:       &rootVolume.IOPS,
			Encrypted:  &rootVolume.Encrypted,
			KMSKey:     convertKMSKeyToMAPI(rootVolume.EncryptionKey),
		},
	})

	for _, volume := range nonRootVolumes {
		blockDeviceMapping = append(blockDeviceMapping, mapi.BlockDeviceMappingSpec{
			DeviceName: &volume.DeviceName,
			EBS: &mapi.EBSBlockDeviceSpec{
				VolumeSize: &volume.Size,
				VolumeType: &volume.Type,
				Iops:       &volume.IOPS,
				Encrypted:  &volume.Encrypted,
				KMSKey:     convertKMSKeyToMAPI(volume.EncryptionKey),
			},
		})
	}

	return blockDeviceMapping
}

// TODO: fix this conversion
// It’s not possible to convert KMSKey back to MAPI, because upstream uses string type to represent ID or ARN.
// We are using AWSResourceReference, this means that during conversion we can’t know where to map string to ID or ARN.
func convertKMSKeyToMAPI(kmsKey string) mapi.AWSResourceReference {
	return mapi.AWSResourceReference{
		ID: &kmsKey,
	}
}

func convertMachineSetToMAPI(capiMachineSet *capi.MachineSet, rawProviderConfig *runtime.RawExtension) *mapi.MachineSet {
	mapiMachineSet := &mapi.MachineSet{}
	mapiMachineSet.ObjectMeta = metav1.ObjectMeta{
		Name:      capiMachineSet.Name,
		Namespace: capiMachineSet.Namespace,
	}
	mapiMachineSet.TypeMeta = metav1.TypeMeta{
		Kind:       mapiMachineSetKind,
		APIVersion: mapiMachineSetAPIVersion,
	}
	mapiMachineSet.Spec.Selector = capiMachineSet.Spec.Selector
	mapiMachineSet.Spec.Template.Labels = capiMachineSet.Spec.Template.Labels
	mapiMachineSet.Spec.Replicas = capiMachineSet.Spec.Replicas
	mapiMachineSet.Spec.Template.Spec.ProviderSpec = mapi.ProviderSpec{
		Value: rawProviderConfig,
	}

	return mapiMachineSet
}
