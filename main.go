package main

import (
	"errors"
	"flag"
	"io/ioutil"

	"github.com/cloud-team-poc/mapi-capi-static-converter/pkg/converter"
)

func main() {
	inputFilePath := flag.String("input", "", "input machine file path")
	outputFilePath := flag.String("input", "", "output machine file path")
	conversionApiType := flag.String("api", "", "conversion api type, can be either capi or mapi")
	cloudProviderName := flag.String("provider", "", "cloud provider name, can be aws, azure, gcp, vsphere")

	inputMachine, err := ioutil.ReadFile(*inputFilePath)
	if err != nil {
		panic("can't read machine yaml")
	}

	converter, err := setupConverter(*cloudProviderName, inputMachine)
	if err != nil {
		panic(err)
	}

	convertedFile, err := convertAPI(*conversionApiType, converter)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(*outputFilePath, convertedFile, 0644)
	if err != nil {
		panic(err)
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

func convertAPI(apiType string, converter converter.Converter) ([]byte, error) {
	switch apiType {
	case "capi":
		return converter.ToCAPI()
	case "mapi":
		return converter.ToMAPI()
	default:
		return []byte{}, errors.New("unkown api type")
	}
}
