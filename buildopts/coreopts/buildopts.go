package coreopts

import (
	"database/sql"

	flowcore "github.com/trimble-oss/tierceron-core/v2/flow"
)

type Option func(*OptionsBuilder)

type OptionsBuilder struct {
	GetFolderPrefix             func(custom []string) string
	GetSupportedTemplates       func(custom []string) []string
	GetVaultInstallRoot         func() string
	IsLocalEndpoint             func(addr string) bool
	GetSupportedDomains         func(bool) []string
	GetSupportedEndpoints       func(bool) [][]string
	GetLocalHost                func() string
	GetRegionByHost             func(hostName string) string
	GetDefaultRegion            func() string
	GetVaultHost                func() string
	GetVaultHostPort            func() string
	GetUserNameField            func() string
	GetUserCodeField            func() string
	ActiveSessions              func(db *sql.DB) ([]map[string]any, error)
	GetSyncedTables             func() []string
	IsSupportedFlow             func(flowName string) bool
	GetDatabaseName             func(flumeDbType flowcore.FlumeDbType) string
	FindIndexForService         func(tfmContext flowcore.FlowMachineContext, project string, service string) (string, []string, string, error)
	DecryptSecretConfig         func(map[string]any, map[string]any) (string, error)
	GetDFSPathName              func() (string, string)
	CompareLastModified         func(dfStatMapA map[string]any, dfStatMapB map[string]any) bool
	PreviousStateCheck          func(currentState int) int
	GetMachineID                func() string
	InitPluginConfig            func(pluginEnvConfig map[string]any) map[string]any
	GetPluginRestrictedMappings func() map[string][][]string
	GetSupportedCertIssuers     func() []string
}

var BuildOptions *OptionsBuilder

func NewOptionsBuilder(opts ...Option) {
	BuildOptions = &OptionsBuilder{}
	for _, opt := range opts {
		opt(BuildOptions)
	}
}
