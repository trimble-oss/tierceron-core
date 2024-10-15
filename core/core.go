package core

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

const (
	PLUGIN_EVENT_START = iota
	PLUGIN_EVENT_STOP
	PLUGIN_EVENT_STATUS
)

const PLUGIN_EVENT_CHANNELS_MAP_KEY = "PluginEventChannelsMap"
const PLUGIN_CHANNEL_EVENT_IN = "PluginChannelEventIn"
const PLUGIN_CHANNEL_EVENT_OUT = "PluginChannelEventOut"
const DATA_FLOW_STAT_CHANNEL = "DataFlowStatisticsChannel"
const ERROR_CHANNEL = "ErrorChannel"

const RFC_ISO_8601 = "2006-01-02 15:04:05 -0700 MST"

type ConfigContext struct {
	Config          *map[string]interface{}
	Env             string // Env being processed
	Start           func()
	ErrorSenderChan *chan error     // Channel for sending errors
	DfsChan         *chan *TTDINode // Channel for sending data flow statistics
	ArgosId         string          // Identifier for data flow statistics
	ConfigCerts     *map[string][]byte
	Log             *log.Logger
}

func Init(properties *map[string]interface{},
	commonCertPath string,
	commonKeyPath string,
	commonPath string,
	dfsKeyHeader string,
	startHandler func(),
	receiverHandler func(chan int),
) (*ConfigContext, error) {
	if properties == nil {
		fmt.Println("Missing initialization components")
		return nil, errors.New("missing initialization components")
	}
	var logger *log.Logger
	if _, ok := (*properties)["log"].(*log.Logger); ok {
		logger = (*properties)["log"].(*log.Logger)
	}

	var env string
	var argosId string
	if e, ok := (*properties)["env"].(string); ok {
		env = e
	} else {
		fmt.Println("Missing env from kernel")
		logger.Println("Missing env from kernel")
		return nil, errors.New("missing env from kernel")
	}

	if len(argosId) == 0 {
		splitEnv := strings.Split(env, "-")
		if len(splitEnv) == 2 {
			argosId = fmt.Sprintf("%s-%s", dfsKeyHeader, splitEnv[1])
		} else {
			argosId = dfsKeyHeader
		}
	}
	logger.Printf("Starting initialization for dataflow: %s\n", argosId)

	var certbytes []byte
	var keybytes []byte
	var config_properties *map[string]interface{}
	if cert, ok := (*properties)[commonCertPath]; ok {
		certbytes = cert.([]byte)
	}
	if key, ok := (*properties)[commonKeyPath]; ok {
		keybytes = key.([]byte)
	}
	if common, ok := (*properties)[commonPath]; ok {
		config_properties = common.(*map[string]interface{})
	} else {
		fmt.Println("Missing config components")
		return nil, errors.New("missing config components")
	}

	var configCerts *map[string][]byte = &map[string][]byte{}

	if len(certbytes) > 0 && len(keybytes) > 0 {
		(*configCerts)[commonCertPath] = certbytes
		(*configCerts)[commonKeyPath] = keybytes
	}

	configContext := &ConfigContext{
		Env:         env,
		Config:      config_properties,
		Start:       startHandler,
		ArgosId:     argosId,
		ConfigCerts: configCerts,
		Log:         logger,
	}

	if channels, ok := (*properties)[PLUGIN_EVENT_CHANNELS_MAP_KEY]; ok {
		if chans, ok := channels.(map[string]interface{}); ok {
			if rchan, ok := chans[PLUGIN_CHANNEL_EVENT_IN]; ok {
				if rc, ok := rchan.(chan int); ok && rc != nil {
					configContext.Log.Println("Receiver initialized.")
					go receiverHandler(rc)
				} else {
					configContext.Log.Println("Unsupported receiving channel passed")
					return nil, errors.New("unsupported receiving channel passed")
				}
			} else {
				configContext.Log.Println("No receiving channel passed")
				return nil, errors.New("no receiving channel passed")
			}
			if schan, ok := chans[PLUGIN_CHANNEL_EVENT_OUT].(map[string]interface{}); ok {
				if sc, ok := schan[ERROR_CHANNEL].(chan error); ok && sc != nil {
					configContext.Log.Println("Error Sender initialized")
					configContext.ErrorSenderChan = &sc
				} else {
					configContext.Log.Println("Unsupported error sending channel passed")
					return nil, errors.New("unsupported error sending channel passed")
				}
				if ttdichan, ok := schan[DATA_FLOW_STAT_CHANNEL].(chan *TTDINode); ok {
					configContext.Log.Println("Data flow statistics channel initialized.")
					configContext.DfsChan = &ttdichan
				} else {
					configContext.Log.Println("Unsupported dataflow statistics sending channel passedUnsupported dataflow statistics sending channel passed")
					return nil, errors.New("unsupported dataflow statistics sending channel passed")
				}
			} else {
				configContext.Log.Println("No sending channels passed")
				return nil, errors.New("no sending channels passed")
			}
		} else {
			configContext.Log.Println("No channels passed")
			return nil, errors.New("no channels passed")
		}
	}
	configContext.Log.Println("Successfully initialized rainier")
	return configContext, nil
}
