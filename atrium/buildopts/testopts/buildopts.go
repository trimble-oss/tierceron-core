package testopts

import flowcore "github.com/trimble-oss/tierceron-core/v2/flow"

type Option func(*OptionsBuilder)

type OptionsBuilder struct {
	GetAdditionalTestFlows    func() []flowcore.FlowDefinition
	GetAdditionalFlowsByState func(teststate string) []flowcore.FlowDefinition
	ProcessTestFlowController func(tfmContext flowcore.FlowMachineContext, tfContext flowcore.FlowContext) error
	GetTestConfig             func(tokenPtr *string, wantPluginPaths bool) map[string]any
}

var BuildOptions *OptionsBuilder

func NewOptionsBuilder(opts ...Option) {
	BuildOptions = &OptionsBuilder{}
	for _, opt := range opts {
		opt(BuildOptions)
	}
}
