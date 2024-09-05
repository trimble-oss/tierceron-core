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

// type DataFlowStatistic struct {
// 	mashupsdk.MashupDetailedElement
// 	FlowGroup string
// 	FlowName  string
// 	StateName string
// 	StateCode string
// 	TimeSplit time.Duration
// 	Mode      int
// }

// type DataFlow struct {
// 	mashupsdk.MashupDetailedElement
// 	Name       string
// 	TimeStart  time.Time
// 	Statistics []DataFlowStatistic
// 	LogStat    bool
// 	LogFunc    func(string, error)
// }

// type DataFlowGroup struct {
// 	mashupsdk.MashupDetailedElement
// 	Name  string
// 	Flows []DataFlow
// }

// type Argosy struct {
// 	mashupsdk.MashupDetailedElement
// 	ArgosyID string
// 	Groups   []DataFlowGroup
// }

// type ArgosyFleet struct {
// 	ArgosyName string
// 	Argosies   []Argosy
// }

type TTDINode struct {
	*mashupsdk.MashupDetailedElement
	//Data       []byte
	ChildNodes []*TTDINode
}

type DeliverStatCtx struct {
	LogStat   *bool
	LogFunc   *func(string, error)
	TimeStart *string
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
	var decoded interface{}
	var decodedData map[string]interface{}
	var timeStart time.Time
	if len(dfs.MashupDetailedElement.Data) > 0 {
		err := json.Unmarshal([]byte(dfs.MashupDetailedElement.Data), &decoded)
		if err != nil {
			logF("Error in decoding data in UpdateDataFlowStatistic", err)
			return
		}
		decodedData = decoded.(map[string]interface{})

		//string to time.time
		if decodedData["TimeStart"] != nil {
			if _, ok := decoded.(time.Time); ok {
				timeStart = decodedData["TimeStart"].(time.Time)
			} else {
				var timeParseErr error
				timeStartStr := decodedData["TimeStart"].(string)
				timeStart, timeParseErr = time.Parse(RFC_ISO_8601, timeStartStr)
				if timeParseErr != nil {
					logF("Error in parsing start time in UpdateDataFlowStatistics", timeParseErr)
					return
				}
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

func (dfs *TTDINode) GetDeliverStatCtx() (*DeliverStatCtx, error) {
	var decoded interface{}
	var dsc DeliverStatCtx
	err := json.Unmarshal([]byte(dfs.MashupDetailedElement.Data), &decoded)
	if err != nil {
		log.Println("Error in decoding data in FinishStatistic")
		return nil, err
	}
	decodedData := decoded.(map[string]interface{})
	if start, ok := decodedData["TimeStart"].(time.Time); ok {
		time := start.Format(time.RFC3339)
		dsc.TimeStart = &time
	}
	if logStat, ok := decodedData["LogStat"].(bool); ok {
		dsc.LogStat = &logStat
	}
	if decodedData["LogFunc"] != nil {
		if logFunc, ok := decodedData["LogFunc"].(func(string, error)); ok {
			dsc.LogFunc = &logFunc
		}
	}
	return &dsc, nil
}

func (dfs *TTDINode) FinishStatistic(id string, indexPath string, idName string, logger *log.Logger, vaultWriteBack bool, dsc *DeliverStatCtx) *map[string]interface{} {
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
	if dsc.TimeStart != nil {
		lastTestedDate = *dsc.TimeStart
	} else if _, ok := decodedStatData["TimeStart"].(string); ok {
		lastTestedDate = decodedStatData["TimeStart"].(string)
	}

	statMap["lastTestedDate"] = lastTestedDate
	return &statMap
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
	var decoded interface{}
	err := json.Unmarshal([]byte(dfs.MashupDetailedElement.Data), &decoded)
	if err != nil {
		log.Println("Error in decoding data in FinishStatisticLog")
		return
	}
	decodedData := decoded.(map[string]interface{})
	if decodedData["LogStat"] == nil || decodedData["LogFunc"] == nil {
		return
	}
	if decodedData["LogFunc"] == nil || (decodedData["LogStat"] != nil && decodedData["LogStat"].(bool)) {
		return
	}
	for _, stat := range dfs.ChildNodes {
		var decodedstat interface{}
		err := json.Unmarshal([]byte(stat.MashupDetailedElement.Data), &decodedstat)
		if err != nil {
			log.Println("Error in decoding data in FinishStatisticLog")
			return
		}
		decodedStatData := decodedstat.(map[string]interface{})
		if decodedStatData["StateName"] != nil && strings.Contains(decodedStatData["StateName"].(string), "Failure") && decodedData["LogFunc"] != nil {
			logFunc := decodedData["LogFunc"].(func(string, error))
			logFunc(decodedStatData["FlowName"].(string)+"-"+decodedStatData["StateName"].(string), errors.New(decodedStatData["StateName"].(string)))
			//dfs.LogFunc(stat.FlowName+"-"+stat.StateName, errors.New(stat.StateName))
			if decodedStatData["Mode"] != nil {
				if modeFloat, ok := decodedStatData["Mode"].(float64); ok {
					if modeFloat == 2 { //Update snapshot Mode on failure so it doesn't repeat

					}
				} else {
					if decodedStatData["Mode"] == 2 { //Update snapshot Mode on failure so it doesn't repeat

					}
				}
			}
		} else {
			logFunc := decodedData["LogFunc"].(func(string, error))
			logFunc(decodedStatData["FlowName"].(string)+"-"+decodedStatData["StateName"].(string), nil)

			//dfs.LogFunc(stat.FlowName+"-"+stat.StateName, nil)
		}
	}
}
