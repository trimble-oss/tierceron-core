package harbingeropts

import flowcore "github.com/trimble-oss/tierceron-core/v2/flow"

type Option func(*OptionsBuilder)

type OptionsBuilder struct {
	GetIdColumnType    func(table string) any
	GetFolderPrefix    func(custom []string) string
	IsValidProjectName func(projectName string) bool
	BuildTableGrant    func(tableName string) (string, error)
	TableGrantNotify   func(tfmContext flowcore.FlowMachineContext, tableName string)
}

var BuildOptions *OptionsBuilder

func NewOptionsBuilder(opts ...Option) {
	BuildOptions = &OptionsBuilder{}
	for _, opt := range opts {
		opt(BuildOptions)
	}
}
