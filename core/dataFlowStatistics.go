package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	//"os"
	"strconv"
	"strings"

	"time"

	"github.com/trimble-oss/tierceron-nute-core/mashupsdk"
)

type TTDINode struct {
	*mashupsdk.MashupDetailedElement
	ChildNodes []*TTDINode
}

type DeliverStatCtx struct {
	FlowGroup string
	FlowName  string
	StateCode string
	StateName string
	LogStat   bool
	LogFunc   *func(string, error)
	//TODO: Make not interface
	TimeStart      any //either string or time.Time
	LastTestedDate any //either string or time.Time
	TimeSplit      any //either float64 or time.Duration
	Mode           any //either float64 or int
}

func (dsc *DeliverStatCtx) GetElapsedTimeStr() string {
	var elapsedTime string
	if _, ok := dsc.TimeSplit.(time.Duration); ok {
		if dsc.TimeSplit != nil && dsc.TimeSplit.(time.Duration).Seconds() < 0 { //Covering corner case of 0 second time durations being slightly off (-.00004 seconds)
			elapsedTime = "0s"
		} else {
			elapsedTime = dsc.TimeSplit.(time.Duration).Truncate(time.Millisecond * 10).String()
		}
	} else if timeFloat, ok := dsc.TimeSplit.(float64); ok {
		elapsedTime = time.Duration(timeFloat * float64(time.Nanosecond)).Truncate(time.Millisecond * 10).String()
	}
	return elapsedTime
}

func (dsc *DeliverStatCtx) GetModeInt() int {
	var modeInt int
	if modeFloat, ok := dsc.Mode.(float64); ok {
		modeInt = int(modeFloat)
	} else if mi, ok := dsc.Mode.(int); ok {
		modeInt = mi
	} else {
		modeInt = -1
	}
	return modeInt
}

func (dsc *DeliverStatCtx) GetLastTestedDateStr() string {
	lastTestedDate := ""
	if t, ok := dsc.TimeStart.(string); ok {
		lastTestedDate = t
	} else if ltd, ok := dsc.TimeStart.(string); ok {
		lastTestedDate = ltd
	}
	return lastTestedDate
}

func InitDataFlow(logF func(string, error), name string, logS bool) *TTDINode {
	var stats []*TTDINode
	data := make(map[string]any)
	data["TimeStart"] = time.Now().Format(RFC_ISO_8601)
	data["LogStat"] = logS
	if logF != nil {
		data["LogFunc"] = logF
	}
	encodedData, err := json.Marshal(&data)
	if err != nil {
		log.Println("Error in encoding data in InitDataFlow")
		return &TTDINode{MashupDetailedElement: &mashupsdk.MashupDetailedElement{}}
	}
	ttdiNode := &TTDINode{&mashupsdk.MashupDetailedElement{Name: name, State: &mashupsdk.MashupElementState{State: int64(mashupsdk.Init)}, Data: string(encodedData)}, stats}
	return ttdiNode
}

func (dfs *TTDINode) UpdateDataFlowStatistic(flowG string, flowN string, stateN string, stateC string, mode int, logF func(string, error)) {
	var decodedData map[string]any
	var timeStart time.Time
	var err error
	if len(dfs.MashupDetailedElement.Data) > 0 {
		var dfsctx *DeliverStatCtx
		dfsctx, decodedData, err = dfs.GetDeliverStatCtx()
		if err != nil {
			logF("Error in decoding data in UpdateDataFlowStatistic", err)
			return
		}

		//string to time.time
		if ts, ok := dfsctx.TimeStart.(time.Time); ok {
			timeStart = ts
		} else {
			var timeParseErr error
			timeStart, timeParseErr = time.Parse(RFC_ISO_8601, dfsctx.TimeStart.(string))
			if timeParseErr != nil {
				logF("Error in parsing start time in UpdateDataFlowStatistics", timeParseErr)
				return
			}
		}
	} else {
		decodedData = make(map[string]any)
		timeStart = time.Now()
		decodedData["TimeStart"] = timeStart.Format(RFC_ISO_8601)

		newEncodedData, err := json.Marshal(decodedData)
		if err != nil {
			logF("Error in encoding data in UpdateDataFlowStatistics", err)
			return
		}
		dfs.MashupDetailedElement.Data = string(newEncodedData)
	}

	newData := make(map[string]any)
	newData["FlowGroup"] = flowG
	newData["FlowName"] = flowN
	newData["StateName"] = stateN
	newData["StateCode"] = stateC
	newData["Mode"] = mode
	newData["TimeSplit"] = time.Since(timeStart)
	newData["TimeStart"] = timeStart
	newEncodedData, err := json.Marshal(newData)
	if err != nil {
		logF("Error in encoding data in UpdateDataFlowStatistics", err)
		return
	}
	newNode := TTDINode{MashupDetailedElement: &mashupsdk.MashupDetailedElement{Data: string(newEncodedData)}, ChildNodes: []*TTDINode{}}
	dfs.ChildNodes = append(dfs.ChildNodes, &newNode)
	newData["decodedData"] = decodedData
	dfs.EfficientLog(newData, logF)
}

func (dfs *TTDINode) UpdateDataFlowStatisticWithTime(flowG string, flowN string, stateN string, stateC string, mode int, elapsedTime time.Duration) {
	newData := make(map[string]any)
	newData["FlowGroup"] = flowG
	newData["FlowName"] = flowN
	newData["StateName"] = stateN
	newData["StateCode"] = stateC
	newData["Mode"] = mode
	newData["TimeSplit"] = elapsedTime
	newEncodedData, err := json.Marshal(newData)
	if err != nil {
		log.Println("Error in encoding data in UpdateDataFlowStatisticWithTime")
		return
	}
	newNode := TTDINode{MashupDetailedElement: &mashupsdk.MashupDetailedElement{State: &mashupsdk.MashupElementState{State: int64(mashupsdk.Init)}, Data: string(newEncodedData)}, ChildNodes: []*TTDINode{}}
	dfs.ChildNodes = append(dfs.ChildNodes, &newNode)
	dfs.EfficientLog(newData, nil)
}

// Doesn't deserialize statistic data for updatedataflowstatistic
func (dfs *TTDINode) EfficientLog(statMap map[string]any, logF func(string, error)) {
	var dfsctx *DeliverStatCtx
	var decodedMap map[string]any = nil
	var err error
	logstat := false
	if statMap["decodedData"] != nil {
		decodedMap = statMap
		decodedData := statMap["decodedData"].(map[string]any)
		if logF == nil {
			if lf, ok := decodedData["LogFunc"].(func(string, error)); ok {
				logF = lf
				logstat = true
			}
			if ls, ok := decodedData["LogStat"].(bool); ok {
				logstat = ls
			}
		} else {
			logstat = true
		}
	}
	dfsctx, _, err = dfs.GetDeliverStatCtx(decodedMap)
	if err != nil {
		if logF != nil {
			logF("Error in decoding data in Log", err)
		}
		return
	}
	if logF != nil {
		dfsctx.LogFunc = &logF
		dfsctx.LogStat = logstat
	}
	if dfsctx.LogStat {
		if strings.Contains(dfsctx.StateName, "Failure") && dfsctx.LogFunc != nil {
			(*dfsctx.LogFunc)(dfsctx.FlowName+"-"+dfsctx.StateName, errors.New(dfsctx.StateName))
		} else if dfsctx.LogFunc != nil {
			(*dfsctx.LogFunc)(dfsctx.FlowName+"-"+dfsctx.StateName, nil)
		}
	}
}

func (dfs *TTDINode) Log() {
	dfsctx, _, err := dfs.GetDeliverStatCtx()
	if err != nil {
		log.Println("Error in decoding data in Log")
		return
	}
	if dfsctx.LogStat {
		stat := dfs.ChildNodes[len(dfs.ChildNodes)-1]
		dfstatctx, _, err := stat.GetDeliverStatCtx()
		if err != nil {
			log.Println("Error in decoding data in Log")
			return
		}
		if strings.Contains(dfstatctx.StateName, "Failure") && dfsctx.LogFunc != nil {
			(*dfsctx.LogFunc)(dfstatctx.FlowName+"-"+dfstatctx.StateName, errors.New(dfstatctx.StateName))
		} else if dfsctx.LogFunc != nil {
			(*dfsctx.LogFunc)(dfstatctx.FlowName+"-"+dfstatctx.StateName, nil)
		}
	}
}

func (dfs *TTDINode) GetDeliverStatCtx(decodedMap ...map[string]any) (*DeliverStatCtx, map[string]any, error) {
	var decodedData map[string]any
	var dsc DeliverStatCtx

	if len(decodedMap) > 0 {
		decodedData = decodedMap[0]
	} else {
		var decoded any
		err := json.Unmarshal([]byte(dfs.MashupDetailedElement.Data), &decoded)
		if err != nil {
			log.Println("Error in decoding data in FinishStatistic")
			return nil, nil, err
		}
		decodedData = decoded.(map[string]any)
	}

	if start, ok := decodedData["TimeStart"].(time.Time); ok {
		dsc.TimeStart = start.Format(time.RFC3339)
	} else {
		dsc.TimeStart = decodedData["TimeStart"]
	}
	if logStat, ok := decodedData["LogStat"].(bool); ok {
		dsc.LogStat = logStat
	}
	if decodedData["LogFunc"] != nil {
		if logFunc, ok := decodedData["LogFunc"].(func(string, error)); ok {
			dsc.LogFunc = &logFunc
		}
	}
	if decodedData["FlowGroup"] != nil {
		if flowGroup, ok := decodedData["FlowGroup"].(string); ok {
			dsc.FlowGroup = flowGroup
		}
	}
	if decodedData["FlowName"] != nil {
		if flowName, ok := decodedData["FlowName"].(string); ok {
			dsc.FlowName = flowName
		}
	}
	if decodedData["StateCode"] != nil {
		if stateCode, ok := decodedData["StateCode"].(string); ok {
			dsc.StateCode = stateCode
		}
	}
	if decodedData["StateName"] != nil {
		if stateName, ok := decodedData["StateName"].(string); ok {
			dsc.StateName = stateName
		}
	}
	if decodedData["Mode"] != nil {
		dsc.Mode = decodedData["Mode"]
	}
	if decodedData["TimeSplit"] != nil {
		dsc.TimeSplit = decodedData["TimeSplit"]
	}
	if decodedData["LastTestedDate"] != nil {
		dsc.LastTestedDate = decodedData["LastTestedDate"]
	}
	return &dsc, decodedData, nil
}

func (dfs *TTDINode) FinishStatistic(id string, indexPath string, idName string, logger *log.Logger, vaultWriteBack bool, dsc *DeliverStatCtx) map[string]any {
	dfsctx, _, err := dfs.GetDeliverStatCtx()
	if err != nil {
		log.Println("Error in decoding data in FinishStatistic")
		return nil
	}
	statMap := make(map[string]any)
	//Change names here
	statMap["flowGroup"] = dfsctx.FlowGroup
	statMap["flowName"] = dfsctx.FlowName
	statMap["stateName"] = dfsctx.StateName
	statMap["stateCode"] = dfsctx.StateCode
	statMap["timeSplit"] = dfsctx.GetElapsedTimeStr()
	statMap["mode"] = dfsctx.GetModeInt()
	statMap["lastTestedDate"] = dfsctx.GetLastTestedDateStr()
	return statMap
}

func (dfs *TTDINode) MapStatistic(data map[string]any, logger *log.Logger) {
	newData := make(map[string]any)
	newData["FlowGroup"] = data["flowGroup"].(string)
	newData["FlowName"] = data["flowName"].(string)
	newData["StateName"] = data["stateName"].(string)
	newData["StateCode"] = data["stateCode"].(string)
	newData["LastTestedDate"] = data["lastTestedDate"].(string)
	if mode, ok := data["mode"]; ok {
		modeStr := fmt.Sprintf("%s", mode) //Treats it as a interface due to weird typing from vault (encoding/json.Number)
		if modeInt, err := strconv.Atoi(modeStr); err == nil {
			newData["Mode"] = modeInt
		}
	}
	if strings.Contains(data["timeSplit"].(string), "seconds") {
		data["timeSplit"] = strings.ReplaceAll(data["timeSplit"].(string), " seconds", "s")
	}
	newData["TimeSplit"], _ = time.ParseDuration(data["timeSplit"].(string))

	newEncodedData, err := json.Marshal(newData)
	if err != nil {
		log.Println("Error encoding data in RetrieveStatistic")
		return
	}
	dfs.MashupDetailedElement.Data = string(newEncodedData)
}

// Set logFunc and logStat = false to use this otherwise it logs as states change with logStat = true
func (dfs *TTDINode) FinishStatisticLog() {
	dfsctx, _, err := dfs.GetDeliverStatCtx()
	if err != nil {
		log.Printf("Error in decoding data in FinishStatisticLog: %s\n", err)
		return
	}
	if dfsctx.LogFunc == nil || dfsctx.LogStat {
		return
	}
	for _, stat := range dfs.ChildNodes {
		dfstatCtx, _, err := stat.GetDeliverStatCtx()
		if err != nil {
			log.Println("Error in decoding data in FinishStatisticLog")
			return
		}
		if strings.Contains(dfstatCtx.StateName, "Failure") && dfsctx.LogFunc != nil {
			(*dfsctx.LogFunc)(dfstatCtx.FlowName+"-"+dfstatCtx.StateName, errors.New(dfstatCtx.StateName))
			if dfstatCtx.Mode == 2 { //Update snapshot Mode on failure so it doesn't repeat

			}
		} else {
			(*dfsctx.LogFunc)(dfstatCtx.FlowName+"-"+dfstatCtx.StateName, nil)
		}
	}
}

// Creating map representation for easier use by persistence functions
func (dfs *TTDINode) StatisticToMap() map[string]any {
	var elapsedTime string
	statMap := make(map[string]any)
	dfsctx, _, err := dfs.GetDeliverStatCtx()
	if err != nil {
		log.Println("Error in decoding data in StatisticToMap")
		return statMap
	}

	statMap["flowGroup"] = dfsctx.FlowGroup
	statMap["flowName"] = dfsctx.FlowName
	statMap["stateName"] = dfsctx.StateName
	statMap["stateCode"] = dfsctx.StateCode
	if _, ok := dfsctx.TimeSplit.(time.Duration); ok {
		if dfsctx.TimeSplit != nil && dfsctx.TimeSplit.(time.Duration).Seconds() < 0 { //Covering corner case of 0 second time durations being slightly off (-.00004 seconds)
			elapsedTime = "0s"
		} else {
			elapsedTime = dfsctx.TimeSplit.(time.Duration).Truncate(time.Millisecond * 10).String()
		}
	} else if timeFloat, ok := dfsctx.TimeSplit.(float64); ok {
		elapsedTime = time.Duration(timeFloat * float64(time.Nanosecond)).Truncate(time.Millisecond * 10).String()
	}
	statMap["timeSplit"] = elapsedTime
	if modeFloat, ok := dfsctx.Mode.(float64); ok {
		statMap["mode"] = int(modeFloat)
	} else {
		if dfsctx.Mode != nil {
			statMap["mode"] = dfsctx.Mode
		} else {
			statMap["mode"] = 0
		}
	}
	if ltd, ok := dfsctx.LastTestedDate.(string); ok {
		statMap["lastTestedDate"] = ltd
	} else {
		statMap["lastTestedDate"] = ""
	}
	return statMap
}
