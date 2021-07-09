package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/cloud-team-poc/mapi-capi-static-converter/pkg/converter"
)

var (
	inputMachineSetFilePath      string
	inputMachineTemplateFilePath string
	conversionApiType            string
	cloudProviderName            string
)

func init() {
	flag.StringVar(&inputMachineSetFilePath, "input-machineset", "ms.yaml", "input machine file path")
	flag.StringVar(&inputMachineTemplateFilePath, "input-machine-template", "mtmpl.yaml", "input machine template file path")
	flag.StringVar(&conversionApiType, "api", "", "api type to covert to, can be either capi or mapi")
	flag.StringVar(&cloudProviderName, "provider", "", "cloud provider name, can be aws, azure, gcp, vsphere")
}

func main() {
	flag.Parse()

	fmt.Printf("Converting from %s, for cloud provider: %s\n", conversionApiType, cloudProviderName)

	inputMachineSet, err := ioutil.ReadFile(inputMachineSetFilePath)
	if err != nil {
		panic("can't read machine yaml")
	}

	inputMachineTemplate, err := ioutil.ReadFile(inputMachineTemplateFilePath)
	if err != nil {
		panic("can't read machine yaml")
	}

	converter, err := setupConverter(cloudProviderName, inputMachineSet, inputMachineTemplate)
	if err != nil {
		panic(err)
	}

	convertedTypes, err := converter.ConvertAPI(conversionApiType)
	if err != nil {
		panic(err)
	}

	for i, convertedType := range convertedTypes {
		err = ioutil.WriteFile(fmt.Sprintf("output-%d.yaml", i), convertedType, 0644)
		if err != nil {
			panic(err)
		}
	}
}

func setupConverter(cloudProviderName string, inputMachineSet, inputMachineTemplate []byte) (converter.Converter, error) {
	switch cloudProviderName {
	case "aws":
		return &converter.AWSConverter{
			MachineSetFile:      inputMachineSet,
			MachineTemplateFile: inputMachineTemplate,
		}, nil
	// case "gcp":
	// case "azure":
	// case "vsphere":
	default:
		return nil, errors.New("unkown cloud provider name")
	}
}
