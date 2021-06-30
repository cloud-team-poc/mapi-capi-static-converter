package converter

type Converter interface {
	ToMAPI() ([]byte, error)
	ToCAPI() ([]byte, error)
	ConvertAPI(apiType string) ([]byte, error)
}
