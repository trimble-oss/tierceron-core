package core

import (
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
)

var activeLogPlugins sync.Map

type pluginLogWriter struct {
	pluginName string
	writer     io.Writer
}

func (writer *pluginLogWriter) Write(message []byte) (int, error) {
	if active, ok := activeLogPlugins.Load(writer.pluginName); ok && !active.(bool) {
		return len(message), nil
	}
	return writer.writer.Write(message)
}

func ActivatePluginLogging(pluginName string) {
	if pluginName != "" {
		activeLogPlugins.LoadOrStore(pluginName, true)
	}
}

func SetPluginLoggingActive(pluginName string, active bool) {
	if pluginName != "" {
		activeLogPlugins.Store(pluginName, active)
	}
}

func IsPluginLoggingActive(pluginName string) bool {
	active, ok := activeLogPlugins.Load(pluginName)
	return !ok || active.(bool)
}

func NewPluginLogger(pluginName string, logger *log.Logger) *log.Logger {
	ActivatePluginLogging(pluginName)
	if logger == nil || pluginName == "" {
		return logger
	}
	prefix := logger.Prefix()
	if !strings.Contains(prefix, "["+pluginName+"]") {
		prefix += "[" + pluginName + "] "
	}
	return log.New(&pluginLogWriter{pluginName: pluginName, writer: logger.Writer()}, prefix, logger.Flags())
}

func configurePluginChatReceiver(configContext *ConfigContext, kernelChatReceiver *chan *ChatMsg) *chan *ChatMsg {
	pluginChatReceiver := make(chan *ChatMsg)
	configContext.ChatReceiverChan = &pluginChatReceiver
	go func() {
		for event := range *kernelChatReceiver {
			if !handlePluginLoggingDirective(configContext, event) {
				pluginChatReceiver <- event
			}
		}
	}()
	return &pluginChatReceiver
}

func handlePluginLoggingDirective(configContext *ConfigContext, event *ChatMsg) bool {
	if configContext == nil || configContext.PluginName == "" || event == nil || event.Response != nil || event.Name == nil || *event.Name != configContext.PluginName || event.Query == nil || len(*event.Query) == 0 || (*event.Query)[0] != "trcshtalk" || event.ChatId == nil {
		return false
	}

	fields := strings.Fields(strings.ToLower(*event.ChatId))
	if len(fields) != 2 || fields[0] != "log" || (fields[1] != "start" && fields[1] != "stop") {
		return false
	}

	active := fields[1] == "start"
	SetPluginLoggingActive(configContext.PluginName, active)
	if configContext.ChatSenderChan == nil || *configContext.ChatSenderChan == nil {
		return true
	}

	pluginName := configContext.PluginName
	state := "disabled"
	if active {
		state = "enabled"
	}
	response := fmt.Sprintf("logging %s for %s", state, pluginName)
	query := []string{(*event.Query)[0]}
	*configContext.ChatSenderChan <- &ChatMsg{
		RoutingId: event.RoutingId,
		Name:      &pluginName,
		Query:     &query,
		Response:  &response,
	}
	return true
}
