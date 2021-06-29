package converter

type AWSConverter struct {
	MachineFile []byte
}

func (converter *AWSConverter) ToCAPI() ([]byte, error) {
	return []byte{}, nil
}

func (converter *AWSConverter) ToMAPI() ([]byte, error) {
	return []byte{}, nil
}
