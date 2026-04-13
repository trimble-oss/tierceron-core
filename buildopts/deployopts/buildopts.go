package deployopts

type Option func(*OptionsBuilder)

type OptionsBuilder struct {
	InitSupportedDeployers func(supportedDeployers []string) []string
	GetDecodedDeployerId   func(sessionId string) (string, bool)
	GetEncodedDeployerId   func(deployment string, env string) (string, bool)
}

var BuildOptions *OptionsBuilder

func NewOptionsBuilder(opts ...Option) {
	BuildOptions = &OptionsBuilder{}
	for _, opt := range opts {
		opt(BuildOptions)
	}
}
