package buildopts

type Option func(*OptionsBuilder)

type OptionsBuilder struct {
	SetLogger      func(logger interface{})
	SetErrorLogger func(logger interface{})

	CheckMemLock func(bucket string, key string) bool
}

func LoadOptions() Option {
	return func(optionsBuilder *OptionsBuilder) {
		optionsBuilder.SetLogger = SetLogger
		optionsBuilder.SetErrorLogger = SetErrorLogger
		optionsBuilder.CheckMemLock = CheckMemLock
	}
}

var BuildOptions *OptionsBuilder

func NewOptionsBuilder(opts ...Option) {
	BuildOptions = &OptionsBuilder{}
	for _, opt := range opts {
		opt(BuildOptions)
	}
}
