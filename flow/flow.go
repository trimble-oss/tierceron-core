package flow

import (
	"io"
	"log"
	"sync"

	tccore "github.com/trimble-oss/tierceron-core/v2/core"
)

type FlowType int64
type FlowNameType string

func (fnt FlowNameType) TableName() string {
	return string(fnt)
}

func (fnt FlowNameType) ServiceName() string {
	return string(fnt)
}

func (fnt FlowNameType) FlowName() string {
	return string(fnt)
}

type PermissionUpdate interface{}
type FlowStateUpdate interface {
	// NewFlowState() CurrentFlowState
}
type CurrentFlowState interface{}
type TemplateData interface{}

type FlowDefinitionContext struct {
	GetTableConfigurationById   func(databaseName string, tableName string, idColumnName string) map[string]interface{}
	GetTableConfigurations      func(db interface{}, secLookup bool) ([]interface{}, error)
	GetTableMap                 func(tableConfig interface{}) map[string]interface{}
	GetTableFromMap             func(tableConfigMap map[string]interface{}) interface{}
	GetFilterFieldFromConfig    func(tableconfig interface{}) string
	GetTableMapFromArray        func(tableArray []interface{}) map[string]interface{}
	GetTableConfigurationInsert func(tableConfigMap map[string]interface{}, databaseName string, tableName string) map[string]interface{}
	GetTableConfigurationUpdate func(tableConfigMap map[string]interface{}, databaseName string, tableName string) map[string]interface{}
	ApplyDependencies           func(tableConfig interface{}, db interface{}, log *log.Logger) error
	GetTableSchema              func(tableName string) interface{}
	GetIndexedPathExt           func(engine interface{}, rowDataMap map[string]interface{}, indexColumnNames interface{}, databaseName string, tableName string, dbCallBack func(interface{}, map[string]interface{}) (string, []string, [][]interface{}, error)) (string, error)
	GetTableIndexColumnName     func() string
}

type FlowContext interface {
	IsInit() bool
	SetInit(bool)
	IsRestart() bool
	SetFlowDefinitionContext(*FlowDefinitionContext)
	GetFlowDefinitionContext() *FlowDefinitionContext
	SetRestart(bool)
	NotifyFlowComponentLoaded() // Notify that a critical flow is loaded
	WaitFlowLoaded()            // Block until all flows are loaded
	CancelTheContext() bool
	FlowSyncModeMatchAny([]string) bool
	FlowSyncModeMatch(string, bool) bool
	GetFlowSyncMode() string
	SetFlowSyncMode(string)
	GetFlowSourceAlias() string
	SetFlowSourceAlias(string)
	SetChangeFlowName(string)
	GetFlowStateState() int64
	GetFlowState() CurrentFlowState
	SetFlowState(CurrentFlowState)
	GetPreviousFlowState() CurrentFlowState
	SetPreviousFlowState(CurrentFlowState)
	TransitionState(string)
	SetFlowData(TemplateData)
	HasFlowSyncFilters() bool
	GetFlowStateSyncFilterRaw() string
	GetFlowSyncFilters() []string
	GetFlowName() string
	NewFlowStateUpdate(string, string) FlowStateUpdate
	GetCurrentFlowStateUpdateByDataSource(string) interface{}
	UpdateFlowStateByDataSource(string)
	PushState(string, FlowStateUpdate)
	GetUpdatePermission() PermissionUpdate
	GetFlowUpdate(CurrentFlowState) FlowStateUpdate
	GetDataSourceRegions(bool) []string
	GetRemoteDataSourceAttribute(string, ...string) interface{} // region, attribute
	// tfContext.NewFlowStateUpdate(strconv.Itoa(int(previousState.State)), tfContext.GetPreviousFlowSyncMode())
	GetLogger() *log.Logger
}

func SyncCheck(syncMode string) string {
	switch syncMode {
	case "nosync":
		return " with no syncing"
	case "push":
		return " with push sync"
	case "pull":
		return " with pull sync"
	case "pullonce":
		return " to pull once"
	case "pushonce":
		return " to push once"
	case "pullsynccomplete":
		return " - Pull synccomplete..waiting for new syncMode value"
	case "pullcomplete":
		return " - Pull complete..waiting for new syncMode value"
	case "pushcomplete":
		return " - Push complete..waiting for new syncMode value"
	case "pusherror":
		return " - Push error..waiting for new syncMode value"
	case "pullerror":
		return " - Pull error..waiting for new syncMode value"
	default:
		return "...waiting for new syncMode value"
	}
}

type FlowMachineContext interface {
	GetEnv() string
	GetFlowContext(FlowNameType) FlowContext
	GetDatabaseName() string
	GetTableModifierLock() *sync.Mutex
	TableCollationIdGen(string) interface{}
	Init(map[string]map[string]interface{}, []string, []FlowNameType, []FlowNameType) error
	AddTableSchema(interface{}, FlowContext)
	CreateTableTriggers(FlowContext, string)
	CreateTable(name string, schema interface{}, collation interface{}) error
	CreateCompositeTableTriggers(FlowContext, string, string, func(string, string, string, string) string, func(string, string, string, string) string, func(string, string, string, string) string)
	CreateDataFlowTableTriggers(FlowContext, string, string, string, func(string, string, string, string, string) string, func(string, string, string, string, string) string, func(string, string, string, string, string) string)
	GetFlowConfiguration(FlowContext, string) (map[string]interface{}, bool)
	ProcessFlow(FlowContext, func(FlowMachineContext, FlowContext) error, map[string]interface{}, map[string]map[string]interface{}, FlowNameType, FlowType) error
	SetPermissionUpdate(FlowContext) // tfmContext.SetPermissionUpdate(tfContext)
	//	seedVaultCycle(FlowContext, string, interface{}, func(interface{}, map[string]interface{}, interface{}, string, string, func(interface{}, map[string]interface{}) (string, []string, [][]interface{}, error)) (string, error), func(FlowContext, map[string]interface{}, map[string]interface{}, []string) error, bool)
	//	seedTrcDbCycle(FlowContext, string, interface{}, func(interface{}, map[string]interface{}, interface{}, string, string, func(interface{}, map[string]interface{}) (string, []string, [][]interface{}, error)) (string, error), func(FlowContext, map[string]interface{}, map[string]interface{}, []string) error, bool, chan bool)
	SyncTableCycle(FlowContext,
		string,
		interface{},
		func(interface{},
			map[string]interface{},
			interface{},
			string,
			string,
			func(interface{}, map[string]interface{}) (string, []string, [][]interface{}, error)) (string, error),
		func(FlowContext, map[string]interface{}) error,
		bool)
	SelectFlowChannel(FlowContext) <-chan interface{}
	GetAuthExtended(func(map[string]interface{}) map[string]interface{}, bool) (map[string]interface{}, error) // Auth for communicating with other services
	GetCacheRefreshSqlConn(FlowContext, string) (interface{}, error)
	CallDBQuery(FlowContext, map[string]interface{}, map[string]interface{}, bool, string, []FlowNameType, string) ([][]interface{}, bool)
	GetDbConn(FlowContext, string, string, map[string]interface{}) (interface{}, error)
	CallAPI(map[string]string, string, string, io.Reader, bool) (map[string]interface{}, int, error)
	SetEncryptionSecret()
	Log(string, error)
	LogInfo(string)
	GetLogger() *log.Logger
	PathToTableRowHelper(FlowContext) ([]interface{}, error)
	DeliverTheStatistic(FlowContext, *tccore.TTDINode, string, string, string, bool)
	LoadBaseTemplate(FlowContext) (TemplateData, error) //var baseTableTemplate extract.TemplateResultData , tfContext.GoMod, tfContext.FlowSource, tfContext.Flow.ServiceName(), tfContext.FlowPath

	//	writeToTableHelper(FlowContext, map[string]string, map[string]string) []interface{}
}
