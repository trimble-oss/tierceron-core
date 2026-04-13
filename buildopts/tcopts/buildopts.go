package tcopts

type Option func(*OptionsBuilder)

type OptionsBuilder struct {
	CheckIncomingColumnName      func(col string) bool
	CheckFlowDataIncoming        func(secretColumns map[string]string, secretValue string, dbName string, tableName string) ([]byte, string, string, string, error)
	CheckIncomingAliasColumnName func(col string) bool
	GetTrcDbUrl                  func(data map[string]any) string
}

var BuildOptions *OptionsBuilder

func NewOptionsBuilder(opts ...Option) {
	BuildOptions = &OptionsBuilder{}
	for _, opt := range opts {
		opt(BuildOptions)
	}
}
