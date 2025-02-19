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
const CHAT_BROADCAST_CHANNEL = "ChatBroadcastChannel"
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
	Config            *map[string]interface{}
	Env               string // Env being processed
	Region            string // Region processed
	Start             func(string)
	ChatSenderChan    *chan *ChatMsg
	ChatReceiverChan  *chan *ChatMsg
	ChatBroadcastChan *chan *ChatMsg
	CmdSenderChan     *chan KernelCmd
	CmdReceiverChan   *chan KernelCmd
	ErrorChan         *chan error     // Channel for sending errors
	DfsChan           *chan *TTDINode // Channel for sending data flow statistics
	ArgosId           string          // Identifier for data flow statistics
	ConfigCerts       *map[string][]byte
	Log               *log.Logger
}

//
// Plugin initialization:
// 1. Kernel calls GetConfigPaths
// 2. Kernel calls Init
//    - certs and configs passed to plugin.
// 3. Kernel makes channel messaging to continue boot process passed to receiverHandler with events:
//    a. PLUGIN_EVENT_START
//    b. PLUGIN_EVENT_STOP -- on shutdown..
// 4. ChatMsg events sent to chatReceiverHandler via chat_receive_chan
//    a. Responses put into *configContext.ChatSenderChan
//       All messages sent by plugins must dump pointers to ChatMsg into
//      *configContext.ChatSenderChan
//       example: *configContext.ChatSenderChan <- &chatResultMsg
//
type ChatMsg struct {
	ChatId      *string   // Only relevant for 3rd party integration.
	Name        *string   // Source plugin name
	KernelId    *string   // Internal use by kernel
	IsBroadcast bool      // Is message intended for broadcast.
	Query       *[]string // List of plugins to send message to.
	Response    *string   // Pointer to json serialized data.
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
	region := ""
	if configProp, ok := (*properties)["config"].(*map[string]interface{}); ok {
		if r, ok := (*configProp)["region"].(string); ok {
			logger.Println("received region value from kernel")
			region = r
		}
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
		Region:      region,
		Config:      config_properties,
		Start:       startHandler,
		ArgosId:     argosId,
		ConfigCerts: configCerts,
		Log:         logger,
	}

	if channels, ok := (*properties)[PLUGIN_EVENT_CHANNELS_MAP_KEY]; ok {
		if chans, ok := channels.(map[string]interface{}); ok {
			if bchan, ok := chans[CHAT_BROADCAST_CHANNEL].(*chan *ChatMsg); ok {
				configContext.Log.Println("Chat broadcast channel initialized.")
				configContext.ChatBroadcastChan = bchan
			} else {
				configContext.Log.Println("Unsupported broadcast channel passed")
				return nil, errors.New("unsupported broadcast channel passed")
			}
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
