package coreopts

type Option func(*OptionsBuilder)

type OptionsBuilder struct {
	IsLocalEndpoint func(addr string) bool
}

func LoadOptions() Option {
	return func(optionsBuilder *OptionsBuilder) {
		optionsBuilder.IsLocalEndpoint = IsLocalEndpoint
	}
}

var BuildOptions *OptionsBuilder

func NewOptionsBuilder(opts ...Option) {
	BuildOptions = &OptionsBuilder{}
	for _, opt := range opts {
		opt(BuildOptions)
	}
}
