package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/cloud-team-poc/mapi-capi-static-converter/pkg/converter"
)

var (
	inputFilePath     string
	conversionApiType string
	cloudProviderName string
)

func init() {
	flag.StringVar(&inputFilePath, "input", "input.yaml", "input machine file path")
	flag.StringVar(&conversionApiType, "api", "", "api type to covert to, can be either capi or mapi")
	flag.StringVar(&cloudProviderName, "provider", "", "cloud provider name, can be aws, azure, gcp, vsphere")
}

func main() {
	flag.Parse()

	fmt.Printf("Converting from %s, for cloud provider: %s\n", conversionApiType, cloudProviderName)

	inputMachine, err := ioutil.ReadFile(inputFilePath)
	if err != nil {
		panic("can't read machine yaml")
	}

	converter, err := setupConverter(cloudProviderName, inputMachine)
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

func setupConverter(cloudProviderName string, inputFile []byte) (converter.Converter, error) {
	switch cloudProviderName {
	case "aws":
		return &converter.AWSConverter{MachineFile: inputFile}, nil
	// case "gcp":
	// case "azure":
	// case "vsphere":
	default:
		return nil, errors.New("unkown cloud provider name")
	}
}
