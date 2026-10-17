package flow

// FlowPersistenceContext provides access to a flow's local database and remote
// records while its synchronization locks are held. Use it only in the callback.
type FlowPersistenceContext interface {
	Source() string
	DatabaseName() string
	Query(string) ([][]any, error)
	ReadRecord(string) (map[string]any, error)
	WriteRecord(string, map[string]any) error
}

// FlowPersistenceMachine is optional; it does not extend FlowMachineContext.
// writeRemote permits remote writes, not local database updates. The callback
// runs synchronously under the same locks as background flow persistence.
type FlowPersistenceMachine interface {
	WithFlowPersistence(FlowNameType, bool, func(FlowPersistenceContext) error) error
}
