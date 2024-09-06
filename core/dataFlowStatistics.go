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

	"github.com/trimble-oss/tierceron-nute/mashupsdk"
)

type TTDINode struct {
	*mashupsdk.MashupDetailedElement
	//Data       []byte
	ChildNodes []*TTDINode
}

type DeliverStatCtx struct {
	FlowGroup      string
	FlowName       string
	StateCode      string
	StateName      string
	LastTestedDate interface{} //TODO: can this be string?
	LogStat        bool
	LogFunc        *func(string, error)
	TimeStart      interface{}
	TimeSplit      interface{} //either float64 or duration
	Mode           interface{} //either float64 or int
}

func InitDataFlow(logF func(string, error), name string, logS bool) *TTDINode {
	var stats []*TTDINode
	data := make(map[string]interface{})
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
	//var newDFStatistic = DataFlow{Name: name, TimeStart: time.Now(), Statistics: stats, LogStat: logS, LogFunc: logF}
	return ttdiNode
}

func (dfs *TTDINode) UpdateDataFlowStatistic(flowG string, flowN string, stateN string, stateC string, mode int, logF func(string, error)) {
	var decodedData map[string]interface{}
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
		decodedData = make(map[string]interface{})
		timeStart = time.Now()
		decodedData["TimeStart"] = timeStart.Format(RFC_ISO_8601)

		newEncodedData, err := json.Marshal(decodedData)
		if err != nil {
			logF("Error in encoding data in UpdateDataFlowStatistics", err)
			return
		}
		dfs.MashupDetailedElement.Data = string(newEncodedData)
	}

	newData := make(map[string]interface{})
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
	//var newDFStat = DataFlowStatistic{mashupsdk.MashupDetailedElement{}, flowG, flowN, stateN, stateC, time.Since(dfs.TimeStart), mode}
	dfs.ChildNodes = append(dfs.ChildNodes, &newNode)
	newData["decodedData"] = decodedData
	dfs.EfficientLog(newData, logF)
}

func (dfs *TTDINode) UpdateDataFlowStatisticWithTime(flowG string, flowN string, stateN string, stateC string, mode int, elapsedTime time.Duration) {
	newData := make(map[string]interface{})
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
	//var newDFStat = DataFlowStatistic{mashupsdk.MashupDetailedElement{}, flowG, flowN, stateN, stateC, elapsedTime, mode}
	dfs.ChildNodes = append(dfs.ChildNodes, &newNode)
	dfs.EfficientLog(newData, nil)
}

// Doesn't deserialize statistic data for updatedataflowstatistic
func (dfs *TTDINode) EfficientLog(statMap map[string]interface{}, logF func(string, error)) {
	var decodedData map[string]interface{}
	if statMap["decodedData"] == nil {
		var decoded interface{}
		err := json.Unmarshal([]byte(dfs.MashupDetailedElement.Data), &decoded)
		if err != nil {
			if logF != nil {
				logF("Error in decoding data in Log", err)
			}
			return
		}
		decodedData = decoded.(map[string]interface{})
	} else if logF != nil {
		decodedData = map[string]interface{}{
			"LogFunc": logF,
			"LogStat": true,
		}
	} else {
		decodedData = statMap["decodedData"].(map[string]interface{})
	}

	if decodedData["LogStat"] != nil && decodedData["LogStat"].(bool) {
		if statMap["StateName"] != nil && strings.Contains(statMap["StateName"].(string), "Failure") && decodedData["LogFunc"] != nil {
			logFunc := decodedData["LogFunc"].(func(string, error))
			logFunc(statMap["FlowName"].(string)+"-"+statMap["StateName"].(string), errors.New(statMap["StateName"].(string)))
			//dfs.LogFunc(stat.FlowName+"-"+stat.StateName, errors.New(stat.StateName))
		} else if decodedData["LogFunc"] != nil {
			logFunc := decodedData["LogFunc"].(func(string, error))
			logFunc(statMap["FlowName"].(string)+"-"+statMap["StateName"].(string), nil)
			//dfs.LogFunc(stat.FlowName+"-"+stat.StateName, nil)
		}
	}
}

func (dfs *TTDINode) Log() {
	var decoded interface{}
	err := json.Unmarshal([]byte(dfs.MashupDetailedElement.Data), &decoded)
	if err != nil {
		log.Println("Error in decoding data in Log")
		return
	}
	decodedData := decoded.(map[string]interface{})
	if decodedData["LogStat"] != nil && decodedData["LogStat"].(bool) {
		stat := dfs.ChildNodes[len(dfs.ChildNodes)-1]
		var decodedstat interface{}
		err := json.Unmarshal([]byte(stat.MashupDetailedElement.Data), &decodedstat)
		if err != nil {
			log.Println("Error in decoding data in Log")
			return
		}
		decodedStatData := decodedstat.(map[string]interface{})
		if decodedStatData["StateName"] != nil && strings.Contains(decodedStatData["StateName"].(string), "Failure") && decodedData["LogFunc"] != nil {
			logFunc := decodedData["LogFunc"].(func(string, error))
			logFunc(decodedStatData["FlowName"].(string)+"-"+decodedStatData["StateName"].(string), errors.New(decodedStatData["StateName"].(string)))
			//dfs.LogFunc(stat.FlowName+"-"+stat.StateName, errors.New(stat.StateName))
		} else if decodedData["LogFunc"] != nil {
			logFunc := decodedData["LogFunc"].(func(string, error))
			logFunc(decodedStatData["FlowName"].(string)+"-"+decodedStatData["StateName"].(string), nil)
			//dfs.LogFunc(stat.FlowName+"-"+stat.StateName, nil)
		}
	}
}

func (dfs *TTDINode) GetDeliverStatCtx() (*DeliverStatCtx, map[string]interface{}, error) {
	var decoded interface{}
	var dsc DeliverStatCtx
	err := json.Unmarshal([]byte(dfs.MashupDetailedElement.Data), &decoded)
	if err != nil {
		log.Println("Error in decoding data in FinishStatistic")
		return nil, nil, err
	}
	decodedData := decoded.(map[string]interface{})
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

func (dfs *TTDINode) FinishStatistic(id string, indexPath string, idName string, logger *log.Logger, vaultWriteBack bool, dsc *DeliverStatCtx) map[string]interface{} {
	var decodedstat interface{}
	err := json.Unmarshal([]byte(dfs.MashupDetailedElement.Data), &decodedstat)
	if err != nil {
		log.Println("Error in decoding data in FinishStatistic")
		return nil
	}
	decodedStatData := decodedstat.(map[string]interface{})
	var elapsedTime string
	statMap := make(map[string]interface{})
	//Change names here
	statMap["flowGroup"] = decodedStatData["FlowGroup"]
	statMap["flowName"] = decodedStatData["FlowName"]
	statMap["stateName"] = decodedStatData["StateName"]
	statMap["stateCode"] = decodedStatData["StateCode"]
	if _, ok := decodedStatData["TimeSplit"].(time.Duration); ok {
		if decodedStatData["TimeSplit"] != nil && decodedStatData["TimeSplit"].(time.Duration).Seconds() < 0 { //Covering corner case of 0 second time durations being slightly off (-.00004 seconds)
			elapsedTime = "0s"
		} else {
			elapsedTime = decodedStatData["TimeSplit"].(time.Duration).Truncate(time.Millisecond * 10).String()
		}
	} else if timeFloat, ok := decodedStatData["TimeSplit"].(float64); ok {
		elapsedTime = time.Duration(timeFloat * float64(time.Nanosecond)).Truncate(time.Millisecond * 10).String()
	}
	statMap["timeSplit"] = elapsedTime
	if modeFloat, ok := decodedStatData["Mode"].(float64); ok {
		statMap["mode"] = int(modeFloat)
	} else {
		statMap["mode"] = decodedStatData["Mode"]
	}
	lastTestedDate := ""
	if t, ok := dsc.TimeStart.(string); ok {
		lastTestedDate = t
	} else if _, ok := decodedStatData["TimeStart"].(string); ok {
		lastTestedDate = decodedStatData["TimeStart"].(string)
	}

	statMap["lastTestedDate"] = lastTestedDate
	return statMap
}

func (dfs *TTDINode) MapStatistic(data map[string]interface{}, logger *log.Logger) {
	newData := make(map[string]interface{})
	newData["FlowGroup"] = data["flowGroup"].(string)
	newData["FlowName"] = data["flowName"].(string)
	newData["StateName"] = data["stateName"].(string)
	newData["StateCode"] = data["stateCode"].(string)
	newData["LastTestedDate"] = data["lastTestedDate"].(string)
	if mode, ok := data["mode"]; ok {
		modeStr := fmt.Sprintf("%s", mode) //Treats it as a interface due to weird typing from vault (encoding/json.Number)
		if modeInt, err := strconv.Atoi(modeStr); err == nil {
			//df.Mode = modeInt
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
func (dfs *TTDINode) StatisticToMap() map[string]interface{} {
	var elapsedTime string
	statMap := make(map[string]interface{})
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
		statMap["mode"] = dfsctx.Mode
	}
	if ltd, ok := dfsctx.LastTestedDate.(string); ok {
		statMap["lastTestedDate"] = ltd
	} else {
		statMap["lastTestedDate"] = ""
	}
	return statMap
}
