package flowopts

import flowcore "github.com/trimble-oss/tierceron-core/v2/flow"

type Option func(*OptionsBuilder)

type OptionsBuilder struct {
	AllowTrcdbInterfaceOverride func() bool
	GetAdditionalFlows          func() []flowcore.FlowDefinition
	GetAdditionalTestFlows      func() []flowcore.FlowDefinition
	GetAdditionalFlowsByState   func(string) []flowcore.FlowDefinition
	ProcessTestFlowController   func(tfmContext flowcore.FlowMachineContext, tfContext flowcore.FlowContext) error
	ProcessFlowController       func(tfmContext flowcore.FlowMachineContext, tfContext flowcore.FlowContext) error
	GetFlowMachineTemplates     func() map[string]any
}

var BuildOptions *OptionsBuilder

func NewOptionsBuilder(opts ...Option) {
	BuildOptions = &OptionsBuilder{}
	for _, opt := range opts {
		opt(BuildOptions)
	}
}
