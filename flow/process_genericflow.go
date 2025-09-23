package flow

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	prod "github.com/trimble-oss/tierceron-core/v2/prod"
)

// True == equal, false = not equal
func CompareRows(a map[string]any, b map[string]any) bool {
	for key, value := range a {
		if _, ok := b[key].(string); ok {
			if b[key] != value {
				return false
			}
		} else {
			valueStr := fmt.Sprintf("%v", b[key])
			if valueStr != value {
				return false
			}
		}
	}
	return true
}

func tableConfigurationIndicesFlowPullRemote(tfmContext FlowMachineContext, tfContext FlowContext) ([]map[string]any, error) {
	tfContext.WaitFlowLoaded()
	// b. Retrieve table configurations
	flowDefinitionContext := tfContext.GetFlowLibraryContext()
	regionSyncList := tfContext.GetDataSourceRegions(true)
	var tableIndexMapArr []map[string]any
	if len(regionSyncList) == 0 {
		return tableIndexMapArr, nil
	}

	var sqlConn *sql.DB
	sort.Strings(regionSyncList) // Puts west at the end
	for _, region := range regionSyncList {
		var ok bool
		sqlConnI, sqlConnErr := tfmContext.GetCacheRefreshSqlConn(tfContext, region)
		if sqlConnErr != nil {
			tfmContext.Log("Unable to obtain data source connection for tableConfiguration for "+region+" during pull", nil)
		}
		sqlConn, ok = sqlConnI.(*sql.DB)
		if !ok {
			tfmContext.Log("Unable to obtain data source connection for tableConfiguration for "+region+" during pull", nil)
		}

		tfmContext.Log("Attempting to pull in table configurations from "+region, nil)
		if flowDefinitionContext.GetTableIndices != nil {
			tableIndices, err := flowDefinitionContext.GetTableIndices(sqlConn) // Staging is a special case that needs to be keyed off existing SEC eids to avoid prod data being pulled in.
			if err != nil {
				return nil, err
			}
			if len(tableIndices) > 0 {
				tfmContext.Log("Found "+fmt.Sprintf("%d", len(tableIndices))+" tableindex rows from "+region, nil)
			}
			for _, tableConfiguration := range tableIndices {
				tableIndexMapArr = append(tableIndexMapArr, flowDefinitionContext.GetSparseTableMap(tableConfiguration))
			}

		}

		tfmContext.Log("Finished pulling in table configurations indices from "+region, nil)

	}
	return tableIndexMapArr, nil
}

func tableConfigurationFlowPullRemoteByIndices(tfmContext FlowMachineContext, tfContext FlowContext, tableIndices []string) ([]map[string]any, error) {
	tfContext.WaitFlowLoaded()
	// b. Retrieve table configurations
	flowDefinitionContext := tfContext.GetFlowLibraryContext()
	regionSyncList := tfContext.GetDataSourceRegions(true)
	var tableConfigMapArr []map[string]any
	if len(regionSyncList) == 0 {
		return tableConfigMapArr, nil
	}

	var sqlConn *sql.DB
	sort.Strings(regionSyncList) // Puts west at the end
	for _, region := range regionSyncList {
		var ok bool
		sqlConnI, sqlConnErr := tfmContext.GetCacheRefreshSqlConn(tfContext, region)
		if sqlConnErr != nil {
			tfmContext.Log("Unable to obtain data source connection for tableConfiguration for "+region+" during pull", nil)
		}
		sqlConn, ok = sqlConnI.(*sql.DB)
		if !ok {
			tfmContext.Log("Unable to obtain data source connection for tableConfiguration for "+region+" during pull", nil)
		}

		tfmContext.Log("Attempting to pull in table configurations from "+region, nil)
		for _, tableId := range tableIndices {
			if flowDefinitionContext.GetTableConfigurationById != nil {
				tableQueryMap := flowDefinitionContext.GetTableConfigurationById(
					tfmContext.GetDatabaseName(TrcDb),
					string(tfContext.GetFlowHeader().Name),
					tableId,
				)
				tableConfigurations, err := flowDefinitionContext.GetTableConfigurationsByQuery(sqlConn, tableQueryMap["TrcQuery"].(string)) // Staging is a special case that needs to be keyed off existing SEC eids to avoid prod data being pulled in.
				if err != nil {
					return nil, err
				}
				if len(tableConfigurations) > 0 {
					tfmContext.Log("Found "+fmt.Sprintf("%d", len(tableConfigurations))+" tableconfiguration rows from "+region, nil)
					for _, tableConfig := range tableConfigurations {
						tableConfigMapArr = append(tableConfigMapArr, flowDefinitionContext.GetTableMap(tableConfig))
					}
				}
			}
		}
		tfmContext.Log("Finished pulling in table configurations from "+region, nil)
	}
	return tableConfigMapArr, nil
}

func tableConfigurationFlowPullRemote(tfmContext FlowMachineContext, tfContext FlowContext) ([]map[string]any, error) {
	tfContext.WaitFlowLoaded()
	// b. Retrieve table configurations
	flowDefinitionContext := tfContext.GetFlowLibraryContext()
	regionSyncList := tfContext.GetDataSourceRegions(true)
	var tableConfigMapArr []map[string]any
	if len(regionSyncList) == 0 {
		return tableConfigMapArr, nil
	}

	var sqlConn *sql.DB
	sort.Strings(regionSyncList) // Puts west at the end
	for _, region := range regionSyncList {
		var ok bool
		sqlConnI, sqlConnErr := tfmContext.GetCacheRefreshSqlConn(tfContext, region)
		if sqlConnI == nil || sqlConnErr != nil {
			tfmContext.Log("Unable to obtain data source connection for tableConfiguration for "+region+" during pull", nil)
			continue
		}
		sqlConn, ok = sqlConnI.(*sql.DB)
		if !ok {
			tfmContext.Log("Unable to obtain data source connection for tableConfiguration for "+region+" during pull", nil)
			continue
		}

		tfmContext.Log("Attempting to pull in table configurations from "+region, nil)
		if flowDefinitionContext.GetTableConfigurationById != nil {
			tableConfigurations, err := flowDefinitionContext.GetTableConfigurations(sqlConn, tfmContext.GetEnv() == "staging") // Staging is a special case that needs to be keyed off existing SEC eids to avoid prod data being pulled in.
			if err != nil {
				return nil, err
			}
			if len(tableConfigurations) > 0 {
				tfmContext.Log("Found "+fmt.Sprintf("%d", len(tableConfigurations))+" tableconfiguration rows from "+region, nil)
			}
			for _, tableConfiguration := range tableConfigurations {
				tableConfigMapArr = append(tableConfigMapArr, flowDefinitionContext.GetTableMap(tableConfiguration))
			}
		} else if flowDefinitionContext.GetRefreshTableConfiguration != nil {
			flowDefinitionContext.GetRefreshTableConfiguration(tfmContext, tfContext, sqlConn) // Staging is a special case that needs to be keyed off existing SEC eids to avoid prod data being pulled in.
		}

		tfmContext.Log("Finished pulling in table configurations from "+region, nil)

	}
	return tableConfigMapArr, nil
}

func tableConfigurationFlowPushRemote(tfContext FlowContext, changedItem map[string]any) error {
	tfContext.WaitFlowLoaded()
	flowDefinitionContext := tfContext.GetFlowLibraryContext()
	regionSyncList := tfContext.GetDataSourceRegions(true)
	if len(regionSyncList) == 0 {
		return nil
	}
	if flowDefinitionContext.GetTableFromMap != nil {
		var sqlConnI any
		for _, region := range regionSyncList {
			if regionSource, ok := tfContext.GetRemoteDataSourceAttribute(region).(map[string]any); ok {
				if conn, ok := regionSource["connection"]; ok && conn != nil {
					if dbConn, ok := conn.(*sql.DB); ok && dbConn != nil {
						sqlConnI = conn
					} else {
						continue
					}
				} else {
					continue
				}
			} else {
				continue
			}

			if _, ok := changedItem["Deleted"]; ok {
				return nil
			} else {
				table := flowDefinitionContext.GetTableFromMap(changedItem)

				if flowDefinitionContext.ApplyDependencies != nil {
					if tfContext.HasFlowSyncFilters() {
						syncFilter := tfContext.GetFlowSyncFilters()
						for _, filter := range syncFilter {
							if filter == flowDefinitionContext.GetFilterFieldFromConfig(table) {
								err := flowDefinitionContext.ApplyDependencies(table, sqlConnI, tfContext.GetLogger()) // Attempts to update only the eid on push
								if err != nil {
									tfContext.PushState("flowStateReceiver", tfContext.NewFlowStateUpdate("2", "pusherror"))
									return err
								}
							}
						}
					} else {
						err := flowDefinitionContext.ApplyDependencies(table, sqlConnI, tfContext.GetLogger()) // Attempts to update only the eid on push
						if err != nil {
							tfContext.PushState("flowStateReceiver", tfContext.NewFlowStateUpdate("2", "pusherror"))
							return err
						}
					}
				}
			}
		}
	}

	return nil
}

func ProcessTableConfigurations(tfmContext FlowMachineContext, tfContext FlowContext) error {
	flowDefinitionContext := tfContext.GetFlowLibraryContext()
	tfmContext.AddTableSchema(flowDefinitionContext.GetTableSchema(tfContext.GetFlowHeader().FlowName()), tfContext)
	if flowDefinitionContext.CreateTableTriggers != nil {
		flowDefinitionContext.CreateTableTriggers(tfmContext, tfContext)
	} else {
		tfmContext.CreateTableTriggers(tfContext, flowDefinitionContext.GetTableIndexColumnNames())
	}
	go tfContext.TransitionState("nosync")
	tfContext.SetInit(true)

	sqlIngestInterval := tfContext.GetRemoteDataSourceAttribute("dbingestinterval").(time.Duration)

	regionList := make([]string, 0)

	if sqlIngestInterval > 0 {
		// Implement pull from remote data source
		// Only pull if ingest interval is set to > 0 value.
		ProcessFlowStatesForInterval(tfContext, tfmContext, flowDefinitionContext, regionList)
		ticker := time.NewTicker(time.Second * sqlIngestInterval)
		defer ticker.Stop()

		for range ticker.C {
			// Logic for start/stopping flow
			ProcessFlowStatesForInterval(tfContext, tfmContext, flowDefinitionContext, regionList)
		}
	}
	tfContext.CancelTheContext()
	return nil
}

func ProcessFlowStatesForInterval(tfContext FlowContext, tfmContext FlowMachineContext, flowDefinitionContext *FlowLibraryContext, regionList []string) int {
	if tfContext.GetFlowStateState() == 3 {
		tfContext.SetRestart(false)
		tfmContext.SetPermissionUpdate(tfContext)
		if tfContext.CancelTheContext() {
			// This cancel also pushes any final changes to vault before closing sync cycle.
			tableDefinition, _ := tfmContext.LoadBaseTemplate(tfContext)
			tfContext.SetFlowData(tableDefinition)
		}
		tfmContext.Log(fmt.Sprintf("%s flow is being stopped...", tfContext.GetFlowHeader().FlowName()), nil)
		tfContext.PushState("flowStateReceiver", tfContext.NewFlowStateUpdate("0", tfContext.GetFlowSyncMode()))
		return 1
	} else if tfContext.GetFlowStateState() == 0 {
		tfmContext.Log(fmt.Sprintf("%s flow is currently offline...", tfContext.GetFlowHeader().FlowName()), nil)
		return 2
	} else if tfContext.GetFlowStateState() == 1 {
		tfmContext.Log(fmt.Sprintf("%s flow is restarting...", tfContext.GetFlowHeader().FlowName()), nil)
		if !tfContext.IsInit() { // init vault sync cycle
			tfContext.SetInit(true)
			tfmContext.CallDBQuery(tfContext, map[string]any{"TrcQuery": "truncate " + tfContext.GetFlowHeader().SourceAlias + "." + tfContext.GetFlowHeader().FlowName()}, nil, false, "DELETE", nil, "")
		}
		tfContext.PushState("flowStateReceiver", tfContext.NewFlowStateUpdate("2", tfContext.GetFlowSyncMode()))
		return 3
	} else if tfContext.GetFlowStateState() == 2 {
		if tfContext.IsInit() { // init vault sync cycle
			tfContext.SetInit(false)
			tfContext.InitNotify()
			shouldSyncRemote := tfContext.GetFlowSyncMode() == "push"
			if shouldSyncFunc := flowDefinitionContext.ShouldSyncRemote; shouldSyncFunc != nil {
				shouldSyncRemote = shouldSyncRemote || shouldSyncFunc(SyncRemoteModeFlowDataCyclic)
			}
			go tfmContext.SyncTableCycle(tfContext, flowDefinitionContext.GetTableIndexColumnNames(), flowDefinitionContext.GetTableIndexColumnNames(), flowDefinitionContext.GetIndexedPathExt, tableConfigurationFlowPushRemote, shouldSyncRemote)
		}
	}

	if tfContext.GetFlowStateState() != 0 && (tfContext.FlowSyncModeMatchAny([]string{"pull", "pullonce", "push", "pushonce", "pusheast"}) && prod.IsProd()) { // pusheast is unique for isProd() as it pushes both east/west
	} else if (tfContext.FlowSyncModeMatch("pull", true) || tfContext.FlowSyncModeMatch("push", true)) && tfContext.GetFlowSyncMode() != "pullerror" && tfContext.GetFlowSyncMode() != "pullcomplete" {
	} else {
		tfmContext.Log(fmt.Sprintf("%s is setup%s.", tfContext.GetFlowHeader().FlowName(), SyncCheck(tfContext.GetFlowSyncMode())), nil)
		return 4
	}

	tfmContext.Log(fmt.Sprintf("%s is running and checking for changes %s.", tfContext.GetFlowHeader().FlowName(), SyncCheck(tfContext.GetFlowSyncMode())), nil)
	if !tfContext.IsPreloaded() {
		return 0
	}

	// Logic for push/pull once
	if tfContext.FlowSyncModeMatch("push", true) {
		switch syncSuffix := strings.TrimPrefix(tfContext.GetFlowSyncMode(), "push"); syncSuffix {
		case "once":
		default:
			if len(tfContext.GetDataSourceRegions(true)) == 0 {
				return 0
			}
			pullRegionFound := false
			for _, region := range regionList {
				if syncSuffix == region {
					pullRegionFound = true
				}
			}
			if !pullRegionFound {
				tfContext.FlowSyncModeMatch("pullregionerror", true)
				tfContext.PushState("flowStateReceiver", tfContext.NewFlowStateUpdate("2", "pushregionerror"))
			}
		}

		rows, _ := tfmContext.CallDBQuery(tfContext, map[string]any{"TrcQuery": "SELECT * FROM " + tfContext.GetFlowHeader().SourceAlias + "." + tfContext.GetFlowHeader().FlowName()}, nil, false, "SELECT", nil, "")
		if len(rows) == 0 {
			tfmContext.Log(fmt.Sprintf("Nothing in %s table to push out yet...", tfContext.GetFlowHeader().FlowName()), nil) // Table is not currently loaded.
			return 5
		}
		for _, value := range rows {
			tableMap := flowDefinitionContext.GetTableMapFromArray(value)
			pushError := tableConfigurationFlowPushRemote(tfContext, tableMap)
			if pushError != nil {
				tfmContext.Log(fmt.Sprintf("Error pushing out %s", tfContext.GetFlowHeader().FlowName()), pushError)
				continue
			}
		}
		tfContext.SetFlowSyncMode("pushcomplete")
		tfContext.PushState("flowStateReceiver", tfContext.NewFlowStateUpdate("2", "pushcomplete"))
		return 6
	}
	var tableIndexKey string
	if flowDefinitionContext.GetTableIndexColumnNames != nil {
		tableIndexKeys := flowDefinitionContext.GetTableIndexColumnNames()
		if len(tableIndexKeys) == 1 {
			tableIndexKey = tableIndexKeys[0]
		}
	}

	if len(tableIndexKey) == 0 {
		tfmContext.Log("Error pulling table configurations.  Missing required GetTableIndexColumnNames", nil)
		tfContext.PushState("flowStateReceiver", tfContext.NewFlowStateUpdate("2", "pullerror"))
		return 7
	}

	var tableConfigurations []map[string]any
	var tableConfigurationsError error
	// 2.5 Check for indices when available
	if flowDefinitionContext.GetTableIndices != nil && tfContext.GetFlowSyncMode() != "pullonce" {
		tableMapIndices, err := tableConfigurationIndicesFlowPullRemote(tfmContext, tfContext)
		if err == nil && len(tableMapIndices) > 0 {
			// Do a precheck...
			var tableIndices []string
			for _, tableIndexMap := range tableMapIndices {
				rows, _ := tfmContext.CallDBQuery(tfContext, flowDefinitionContext.GetTableConfigurationById(tfContext.GetFlowHeader().SourceAlias, tfContext.GetFlowHeader().FlowName(), tableIndexMap[tableIndexKey].(string)), nil, false, "SELECT", nil, "")
				if len(rows) == 0 {
					// store index for later...
					tableIndices = append(tableIndices, tableIndexMap[tableIndexKey].(string))
				} else {
					for _, value := range rows {
						// tableConfig is db, value is what's in vault...
						if CompareRows(tableIndexMap, flowDefinitionContext.GetTableMapFromArray(value)) { // If equal-> do nothing
							continue
						} else { // If not equal -> update
							tableIndices = append(tableIndices, tableIndexMap[tableIndexKey].(string))
						}
					}
				}
			}
			if len(tableIndices) > 0 {
				// Do something with tableIndices
				tableConfigurations, tableConfigurationsError = tableConfigurationFlowPullRemoteByIndices(tfmContext, tfContext, tableIndices)
			}
		}
	} else {
		tableConfigurations, tableConfigurationsError = tableConfigurationFlowPullRemote(tfmContext, tfContext)
	}

	// 3. Retrieve table configurations from mysql.
	if tableConfigurationsError != nil {
		tfmContext.Log("Error grabbing table configurations", tableConfigurationsError)
		tfContext.PushState("flowStateReceiver", tfContext.NewFlowStateUpdate("2", "pullerror"))
		return 7
	}

	var filterTableConfigurations []map[string]any
	if tfContext.HasFlowSyncFilters() {
		syncFilter := tfContext.GetFlowSyncFilters()
		for _, filter := range syncFilter {
			for _, table := range tableConfigurations {
				if filter == table[tableIndexKey].(string) {
					filterTableConfigurations = append(filterTableConfigurations, table)
				}
			}
		}
		tableConfigurations = filterTableConfigurations
	}

	for _, table := range tableConfigurations {
		rows, _ := tfmContext.CallDBQuery(tfContext, flowDefinitionContext.GetTableConfigurationById(tfContext.GetFlowHeader().SourceAlias, tfContext.GetFlowHeader().FlowName(), table[tableIndexKey].(string)), nil, false, "SELECT", nil, "")
		if len(rows) == 0 {
			tfmContext.CallDBQuery(tfContext, flowDefinitionContext.GetTableConfigurationInsert(table, tfContext.GetFlowHeader().SourceAlias, tfContext.GetFlowHeader().FlowName()), nil, true, "INSERT", []FlowNameType{tfContext.GetFlowHeader().FlowNameType()}, "") // if DNE -> insert
		} else {
			for _, value := range rows {
				// tableConfig is db, value is what's in vault...
				if CompareRows(table, flowDefinitionContext.GetTableMapFromArray(value)) { // If equal-> do nothing
					continue
				} else { // If not equal -> update
					tfmContext.CallDBQuery(tfContext, flowDefinitionContext.GetTableConfigurationUpdate(table, tfContext.GetFlowHeader().SourceAlias, tfContext.GetFlowHeader().FlowName()), nil, true, "UPDATE", []FlowNameType{tfContext.GetFlowHeader().FlowNameType()}, "")
				}
			}
		}
	}

	if tfContext.GetFlowSyncMode() != "pullerror" && tfContext.GetFlowSyncMode() != "pullcomplete" && tfContext.GetFlowSyncMode() != "pull" {
		tfContext.SetFlowSyncMode("pullcomplete")
		tfContext.PushState("flowStateReceiver", tfContext.NewFlowStateUpdate("2", "pullcomplete"))
		// Now go to vault.
		// tfContext.Restart = true
		// tfContext.CancelTheContext() // Anti pattern...
	}
	return 0
}
