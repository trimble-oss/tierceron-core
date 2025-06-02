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
func CompareRows(a map[string]interface{}, b map[string]interface{}) bool {
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

func tableConfigurationFlowPullRemote(tfmContext FlowMachineContext, tfContext FlowContext) ([]map[string]interface{}, error) {
	// b. Retrieve table configurations
	flowDefinitionContext := tfContext.GetFlowDefinitionContext()
	regionSyncList := tfContext.GetDataSourceRegions(true)
	var tableConfigMapArr []map[string]interface{}
	if len(regionSyncList) == 0 {
		return tableConfigMapArr, nil
	}

	var sqlConn *sql.DB
	sort.Strings(regionSyncList) //Puts west at the end
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
		if flowDefinitionContext.GetTableConfigurationById != nil {
			tableConfigurations, err := flowDefinitionContext.GetTableConfigurations(sqlConn, tfmContext.GetEnv() == "staging") //Staging is a special case that needs to be keyed off existing SEC eids to avoid prod data being pulled in.
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
			flowDefinitionContext.GetRefreshTableConfiguration(tfmContext, tfContext, sqlConn) //Staging is a special case that needs to be keyed off existing SEC eids to avoid prod data being pulled in.
		}

		tfmContext.Log("Finished pulling in table configurations from "+region, nil)

	}
	return tableConfigMapArr, nil
}

func tableConfigurationFlowPushRemote(tfContext FlowContext, changedItem map[string]interface{}) error {
	flowDefinitionContext := tfContext.GetFlowDefinitionContext()
	regionSyncList := tfContext.GetDataSourceRegions(true)
	if len(regionSyncList) == 0 {
		return nil
	}
	if flowDefinitionContext.GetTableFromMap != nil {
		var sqlConnI interface{}
		for _, region := range regionSyncList {
			if regionSource, ok := tfContext.GetRemoteDataSourceAttribute(region).(map[string]interface{}); ok {
				if conn, ok := regionSource["connection"]; ok && conn != nil {
					sqlConnI = conn
				}
			} else {
				continue
			}

			sqlIngestInterval := tfContext.GetRemoteDataSourceAttribute("dbingestinterval").(time.Duration)
			if sqlIngestInterval > 0 && tfContext.FlowSyncModeMatch("push", true) {
				if _, ok := changedItem["Deleted"]; ok {
					return nil
				} else {
					table := flowDefinitionContext.GetTableFromMap(changedItem)

					/*
						//region check before pushing
						if !strings.Contains(table.Region.String, trcRemoteDataSource[region].(map[string]interface{})["dbsourceregion"].(string)) {
							if trcRemoteDataSource[region].(map[string]interface{})["dbsourceregion"].(string) != "west" { //default to west if region doesn't match.
								return nil
							}
						}
					*/

					if flowDefinitionContext.ApplyDependencies != nil {
						if tfContext.HasFlowSyncFilters() {
							syncFilter := tfContext.GetFlowSyncFilters()
							for _, filter := range syncFilter {
								if filter == flowDefinitionContext.GetFilterFieldFromConfig(table) {
									err := flowDefinitionContext.ApplyDependencies(table, sqlConnI, tfContext.GetLogger()) //Attempts to update only the eid on push
									if err != nil {
										tfContext.PushState("flowStateReceiver", tfContext.NewFlowStateUpdate("2", "pusherror"))
										return err
									}
								}
							}
						} else {
							err := flowDefinitionContext.ApplyDependencies(table, sqlConnI, tfContext.GetLogger()) //Attempts to update only the eid on push
							if err != nil {
								tfContext.PushState("flowStateReceiver", tfContext.NewFlowStateUpdate("2", "pusherror"))
								return err
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func ProcessTableConfigurations(tfmContext FlowMachineContext, tfContext FlowContext) error {
	flowDefinitionContext := tfContext.GetFlowDefinitionContext()
	tfmContext.AddTableSchema(flowDefinitionContext.GetTableSchema(tfContext.GetFlowName()), tfContext)
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
		afterTime := time.Duration(0)
		for {
			select {
			case <-time.After(time.Millisecond * afterTime):
				afterTime = sqlIngestInterval
				//Logic for start/stopping flow
				if tfContext.GetFlowStateState() == 3 {
					tfContext.SetRestart(false)
					tfmContext.SetPermissionUpdate(tfContext)
					if tfContext.CancelTheContext() {
						//This cancel also pushes any final changes to vault before closing sync cycle.
						tableDefinition, _ := tfmContext.LoadBaseTemplate(tfContext)
						tfContext.SetFlowData(tableDefinition)
					}
					tfmContext.Log(fmt.Sprintf("%s flow is being stopped...", tfContext.GetFlowName()), nil)
					tfContext.PushState("flowStateReceiver", tfContext.NewFlowStateUpdate("0", tfContext.GetFlowSyncMode()))
					continue
				} else if tfContext.GetFlowStateState() == 0 {
					tfmContext.Log(fmt.Sprintf("%s flow is currently offline...", tfContext.GetFlowName()), nil)
					continue
				} else if tfContext.GetFlowStateState() == 1 {
					tfmContext.Log(fmt.Sprintf("%s flow is restarting...", tfContext.GetFlowName()), nil)
					if !tfContext.IsInit() { //init vault sync cycle
						tfContext.SetInit(true)
						tfmContext.CallDBQuery(tfContext, map[string]interface{}{"TrcQuery": "truncate " + tfContext.GetFlowSourceAlias() + "." + tfContext.GetFlowName()}, nil, false, "DELETE", nil, "")
					}
					tfContext.PushState("flowStateReceiver", tfContext.NewFlowStateUpdate("2", tfContext.GetFlowSyncMode()))
					continue
				} else if tfContext.GetFlowStateState() == 2 {
					if tfContext.IsInit() { //init vault sync cycle
						go tfmContext.SyncTableCycle(tfContext, flowDefinitionContext.GetTableIndexColumnNames(), flowDefinitionContext.GetTableIndexColumnNames(), flowDefinitionContext.GetIndexedPathExt, tableConfigurationFlowPushRemote, tfContext.GetFlowSyncMode() == "push")
					}
				}

				if tfContext.GetFlowStateState() != 0 && (tfContext.FlowSyncModeMatchAny([]string{"pull", "pullonce", "push", "pushonce", "pusheast"}) && prod.IsProd()) { //pusheast is unique for isProd() as it pushes both east/west
				} else if tfContext.FlowSyncModeMatch("pull", true) || tfContext.FlowSyncModeMatch("push", true) {
				} else {
					tfmContext.Log(fmt.Sprintf("%s is setup%s.", tfContext.GetFlowName(), SyncCheck(tfContext.GetFlowSyncMode())), nil)
					continue
				}

				tfmContext.Log(fmt.Sprintf("%s is running and checking for changes %s.", tfContext.GetFlowName(), SyncCheck(tfContext.GetFlowSyncMode())), nil)

				//Logic for push/pull once
				if tfContext.FlowSyncModeMatch("push", true) {
					switch syncSuffix := strings.TrimPrefix(tfContext.GetFlowSyncMode(), "push"); syncSuffix {
					case "once":
					default:
						if len(tfContext.GetDataSourceRegions(true)) == 0 {
							continue
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

					rows, _ := tfmContext.CallDBQuery(tfContext, map[string]interface{}{"TrcQuery": "SELECT * FROM " + tfContext.GetFlowSourceAlias() + "." + tfContext.GetFlowName()}, nil, false, "SELECT", nil, "")
					if len(rows) == 0 {
						tfmContext.Log(fmt.Sprintf("Nothing in %s table to push out yet...", tfContext.GetFlowName()), nil) //Table is not currently loaded.
						continue
					}
					for _, value := range rows {
						tableMap := flowDefinitionContext.GetTableMapFromArray(value)
						pushError := tableConfigurationFlowPushRemote(tfContext, tableMap)
						if pushError != nil {
							tfmContext.Log(fmt.Sprintf("Error pushing out %s", tfContext.GetFlowName()), pushError)
							continue
						}
					}
					tfContext.SetFlowSyncMode("pushcomplete")
					tfContext.PushState("flowStateReceiver", tfContext.NewFlowStateUpdate("2", "pushcomplete"))
					continue
				}

				// 3. Retrieve table configurations from mysql.
				tableConfigurations, err := tableConfigurationFlowPullRemote(tfmContext, tfContext)
				if err != nil {
					tfmContext.Log("Error grabbing table configurations", err)
					tfContext.PushState("flowStateReceiver", tfContext.NewFlowStateUpdate("2", "pullerror"))
					continue
				}

				var filterTableConfigurations []map[string]interface{}
				if tfContext.HasFlowSyncFilters() {
					syncFilter := tfContext.GetFlowSyncFilters()
					for _, filter := range syncFilter {
						for _, table := range tableConfigurations {
							if filter == table["tableId"].(string) {
								filterTableConfigurations = append(filterTableConfigurations, table)
							}
						}
					}
					tableConfigurations = filterTableConfigurations
				}

				for _, table := range tableConfigurations {
					rows, _ := tfmContext.CallDBQuery(tfContext, flowDefinitionContext.GetTableConfigurationById(tfContext.GetFlowSourceAlias(), tfContext.GetFlowName(), table["tableId"].(string)), nil, false, "SELECT", nil, "")
					if len(rows) == 0 {
						tfmContext.CallDBQuery(tfContext, flowDefinitionContext.GetTableConfigurationInsert(table, tfContext.GetFlowSourceAlias(), tfContext.GetFlowName()), nil, true, "INSERT", []FlowNameType{FlowNameType(tfContext.GetFlowName())}, "") //if DNE -> insert
					} else {
						for _, value := range rows {
							// tableConfig is db, value is what's in vault...
							if CompareRows(table, flowDefinitionContext.GetTableMapFromArray(value)) { //If equal-> do nothing
								continue
							} else { //If not equal -> update
								tfmContext.CallDBQuery(tfContext, flowDefinitionContext.GetTableConfigurationUpdate(table, tfContext.GetFlowSourceAlias(), tfContext.GetFlowName()), nil, true, "UPDATE", []FlowNameType{FlowNameType(tfContext.GetFlowName())}, "")
							}
						}
					}
				}

				if tfContext.GetFlowSyncMode() != "pullerror" && tfContext.GetFlowSyncMode() != "pullcomplete" {
					tfContext.SetFlowSyncMode("pullcomplete")
					tfContext.PushState("flowStateReceiver", tfContext.NewFlowStateUpdate("2", "pullcomplete"))
					// Now go to vault.
					//tfContext.Restart = true
					//tfContext.CancelTheContext() // Anti pattern...
				}
			}
		}
	}
	tfContext.CancelTheContext()
	return nil
}
