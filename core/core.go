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
const CMD_CHANNEL = "CommandChannel"
const CHAT_CHANNEL = "ChatChannel"
const DATA_FLOW_STAT_CHANNEL = "DataFlowStatisticsChannel"
const ERROR_CHANNEL = "ErrorChannel"

const RFC_ISO_8601 = "2006-01-02 15:04:05 -0700 MST"

const (
	TRCSHHIVEK_CERT = "Common/servicecert.crt.mf.tmpl"
	TRCSHHIVEK_KEY  = "Common/servicekey.key.mf.tmpl"
)

type KernelCmd struct {
	PluginName string
	Command    int
}

type ConfigContext struct {
	Config           *map[string]interface{}
	Env              string // Env being processed
	Start            func(string)
	ChatSenderChan   *chan *ChatMsg
	ChatReceiverChan *chan *ChatMsg
	CmdSenderChan    *chan KernelCmd
	CmdReceiverChan  *chan KernelCmd
	ErrorChan        *chan error     // Channel for sending errors
	DfsChan          *chan *TTDINode // Channel for sending data flow statistics
	ArgosId          string          // Identifier for data flow statistics
	ConfigCerts      *map[string][]byte
	Log              *log.Logger
}

type ChatMsg struct {
	ChatId   *string
	Name     *string //plugin name
	KernelId *string
	Query    *[]string
	Response *string
}

func Init(properties *map[string]interface{},
	commonCertPath string,
	commonKeyPath string,
	commonPath string,
	dfsKeyHeader string,
	startHandler func(string),
	receiverHandler func(chan KernelCmd),
	chatHandler func(chan *ChatMsg),
) (*ConfigContext, error) {
	if properties == nil ||
		startHandler == nil ||
		receiverHandler == nil ||
		chatHandler == nil {
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
			if rchan, ok := chans[PLUGIN_CHANNEL_EVENT_IN].(map[string]interface{}); ok {
				if rc, ok := rchan[CMD_CHANNEL].(*chan KernelCmd); ok && rc != nil {
					configContext.Log.Println("Command Receiver initialized.")
					configContext.CmdReceiverChan = rc
					go receiverHandler(*rc)
				} else {
					configContext.Log.Println("Unsupported command receiving channel passed")
					return nil, errors.New("unsupported command receiving channel passed")
				}
				if cr, ok := rchan[CHAT_CHANNEL].(*chan *ChatMsg); ok && cr != nil {
					configContext.Log.Println("Chat Receiver initialized.")
					configContext.ChatReceiverChan = cr
					go chatHandler(*cr)
				} else {
					configContext.Log.Println("Unsupported chat message receiving channel passed")
					return nil, errors.New("unsupported chat message receiving channel passed")
				}
			} else {
				configContext.Log.Println("No receiving channel passed")
				return nil, errors.New("no receiving channel passed")
			}
			if schan, ok := chans[PLUGIN_CHANNEL_EVENT_OUT].(map[string]interface{}); ok {
				if sc, ok := schan[ERROR_CHANNEL].(*chan error); ok && sc != nil {
					configContext.Log.Println("Error Sender initialized")
					configContext.ErrorChan = sc
				} else {
					configContext.Log.Println("Unsupported error sending channel passed")
					return nil, errors.New("unsupported error sending channel passed")
				}
				if ttdichan, ok := schan[DATA_FLOW_STAT_CHANNEL].(*chan *TTDINode); ok {
					configContext.Log.Println("Data flow statistics channel initialized.")
					configContext.DfsChan = ttdichan
				} else {
					configContext.Log.Println("Unsupported dataflow statistics sending channel passedUnsupported dataflow statistics sending channel passed")
					return nil, errors.New("unsupported dataflow statistics sending channel passed")
				}
				if cmdsender, ok := schan[CMD_CHANNEL].(*chan KernelCmd); ok {
					configContext.Log.Println("Command status sending channel initialized.")
					configContext.CmdSenderChan = cmdsender
				} else {
					configContext.Log.Println("Unsupported command status sending channel passed")
					return nil, errors.New("unsupported command status sending channel passed")
				}
				if chsender, ok := schan[CHAT_CHANNEL].(*chan *ChatMsg); ok {
					configContext.Log.Println("Chat message sending channel initialized.")
					configContext.ChatSenderChan = chsender
				} else {
					configContext.Log.Println("Unsupported chat message sending channel passed")
					return nil, errors.New("unsupported chat message sending channel passed")
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
	configContext.Log.Println("Successfully initialized plugin")
	return configContext, nil
}

func SanitizeForLogging(errMsg string) string {
	errMsgSanitized := strings.ReplaceAll(errMsg, "\n", "")
	errMsgSanitized = strings.ReplaceAll(errMsgSanitized, "\r", "")
	return errMsgSanitized
}
