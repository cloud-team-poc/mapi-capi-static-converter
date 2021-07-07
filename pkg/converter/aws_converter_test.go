package converter

import (
	"testing"

	"github.com/cloud-team-poc/mapi-capi-static-converter/pkg/mapi"
	. "github.com/onsi/gomega"
	"k8s.io/utils/pointer"
)

func TestConvertAWSResourceReferenceToCAPI(t *testing.T) {
	g := NewWithT(t)

	mapiAWSResourceRefence := mapi.AWSResourceReference{
		ID:  pointer.StringPtr("testID"),
		ARN: pointer.StringPtr("testARN"),
		Filters: []mapi.Filter{
			{
				Name:   "testName",
				Values: []string{"val"},
			},
		},
	}

	capiAWSResourceRefence := convertAWSResourceReferenceToCAPI(mapiAWSResourceRefence)

	g.Expect(capiAWSResourceRefence.ID).To(Equal(mapiAWSResourceRefence.ID))
	g.Expect(capiAWSResourceRefence.ARN).To(Equal(mapiAWSResourceRefence.ARN))
	g.Expect(len(capiAWSResourceRefence.Filters)).To(Equal(len(mapiAWSResourceRefence.Filters)))
	g.Expect(capiAWSResourceRefence.Filters[0].Name).To(Equal(mapiAWSResourceRefence.Filters[0].Name))
	g.Expect(capiAWSResourceRefence.Filters[0].Values).To(Equal(mapiAWSResourceRefence.Filters[0].Values))
}

func TestConvertAWSFiltersToCAPI(t *testing.T) {
	g := NewWithT(t)

	mapiAWSFilters := []mapi.Filter{
		{
			Name:   "testName1",
			Values: []string{"val1"},
		},
		{
			Name:   "testName2",
			Values: []string{"val2"},
		},
	}

	capiAWSFilters := convertAWSFiltersToCAPI(mapiAWSFilters)

	g.Expect(len(capiAWSFilters)).To(Equal(len(mapiAWSFilters)))
	g.Expect(capiAWSFilters[0].Name).To(Equal(mapiAWSFilters[0].Name))
	g.Expect(capiAWSFilters[0].Values).To(Equal(mapiAWSFilters[0].Values))
	g.Expect(capiAWSFilters[1].Name).To(Equal(mapiAWSFilters[1].Name))
	g.Expect(capiAWSFilters[1].Values).To(Equal(mapiAWSFilters[1].Values))
}

func TestConvertAWSTagsToCAPI(t *testing.T) {
	g := NewWithT(t)

	mapiAWSTags := []mapi.TagSpecification{
		{
			Name:  "name1",
			Value: "value1",
		},
		{
			Name:  "name2",
			Value: "value2",
		},
	}

	capiAWSTags := convertAWSTagsToCAPI(mapiAWSTags)

	g.Expect(len(capiAWSTags)).To(Equal(len(mapiAWSTags)))
	g.Expect(capiAWSTags).To(HaveKeyWithValue(mapiAWSTags[0].Name, mapiAWSTags[0].Value))
	g.Expect(capiAWSTags).To(HaveKeyWithValue(mapiAWSTags[1].Name, mapiAWSTags[1].Value))
}

func TestConvertAWSSecurityGroupstoCAPI(t *testing.T) {
	g := NewWithT(t)

	mapiSGs := []mapi.AWSResourceReference{
		{
			ID:  pointer.StringPtr("testID1"),
			ARN: pointer.StringPtr("testARN1"),
			Filters: []mapi.Filter{
				{
					Name:   "testName1",
					Values: []string{"val1"},
				},
			},
		},
		{
			ID:  pointer.StringPtr("testID2"),
			ARN: pointer.StringPtr("testARN2"),
			Filters: []mapi.Filter{
				{
					Name:   "testName2",
					Values: []string{"val2"},
				},
			},
		},
	}

	capiSGs := convertAWSSecurityGroupstoCAPI(mapiSGs)

	g.Expect(len(capiSGs)).To(Equal(len(mapiSGs)))
	g.Expect(capiSGs[0].ID).To(Equal(mapiSGs[0].ID))
	g.Expect(capiSGs[0].ARN).To(Equal(mapiSGs[0].ARN))
	g.Expect(len(capiSGs[0].Filters)).To(Equal(len(mapiSGs[0].Filters)))
	g.Expect(capiSGs[0].Filters[0].Name).To(Equal(mapiSGs[0].Filters[0].Name))
	g.Expect(capiSGs[0].Filters[0].Values).To(Equal(mapiSGs[0].Filters[0].Values))
	g.Expect(capiSGs[1].ID).To(Equal(mapiSGs[1].ID))
	g.Expect(capiSGs[1].ARN).To(Equal(mapiSGs[1].ARN))
	g.Expect(len(capiSGs[1].Filters)).To(Equal(len(mapiSGs[1].Filters)))
	g.Expect(capiSGs[1].Filters[0].Name).To(Equal(mapiSGs[1].Filters[0].Name))
	g.Expect(capiSGs[1].Filters[0].Values).To(Equal(mapiSGs[1].Filters[0].Values))
}

func TestConvertAWSSpotMarketOptionsToCAPI(t *testing.T) {
	g := NewWithT(t)

	capiSpotMarketOptions := convertAWSSpotMarketOptionsToCAPI(nil)
	g.Expect(capiSpotMarketOptions).To(BeNil())

	mapiSpotMarketOptions := &mapi.SpotMarketOptions{
		MaxPrice: pointer.String("1"),
	}

	capiSpotMarketOptions = convertAWSSpotMarketOptionsToCAPI(mapiSpotMarketOptions)
	g.Expect(capiSpotMarketOptions).ToNot(BeNil())
	g.Expect(capiSpotMarketOptions.MaxPrice).To(Equal(mapiSpotMarketOptions.MaxPrice))
}

func TestConvertAWSBlockDeviceMappingSpecToCAPI(t *testing.T) {
	g := NewWithT(t)

	mapiRootVolume := mapi.BlockDeviceMappingSpec{
		EBS: &mapi.EBSBlockDeviceSpec{
			VolumeSize: pointer.Int64(1),
			VolumeType: pointer.String("type1"),
			Iops:       pointer.Int64(1),
			Encrypted:  pointer.Bool(false),
			KMSKey: mapi.AWSResourceReference{
				ID: pointer.String("test1"),
			},
		},
	}

	mapiNonRootVolume := mapi.BlockDeviceMappingSpec{
		DeviceName: pointer.String("nonrootdevice"),
		EBS: &mapi.EBSBlockDeviceSpec{
			VolumeSize: pointer.Int64(2),
			VolumeType: pointer.String("type2"),
			Iops:       pointer.Int64(2),
			Encrypted:  pointer.Bool(false),
			KMSKey: mapi.AWSResourceReference{
				ID: pointer.String("test2"),
			},
		},
	}

	mapiBlockDeviceMapping := []mapi.BlockDeviceMappingSpec{mapiRootVolume, mapiNonRootVolume}

	capiRootVolume, capiNonRootVolumes := convertAWSBlockDeviceMappingSpecToCAPI(mapiBlockDeviceMapping)

	g.Expect(capiRootVolume).ToNot(BeNil())
	g.Expect(capiRootVolume.DeviceName).To(Equal(""))
	g.Expect(capiRootVolume.Encrypted).To(Equal(*mapiRootVolume.EBS.Encrypted))
	g.Expect(capiRootVolume.EncryptionKey).To(Equal(*mapiRootVolume.EBS.KMSKey.ID))
	g.Expect(capiRootVolume.IOPS).To(Equal(*mapiRootVolume.EBS.Iops))
	g.Expect(capiRootVolume.Size).To(Equal(*mapiRootVolume.EBS.VolumeSize))
	g.Expect(capiRootVolume.Type).To(Equal(*mapiRootVolume.EBS.VolumeType))

	g.Expect(len(capiNonRootVolumes)).To(Equal(1))
	g.Expect(capiNonRootVolumes[0].DeviceName).To(Equal(*mapiNonRootVolume.DeviceName))
	g.Expect(capiNonRootVolumes[0].Encrypted).To(Equal(*mapiNonRootVolume.EBS.Encrypted))
	g.Expect(capiNonRootVolumes[0].EncryptionKey).To(Equal(*mapiNonRootVolume.EBS.KMSKey.ID))
	g.Expect(capiNonRootVolumes[0].IOPS).To(Equal(*mapiNonRootVolume.EBS.Iops))
	g.Expect(capiNonRootVolumes[0].Size).To(Equal(*mapiNonRootVolume.EBS.VolumeSize))
	g.Expect(capiNonRootVolumes[0].Type).To(Equal(*mapiNonRootVolume.EBS.VolumeType))
}

func TestConvertKMSKeyToCAPI(t *testing.T) {
	g := NewWithT(t)

	kmsKey := mapi.AWSResourceReference{
		ID: pointer.String("test1"),
	}

	keyFromID := convertKMSKeyToCAPI(kmsKey)
	g.Expect(*kmsKey.ID).To(Equal(keyFromID))

	kmsKey = mapi.AWSResourceReference{
		ARN: pointer.String("test2"),
	}

	keyFromARN := convertKMSKeyToCAPI(kmsKey)
	g.Expect(*kmsKey.ARN).To(Equal(keyFromARN))
}
