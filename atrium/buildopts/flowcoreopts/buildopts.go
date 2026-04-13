package flowcoreopts

const (
	DataflowTestNameColumn      = "flowName"
	DataflowGroupTestNameColumn = "flowGroup"
	DataflowTestIdColumn        = "argosId"
	DataflowTestStateCodeColumn = "stateCode"
)

type Option func(*OptionsBuilder)

type OptionsBuilder struct {
	GetIdColumnType func(table string) any
}

func LoadOptions() Option {
	return func(optionsBuilder *OptionsBuilder) {
		optionsBuilder.GetIdColumnType = GetIdColumnType
	}
}

var BuildOptions *OptionsBuilder

func NewOptionsBuilder(opts ...Option) {
	BuildOptions = &OptionsBuilder{}
	for _, opt := range opts {
		opt(BuildOptions)
	}
}

func GetIdColumnType(table string) any {
	return nil
}

func IsCreateTableEnabled() bool {
	return false
}
