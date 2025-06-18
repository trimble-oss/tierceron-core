package flow

import (
	"io"
	"log"
	"sync"

	tccore "github.com/trimble-oss/tierceron-core/v2/core"
)

type FlowType int64
type FlowColumnType int64
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

var DataFlowStatConfigurationsFlow FlowNameType = "DataFlowStatistics"
var ArgosSociiFlow FlowNameType = "ArgosSocii"

type PermissionUpdate any
type FlowStateUpdate interface {
	// NewFlowState() CurrentFlowState
}
type CurrentFlowState any
type TemplateData any

const (
	TinyText FlowColumnType = iota
	Text
	MediumText
	LongText
	TinyBlob
	Blob
	MediumBlob
	LongBlob
	Int8
	Uint8
	Int16
	Uint16
	Int24
	Uint24
	Int32
	Uint32
	Int64
	Uint64
	Float32
	Float64
	Timestamp
)

// The following are mappable types to go-mysql-server/sql column
type FlowColumn struct {
	Name           string
	Type           FlowColumnType
	AutoIncrement  bool
	Nullable       bool
	Source         string
	DatabaseSource string
	PrimaryKey     bool
	Comment        string
	Extra          string
}

type FlowDefinitionContext struct {
	GetTableConfigurationById        func(databaseName string, tableName string, idColumnName ...string) map[string]any
	GetTableConfigurations           func(db any, secLookup bool) ([]any, error)
	CreateTableTriggers              func(tfmContext FlowMachineContext, tfContext FlowContext) // Optional override
	GetRefreshTableConfiguration     func(tfmContext FlowMachineContext, tfContext FlowContext, dbI any) ([]any, error)
	GetTableMap                      func(tableConfig any) map[string]any
	GetTableFromMap                  func(tableConfigMap map[string]any) any
	GetFilterFieldFromConfig         func(tableconfig any) string
	GetTableMapFromArray             func(tableArray []any) map[string]any
	GetTableConfigurationInsert      func(tableConfigMap map[string]any, databaseName string, tableName string) map[string]any
	GetTableConfigurationUpdate      func(tableConfigMap map[string]any, databaseName string, tableName string) map[string]any
	ApplyDependencies                func(tableConfig any, db any, log *log.Logger) error
	GetTableSchema                   func(tableName string) any
	GetIndexedPathExt                func(engine any, rowDataMap map[string]any, indexColumnNames any, databaseName string, tableName string, dbCallBack func(any, map[string]any) (string, []string, [][]any, error)) (string, error)
	GetTableIndexColumnNames         func() []string
	GetTableGrant                    func(string) (string, string, error)
	GetFlowIndexComplex              func() (string, []string, string, error)
	TableConfigurationFlowPullRemote func(tfmContext FlowMachineContext, tfContext FlowContext) error
}

type FlowContext interface {
	IsInit() bool
	SetInit(bool)
	IsRestart() bool
	SetRestart(bool)
	SetFlowDefinitionContext(*FlowDefinitionContext)
	GetFlowDefinitionContext() *FlowDefinitionContext
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
	GetCurrentFlowStateUpdateByDataSource(string) any
	UpdateFlowStateByDataSource(string)
	PushState(string, FlowStateUpdate)
	GetUpdatePermission() PermissionUpdate
	GetFlowUpdate(CurrentFlowState) FlowStateUpdate
	GetDataSourceRegions(bool) []string
	GetRemoteDataSourceAttribute(string, ...string) any // region, attribute
	// tfContext.NewFlowStateUpdate(strconv.Itoa(int(previousState.State)), tfContext.GetPreviousFlowSyncMode())
	SetCustomSeedTrcdbFunc(func(FlowMachineContext, FlowContext) error)
	GetLogger() *log.Logger
	Log(string, error)
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
	TableCollationIdGen(string) any
	Init(map[string]map[string]any, []string, []FlowNameType, []FlowNameType) error
	AddTableSchema(any, FlowContext)
	CreateTableTriggers(FlowContext, []string)
	CreateTable(name string, schema any, collation any) error
	CreateCompositeTableTriggers(FlowContext, string, string, func(string, string, string, string) string, func(string, string, string, string) string, func(string, string, string, string) string)
	CreateDataFlowTableTriggers(FlowContext, string, string, string, func(string, string, string, string, string) string, func(string, string, string, string, string) string, func(string, string, string, string, string) string)
	GetFlowConfiguration(FlowContext, string) (map[string]any, bool)
	ProcessFlow(FlowContext, func(FlowMachineContext, FlowContext) error, map[string]any, map[string]map[string]any, FlowNameType, FlowType) error
	SetPermissionUpdate(FlowContext) // tfmContext.SetPermissionUpdate(tfContext)
	//	seedVaultCycle(FlowContext, string, any, func(any, map[string]any, any, string, string, func(any, map[string]any) (string, []string, [][]any, error)) (string, error), func(FlowContext, map[string]any, map[string]any, []string) error, bool)
	//	seedTrcDbCycle(FlowContext, string, any, func(any, map[string]any, any, string, string, func(any, map[string]any) (string, []string, [][]any, error)) (string, error), func(FlowContext, map[string]any, map[string]any, []string) error, bool, chan bool)
	SyncTableCycle(FlowContext,
		[]string, // index column names
		any,
		func(any,
			map[string]any,
			any,
			string,
			string,
			func(any, map[string]any) (string, []string, [][]any, error)) (string, error),
		func(FlowContext, map[string]any) error,
		bool)
	SelectFlowChannel(FlowContext) <-chan any
	GetAuthExtended(func(map[string]any) map[string]any, bool) (map[string]any, error) // Auth for communicating with other services
	GetCacheRefreshSqlConn(FlowContext, string) (any, error)
	CallDBQuery(FlowContext, map[string]any, map[string]any, bool, string, []FlowNameType, string) ([][]any, bool)
	CallDBQueryN(*tccore.TrcdbExchange, map[string]any, map[string]any, bool, string, []FlowNameType, string) (*tccore.TrcdbExchange, bool)
	GetDbConn(FlowContext, string, string, map[string]any) (any, error)
	CallAPI(map[string]string, string, string, io.Reader, bool) (map[string]any, int, error)
	SetEncryptionSecret()
	Log(string, error)
	LogInfo(string)
	GetLogger() *log.Logger
	PathToTableRowHelper(FlowContext) ([]any, error)
	DeliverTheStatistic(FlowContext, *tccore.TTDINode, string, string, string, bool)
	LoadBaseTemplate(FlowContext) (TemplateData, error) //var baseTableTemplate extract.TemplateResultData , tfContext.GoMod, tfContext.FlowSource, tfContext.Flow.ServiceName(), tfContext.FlowPath
	WaitAllFlowsLoaded()                                // Block until all flows are loaded

	//	writeToTableHelper(FlowContext, map[string]string, map[string]string) []any
}
