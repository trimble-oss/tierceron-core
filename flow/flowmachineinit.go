package flow

type FlowDefinition struct {
	FlowName         FlowNameType
	FlowTemplatePath string
	FlowSource       string
	FlowType         FlowType // 0 = Flow, 1 = Business, 2 = Test
}

type FlowMachineInitContext struct {
	FlowMachineInterfaceConfigs map[string]interface{}
	GetDatabaseName             func() string
	GetTableFlows               func() []FlowDefinition                     // Required
	GetBusinessFlows            func() []FlowNameType                       // Optional
	GetTestFlows                func() []FlowNameType                       // Optional
	GetTestFlowsByState         func(string) []FlowNameType                 // Optional
	FlowController              func(FlowMachineContext, FlowContext) error // Required
	TestFlowController          func(FlowMachineContext, FlowContext) error // Required
}

/*
GetTableFlows - driverConfigBasis.VersionFilter
GetBusinessFlows - flowopts.BuildOptions.GetAdditionalFlows()
GetTestFlows - testopts.BuildOptions.GetAdditionalTestFlows()
 GetTestFlowsByState - flowopts.BuildOptions.GetAdditionalFlowsByState
*/
