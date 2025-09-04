package flow

import (
	"github.com/trimble-oss/tierceron-core/v2/core"
)

type FlowDefinition struct {
	FlowHeader       FlowHeaderType
	FlowTemplatePath string
	FlowType         FlowType // 0 = Flow, 1 = Business, 2 = Test
}

type FlowMachineInitContext struct {
	GetFlowMachineTemplates     func() map[string]any
	FlowMachineInterfaceConfigs map[string]any
	GetDatabaseName             func(FlumeDbType) string
	IsSupportedFlow             func(string) bool                           // Required
	GetTableFlows               func() []FlowDefinition                     // Required
	GetBusinessFlows            func() []FlowDefinition                     // Optional
	GetTestFlows                func() []FlowDefinition                     // Optional
	GetTestFlowsByState         func(string) []FlowDefinition               // Optional
	FlowController              func(FlowMachineContext, FlowContext) error // Required
	TestFlowController          func(FlowMachineContext, FlowContext) error // Required
	DfsChan                     *chan *core.TTDINode                        // Channel for sending data flow statistics
	FlowChatMsgSenderChan       *chan *core.ChatMsg                         // Channel for sending chat messages from flow to plugin
}

var HARBINGER_INTERFACE_CONFIG = "./config.yml"

func (fmic FlowMachineInitContext) GetFilteredBusinessFlows(kernelId string) []FlowDefinition {
	hasRestrictedFlow := false
	for _, flow := range fmic.GetBusinessFlows() {
		if flow.FlowHeader.GetInstances() != "*" {
			hasRestrictedFlow = true
			break
		}
	}
	if !hasRestrictedFlow {
		return fmic.GetBusinessFlows()
	} else {
		var filteredFlows []FlowDefinition
		for _, flow := range fmic.GetBusinessFlows() {
			if flow.FlowHeader.GetInstances() == kernelId || flow.FlowHeader.GetInstances() == "*" {
				filteredFlows = append(filteredFlows, flow)
			}
		}
		return filteredFlows
	}
}

func (fmic FlowMachineInitContext) GetFilteredBusinessFlowNames(kernelId string) []FlowNameType {
	var filteredFlowNames []FlowNameType
	for _, flow := range fmic.GetFilteredBusinessFlows(kernelId) {
		filteredFlowNames = append(filteredFlowNames, flow.FlowHeader.FlowNameType())
	}
	return filteredFlowNames
}

func (fmic FlowMachineInitContext) GetFilteredTableFlowDefinitions(kernelId string) []FlowDefinition {
	hasRestrictedFlow := false
	for _, flow := range fmic.GetTableFlows() {
		if flow.FlowHeader.GetInstances() != "*" {
			hasRestrictedFlow = true
			break
		}
	}
	if !hasRestrictedFlow {
		return fmic.GetTableFlows()
	} else {
		var filteredFlows []FlowDefinition
		for _, flow := range fmic.GetTableFlows() {
			if flow.FlowHeader.GetInstances() == kernelId || flow.FlowHeader.GetInstances() == "*" {
				filteredFlows = append(filteredFlows, flow)
			}
		}
		return filteredFlows
	}
}

func (fmic FlowMachineInitContext) GetFilteredTableFlows(kernelId string) []FlowDefinition {
	var filteredFlowDefinitions []FlowDefinition
	for _, flow := range fmic.GetFilteredTableFlowDefinitions(kernelId) {
		filteredFlowDefinitions = append(filteredFlowDefinitions, flow)
	}
	return filteredFlowDefinitions
}

func (fmic FlowMachineInitContext) GetFilteredTableFlowNames(kernelId string) []string {
	var filteredFlowNames []string
	for _, flow := range fmic.GetFilteredTableFlowDefinitions(kernelId) {
		filteredFlowNames = append(filteredFlowNames, flow.FlowHeader.FlowName())
	}
	return filteredFlowNames
}

func (fmic FlowMachineInitContext) GetFilteredTestFlowNames(kernelId string) []FlowNameType {
	var filteredFlowNames []FlowNameType
	for _, flow := range fmic.GetTestFlows() {
		filteredFlowNames = append(filteredFlowNames, flow.FlowHeader.FlowNameType())
	}
	return filteredFlowNames
}

/*
GetTableFlows - driverConfigBasis.VersionFilter
GetBusinessFlows - flowopts.BuildOptions.GetAdditionalFlows()
GetTestFlows - testopts.BuildOptions.GetAdditionalTestFlows()
 GetTestFlowsByState - flowopts.BuildOptions.GetAdditionalFlowsByState
*/
