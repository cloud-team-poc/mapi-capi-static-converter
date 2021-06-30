package converter

import (
	"errors"
	"fmt"

	"github.com/cloud-team-poc/mapi-capi-static-converter/pkg/capi"
	"github.com/cloud-team-poc/mapi-capi-static-converter/pkg/mapi"
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

	capiAWSTemplate.Spec.Template.Spec.AMI = converteAWSResourceReferenceToCAPI(mapiProviderConfig.AMI)

	return []byte{}, nil
}

func (converter *AWSConverter) ToMAPI() ([]byte, error) {
	// TODO
	return []byte{}, nil
}

func converteAWSResourceReferenceToCAPI(mapiReference mapi.AWSResourceReference) capi.AWSResourceReference {
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
