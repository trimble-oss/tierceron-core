package flow

type FlowDefinition struct {
	FlowName         FlowNameType
	FlowTemplatePath string
	FlowSource       string
	FlowType         FlowType // 0 = Flow, 1 = Business, 2 = Test
}

type FlowMachineInitContext struct {
	GetFlowMachineTemplates     func() map[string]any
	FlowMachineInterfaceConfigs map[string]any
	GetDatabaseName             func() string
	GetTableFlows               func() []FlowDefinition                     // Required
	GetBusinessFlows            func() []FlowNameType                       // Optional
	GetTestFlows                func() []FlowNameType                       // Optional
	GetTestFlowsByState         func(string) []FlowNameType                 // Optional
	FlowController              func(FlowMachineContext, FlowContext) error // Required
	TestFlowController          func(FlowMachineContext, FlowContext) error // Required
}

var HARBINGER_INTERFACE_CONFIG = "./config.yml"

func (fmic FlowMachineInitContext) GetFiltererBusinessFlows(kernelId string) []FlowNameType {
	hasRestrictedFlow := false
	for _, flow := range fmic.GetBusinessFlows() {
		if flow.Instances != "*" {
			hasRestrictedFlow = true
			break
		}
	}
	if !hasRestrictedFlow {
		return fmic.GetBusinessFlows()
	} else {
		var filteredFlows []FlowNameType
		for _, flow := range fmic.GetBusinessFlows() {
			if flow.Instances == kernelId || flow.Instances == "*" {
				filteredFlows = append(filteredFlows, flow)
			}
		}
		return filteredFlows
	}
}

func (fmic FlowMachineInitContext) GetFiltererTableFlowDefinitions(kernelId string) []FlowDefinition {
	hasRestrictedFlow := false
	for _, flow := range fmic.GetTableFlows() {
		if flow.FlowName.Instances != "*" {
			hasRestrictedFlow = true
			break
		}
	}
	if !hasRestrictedFlow {
		return fmic.GetTableFlows()
	} else {
		var filteredFlows []FlowDefinition
		for _, flow := range fmic.GetTableFlows() {
			if flow.FlowName.Instances == kernelId || flow.FlowName.Instances == "*" {
				filteredFlows = append(filteredFlows, flow)
			}
		}
		return filteredFlows
	}
}

func (fmic FlowMachineInitContext) GetFiltererTableFlows(kernelId string) []FlowNameType {
	var filteredFlowNames []FlowNameType
	for _, flow := range fmic.GetFiltererTableFlowDefinitions(kernelId) {
		filteredFlowNames = append(filteredFlowNames, flow.FlowName)
	}
	return filteredFlowNames
}

func (fmic FlowMachineInitContext) GetFiltererTableFlowNames(kernelId string) []string {
	var filteredFlowNames []string
	for _, flow := range fmic.GetFiltererTableFlowDefinitions(kernelId) {
		filteredFlowNames = append(filteredFlowNames, flow.FlowName.FlowName())
	}
	return filteredFlowNames
}

/*
GetTableFlows - driverConfigBasis.VersionFilter
GetBusinessFlows - flowopts.BuildOptions.GetAdditionalFlows()
GetTestFlows - testopts.BuildOptions.GetAdditionalTestFlows()
 GetTestFlowsByState - flowopts.BuildOptions.GetAdditionalFlowsByState
*/
